package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var settings *Settings
var ctl *ProxyCtl

func init() {
	settings = loadSettings("v2go.json")

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetLevel(func() (level log.Level) {
		switch settings.LogLevel {
		case "debug":
			level = log.DebugLevel
		case "info":
			level = log.InfoLevel
		case "warn":
			level = log.WarnLevel
		case "error":
			level = log.ErrorLevel
		default:
			level = log.InfoLevel
		}
		return
	}())

	log.Debugf("application home set to %s", settings.ApplicationHome)
	ctl = createProxyCtl()
}

func main() {
	servers := fetch(settings.Urls)
	if len(servers) == 0 {
		log.Fatalln("no subscription found")
	}
	fastest := test(servers)
	startFromFastestServer(fastest)
}

func test(servers []*Server) []*Server {
	testAllServerLatency(servers)

	arr := make([]*Server, 0, len(servers))
	for i := range servers {
		if servers[i].Latency > -1 {
			arr = append(arr, servers[i])
		}
	}
	if len(arr) == 0 {
		log.Fatal("no server available")
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Latency < arr[j].Latency
	})

	var finals []*Server
	if len(settings.Proxies) > 0 {
		for _, p := range settings.Proxies {
			re := regexp.MustCompile(p.Selector)
			var final *Server
			for _, s := range arr {
				if re.MatchString(s.Remarks) {
					s.Tag = p.Tag
					final = s
					break
				}
			}
			if final == nil {
				log.Printf("selected proxy no server available: '%s'", p.Tag)
				continue
			}
			found := false
			for _, s := range finals {
				if s.Remarks == final.Remarks {
					found = true
					break
				}
			}
			if found {
				log.Printf("selected proxy already exists: '%s'", p.Tag)
				continue
			}
			finals = append(finals, final)
			log.Printf("selected proxy: '%s', the fastest server is '%s', latency: %dms", p.Tag, final.Remarks, final.Latency)
		}
	} else {
		final := arr[0]
		finals = append(finals, final)
		log.Printf("the fastest server is '%s', latency: %dms", final.Remarks, final.Latency)
	}
	if len(finals) == 0 {
		log.Fatal("no server available")
	}

	return finals
}

func startFromFastestServer(servers []*Server) {
	log.Infof("starting service, choose the %d fastest servers", len(servers))

	path := settings.Profile
	if len(path) == 0 {
		log.Fatal("no profile specified")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(ctl.Home, path)
	}

	for _, l := range settings.Listeners {
		log.Infof("listening on %s %s:%d", l.Protocol, l.Address, l.Port)
	}
	process, err := start(ctl.Core, path, servers, settings.Listeners)
	if err != nil {
		log.Fatal(err)
	}

	// listening program close
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	// wait for program close message
	<-ch

	err = process.Cmd.Process.Kill()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("stop service successfully")
}
