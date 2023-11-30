package main

import (
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

	lks, err := subscription.Fetch(setting.Urls)
	if err != nil {
		log.Fatal(err)
	}

	conc := func() int {
		if setting.Concurrency <= 0 {
			return 1
		}
		return setting.Concurrency
	}()

	outlks := proxyctl.ParallelMeasureDelay(lks, conc, setting.Times, setting.Timeout)
	sellks := selectLink(outlks, setting.Proxies)
	if len(sellks) == 0 {
		log.Fatal("no server available")
	}

	printFastestLink(sellks)
	startProxy(sellks)
}

func selectLink(lks []*subscription.Link, proxies []*settings.Proxy) []*subscription.Link {
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
