package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"main/proxyctl"
	"main/settings"
	"main/subscription"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strconv"
	"syscall"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"

	log "github.com/sirupsen/logrus"
)

var setting *settings.Setting
var err error

var (
	flags = struct {
		verbose bool   // 显示详细日志
		sub     bool   // 重新获取订阅
		ping    bool   // 重新测试节点延迟
		filter  string // 过滤测试的节点
		url     string // 追加订阅地址

		socks uint64 // socks5 代理端口
		http  uint64 // http 代理端口
	}{}
	commands = [][]string{
		{"start", "Start proxy"},             // 启动代理
		{"ls", "View configuration details"}, // 查看配置详情
		{"rmu", "Remove subscription"},       // 删除指定的订阅地址
		{"rmf", "Remove filter"},             // 删除指定的过滤器
	}
)

type TableRow struct {
	Index int
	Name  string
	Addr  string
	Ping  int32
}

func init() {
	setting, err = settings.LoadSettings()
	if err != nil {
		panic(err)
	}

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetLevel(func() log.Level {
		if setting.Verbose {
			return log.DebugLevel
		}
		return log.WarnLevel
	}())
}

func main() {
	flag.Usage = func() {
		fmt.Println(`__  ___ __ ___
\ \/ / '__/ __|
 >  <| | | (__
/_/\_\_|  \___|`)
		fmt.Println("A xray client.")
		fmt.Println()

		fmt.Println("Usage:")
		fmt.Println("  xrc [flags]")
		fmt.Println("  xrc [command]")
		fmt.Println()

		fmt.Println("Available Commands:")
		for _, cmd := range commands {
			fmt.Printf("  %-6s \t%-6s\n", cmd[0], cmd[1])
		}
		fmt.Println()

		fmt.Println("Flags:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  -%-6s \t%-6s\n", f.Name, f.Usage)
		})
		fmt.Println()

		fmt.Println("Example:")
		fmt.Println("  xrc -url https://your-subscription-url -sub -ping")
		fmt.Println("  xrc -socks 1088 -http 1099 start")
	}

	flag.BoolVar(&flags.verbose, "v", false, "Show more logs")
	flag.BoolVar(&flags.sub, "sub", false, "Retrieve subscriptions")
	flag.BoolVar(&flags.ping, "ping", false, "Test node delay")
	flag.StringVar(&flags.filter, "f", "", "Filter test nodes")
	flag.StringVar(&flags.url, "url", "", "Append subscription address")
	flag.Uint64Var(&flags.socks, "socks", 0, "Set socks proxy port")
	flag.Uint64Var(&flags.http, "http", 0, "Set http proxy port")

	flag.Parse()

	var firstArg string
	args := flag.Args()
	if len(args) > 0 {
		firstArg = args[0]

		switch firstArg {
		case "ls":
			printDetailedInfo()
			return
		case "rmu":
			tag, err := strconv.Atoi(args[1])
			if err == nil {
				setting.Urls = slices.Delete(setting.Urls, tag, tag+1)
			}
			setting.Save()
			return
		case "rmf":
			tag, err := strconv.Atoi(args[1])
			if err == nil {
				setting.Filters = slices.Delete(setting.Filters, tag, tag+1)
			}
			setting.Save()
			return
		}
	}

	if flags.verbose {
		setting.Verbose = flags.verbose
	}

	if flags.url != "" {
		if !slices.ContainsFunc(setting.Urls, func(s string) bool {
			return s == flags.url
		}) {
			setting.Urls = append(setting.Urls, flags.url)
			flags.sub = true
		}
	}

	if flags.filter != "" {
		if !slices.ContainsFunc(setting.Filters, func(s *settings.Filter) bool {
			return s.Selector == flags.filter
		}) {
			tag := uuid.New()
			setting.Filters = append(setting.Filters, &settings.Filter{
				Tag:      tag.String()[:8],
				Selector: flags.filter,
			})
		}
	}

	if flags.http > 0 {
		http := &settings.Listen{Protocol: "http", Port: uint32(flags.http)}
		if len(setting.Listens) == 0 {
			setting.Listens = append(setting.Listens, http)
		} else {
			f := false
			for i, l := range setting.Listens {
				if l.Protocol == "http" {
					setting.Listens[i].Port = uint32(flags.http)
					f = true
					break
				}
			}
			if !f {
				setting.Listens = append(setting.Listens, http)
			}
		}
	}

	if flags.socks > 0 {
		sock := &settings.Listen{Protocol: "socks", Port: uint32(flags.socks)}
		if len(setting.Listens) == 0 {
			setting.Listens = append(setting.Listens, sock)
		} else {
			f := false
			for i, l := range setting.Listens {
				if l.Protocol == "socks" {
					setting.Listens[i].Port = uint32(flags.socks)
					f = true
					break
				}
			}
			if !f {
				setting.Listens = append(setting.Listens, sock)
			}
		}
	}

	setting.Save()

	if flags.sub {
		_, err := subscription.Resubscribe(setting.Urls)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		// if len(lks) == 0 {
		// 	fmt.Println("no subscription found")
		// } else {
		// 	var nodeData [][]string
		// 	for i, l := range lks {
		// 		nodeData = append(nodeData, []string{
		// 			fmt.Sprintf("%d", i),
		// 			l.Remarks,
		// 			l.Address,
		// 		})
		// 	}
		// 	printAsTable([]string{"tag", "name", "addr"}, nodeData)
		// }
	}

	if flags.ping {
		lks, err := remeasureDelay()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		printFastestLink(lks)
	}

	if firstArg == "start" {
		var lks []*subscription.Link
		var err error

		dirty := false

		// 尝试读取本地存储的节点
		lks, dirty = getSelectedNodes(true)

		if len(lks) == 0 || dirty {
			lks, err = remeasureDelay()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		}

		startProxy(lks)
	}
}

func printDetailedInfo() {
	// 输出所有支持的核心
	printAsTable([]string{"core", "version"}, [][]string{{"Xray", proxyctl.CoreVersion()}})

	// 输出已配置的订阅地址
	var subData [][]string
	for i, u := range setting.Urls {
		subData = append(subData, []string{
			fmt.Sprintf("%d", i),
			u,
		})
	}
	printAsTable([]string{"tag", "subscription"}, subData)

	// 输出选中的节点
	lks, _ := getSelectedNodes(true)
	var nodeData [][]string
	for i, l := range lks {
		nodeData = append(nodeData, []string{
			fmt.Sprintf("%d", i),
			l.Remarks,
			l.Address,
			fmt.Sprintf("%d", l.Delay),
		})
	}
	printAsTable([]string{"tag", "name", "addr", "ping"}, nodeData)

	// 输出已配置的过滤器
	var filterData = [][]string{}
	for _, f := range setting.Filters {
		filterData = append(filterData, []string{
			f.Tag,
			f.Selector,
		})
	}
	printAsTable([]string{"tag", "selector"}, filterData)
}

func getSelectedNodes(refresh bool) ([]*subscription.Link, bool) {
	var lks []*subscription.Link

	// 标记本地存储的节点测试状态，如果是 true 则需要重新测试所有节点
	dirty := false

	upf, err := os.Open(settings.GetUserProfilePath())
	if err == nil {
		defer upf.Close()
		b, err := io.ReadAll(upf)
		if err == nil {
			err = json.Unmarshal(b, &lks)
			if err != nil {
				return lks, true
			}

			// 重新再测试一次延迟
			if refresh {
				lks = proxyctl.ParallelMeasureDelay(lks, setting.Concurrency, setting.Times, setting.Timeout)
				dirty = len(lks) == 0
			}
		}
	}

	return lks, dirty
}

func matchSelector(lks []*subscription.Link, proxies []*settings.Filter) []*subscription.Link {
	var sellks []*subscription.Link
	if len(lks) == 0 {
		return sellks
	}

	if len(proxies) > 0 {
		for _, proxy := range proxies {
			re := regexp.MustCompile(proxy.Selector)
			var f *subscription.Link
			for _, lk := range lks {
				if re.MatchString(lk.Remarks) {
					lk.Tag = proxy.Tag
					f = lk
					break
				}
			}
			if f == nil {
				log.Debugf("selected proxy no server available: '%s'", proxy.Tag)
				continue
			}
			found := false
			for _, lk := range sellks {
				if lk.Remarks == f.Remarks {
					found = true
					break
				}
			}
			if found {
				log.Debugf("selected proxy already exists: '%s'", proxy.Tag)
				continue
			}
			sellks = append(sellks, f)
		}
	} else {
		sellks = append(sellks, lks[0])
	}

	return sellks
}

func printFastestLink(lks []*subscription.Link) {
	if len(lks) == 1 {
		lk := lks[0]
		fmt.Printf("the fastest server is '%s', latency: %dms\n", lk.Remarks, lk.Delay)
	} else {
		for _, lk := range lks {
			fmt.Printf("selected proxy: '%s', the fastest server is '%s', latency: %dms\n", lk.Tag, lk.Remarks, lk.Delay)
		}
	}
}

func startProxy(lks []*subscription.Link) {
	x, err := proxyctl.Start(lks, setting)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	defer x.Close()

	// 监听程序关闭
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	// 等待程序关闭消息
	<-ch

	log.Info("stop service successfully")
}

func remeasureDelay() ([]*subscription.Link, error) {
	lks, err := subscription.Fetch(setting.Urls)
	if err != nil {
		return nil, err
	}

	outlks := proxyctl.ParallelMeasureDelay(lks, setting.Concurrency, setting.Times, setting.Timeout)
	final := matchSelector(outlks, setting.Filters)
	if len(final) == 0 {
		return nil, errors.New("no server available")
	}

	// 尝试将节点写到本地
	b, err := json.Marshal(final)
	if err == nil {
		os.WriteFile(settings.GetUserProfilePath(), b, 0644)
	}

	return final, nil
}

func printAsTable(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)
	table.Render()
	fmt.Println()
}
