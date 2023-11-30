package main

import (
	"encoding/json"
	"io"
	"main/proxyctl"
	"main/settings"
	"main/subscription"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var workDir string
var setting *settings.Setting
var err error

func init() {
	h, _ := os.UserHomeDir()
	workDir = filepath.Join(h, ".xrc")

	err = os.MkdirAll(workDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	setting, err = settings.LoadSettings(filepath.Join(workDir, "config.json"))
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
	log.Debugf("core version: %s", proxyctl.CoreVersion())

	userProfile := filepath.Join(workDir, "user_profile.json")
	var final []*subscription.Link

	// 标记本地存储的节点测试状态，如果是 true 则需要重新测试所有节点
	dirty := false

	// 尝试读取本地存储的节点
	upf, err := os.Open(userProfile)
	if err == nil {
		defer upf.Close()
		b, err := io.ReadAll(upf)
		if err == nil {
			err = json.Unmarshal(b, &final)

			// 重新再测试一次延迟
			if err == nil {
				final = proxyctl.ParallelMeasureDelay(final, setting.Concurrency, setting.Times, setting.Timeout)
				dirty = len(final) == 0
			}
		}
	}

	if len(final) == 0 || dirty {
		lks, err := subscription.Fetch(setting.Urls)
		if err != nil {
			log.Fatal(err)
		}

		outlks := proxyctl.ParallelMeasureDelay(lks, setting.Concurrency, setting.Times, setting.Timeout)
		final = matchSelector(outlks, setting.Proxies)
		if len(final) == 0 {
			log.Fatal("no server available")
		}

		// 尝试将节点写到本地
		b, err := json.Marshal(final)
		if err == nil {
			os.WriteFile(userProfile, b, 0644)
		}
	}

	printFastestLink(final)
	startProxy(final)
}

func matchSelector(lks []*subscription.Link, proxies []*settings.Proxy) []*subscription.Link {
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
		f := lks[0]
		log.Printf("the fastest server is '%s', latency: %dms", f.Remarks, f.Delay)
	} else {
		for _, lk := range lks {
			log.Printf("selected proxy: '%s', the fastest server is '%s', latency: %dms", lk.Tag, lk.Remarks, lk.Delay)
		}
	}
}

func startProxy(lks []*subscription.Link) {
	x, err := proxyctl.Start(lks, setting)
	if err != nil {
		log.Fatal(err)
	}
	defer x.Close()

	// 监听程序关闭
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	// 等待程序关闭消息
	<-ch

	log.Info("stop service successfully")
}
