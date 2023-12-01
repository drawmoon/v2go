package cmd

import (
	"main/proxyctl"
	"main/settings"
	"main/subscription"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func Run(setting *settings.Setting) {
	var final []*subscription.Link
	var err error

	dirty := false

	// 尝试读取本地存储的节点
	final, dirty = getLocalSelectedNodes(setting, true)

	if len(final) == 0 || dirty {
		final, err = remeasureDelay(setting)
		if err != nil {
			log.Fatal(err)
		}
	}

	printFastestLink(final)
	startProxy(final, setting)
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

func startProxy(lks []*subscription.Link, setting *settings.Setting) {
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
