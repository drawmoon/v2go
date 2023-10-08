package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/sourcegraph/conc/pool"

	log "github.com/sirupsen/logrus"
)

type ProxyCtl struct {
	Core ProxyCore
	Home string
}

type ProxyCore interface {
	Start(file string) (*exec.Cmd, error)
	GenerateFullProfile(servers []*Server, listeners []*Listener) (string, error)
	Validate(server *Server) bool
}

type ProxyCoreProcess struct {
	ConfigPath string
	Cmd        *exec.Cmd
	Listeners  []*Listener
}

func createProxyCtl() *ProxyCtl {
	if settings.Core == "xray" {
		filename := settings.Core
		if runtime.GOOS == "windows" && !strings.HasSuffix(filename, ".exe") {
			filename += ".exe"
		}
		home := filepath.Join(settings.ApplicationHome, settings.Core)
		src := filepath.Join(home, filename)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			log.Fatalf("program not found at: %s", src)
		}
		x, err := createXray(src)
		if err != nil {
			log.Fatal(err)
		}
		return &ProxyCtl{x, home}
	} else {
		log.Fatalf("unsupported core: %s", settings.Core)
	}
	return nil
}

func start(app ProxyCore, path string, servers []*Server, listeners []*Listener) (*ProxyCoreProcess, error) {
	conf, err := app.GenerateFullProfile(servers, listeners)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	err = os.WriteFile(path, []byte(conf), 0644)
	if err != nil {
		return nil, err
	}
	cmd, err := app.Start(path)
	if err != nil {
		return nil, err
	}
	log.Debugf("Starting service, profile at %s", path)
	return &ProxyCoreProcess{path, cmd, listeners}, nil
}

func stop(process *ProxyCoreProcess) error {
	err := process.Cmd.Process.Kill()
	if err != nil {
		return err
	}
	return os.Remove(process.ConfigPath)
}

func testAllServerLatency(servers []*Server) {
	t := func() int {
		conc := settings.Concurrency
		if conc <= 0 {
			conc = 1
		}
		return conc
	}()
	log.Printf("ping with %d threads", t)

	p := pool.New().WithMaxGoroutines(t)
	for _, s := range servers {
		s := s
		if ctl.Core.Validate(s) {
			log.Printf("protocol '%s' not supported", s.Protocol)
			s.Latency = -1
			continue
		}
		p.Go(func() {
			path := filepath.Join(os.TempDir(), fmt.Sprintf("v2go-running-%s.json", uuid.NewString()))
			port := pickFreeTcpPort()
			listeners := []*Listener{
				{
					Protocol: "http",
					Address:  "127.0.0.1",
					Port:     port,
				},
			}
			process, err := start(ctl.Core, path, []*Server{s}, listeners)
			if err != nil {
				log.Fatal(err)
			}
			latency := ping(s.Remarks, fmt.Sprintf("http://127.0.0.1:%d", port), settings.Times, settings.Timeout)
			if err = stop(process); err != nil {
				log.Fatal(err)
			}
			s.Latency = latency
		})
	}
	p.Wait()
}
