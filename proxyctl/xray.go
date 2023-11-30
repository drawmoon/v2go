package proxyctl

import (
	"context"
	"encoding/json"
	"fmt"
	"main/settings"
	"main/subscription"
	"net"
	"net/http"
	"time"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"

	commlog "github.com/xtls/xray-core/common/log"
	commnet "github.com/xtls/xray-core/common/net"

	log "github.com/sirupsen/logrus"
)

var (
	defaultDnsAddr = "8.8.8.8"
	localDnsAddr   = "127.0.0.1"
)

type XrayLogger struct {
	verbose bool
}

func (x *XrayLogger) Write(s string) error {
	if x.verbose {
		log.Print(s)
	}
	return nil
}

func (x *XrayLogger) Close() error {
	return nil
}

type Xray struct {
	inst *core.Instance
}

func NewXray(lks []*subscription.Link, listens []*settings.Listen, verbose, localDns bool, mux int16) (*Xray, error) {
	config, err := profile(lks, listens, verbose, localDns, mux)
	if err != nil {
		return nil, err
	}

	// xray log
	commlog.RegisterHandler(commlog.NewLogger(func() commlog.Writer {
		return &XrayLogger{verbose: verbose}
	}))

	inst, err := core.New(config)
	if err != nil {
		return nil, err
	}

	x := &Xray{inst: inst}
	return x, nil
}

func (x *Xray) NewHttpClient(timeout uint64) *http.Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := commnet.ParseDestination(fmt.Sprintf("%s:%s", network, addr))
			if err != nil {
				return nil, err
			}
			return core.Dial(ctx, x.inst, dest)
		},
	}

	c := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
	}

	return c
}

func (x *Xray) Start() error {
	return x.inst.Start()
}

func (x *Xray) Close() error {
	return x.inst.Close()
}

func profile(lks []*subscription.Link, listens []*settings.Listen, verbose, localDns bool, mux int16) (*core.Config, error) {
	dnsAddr := defaultDnsAddr
	if localDns {
		dnsAddr = localDnsAddr
	}

	c := &conf.Config{
		LogConfig:       createLog(verbose),
		DNSConfig:       createDns(dnsAddr),
		InboundConfigs:  createInbound(listens),
		OutboundConfigs: createOutbound(lks, dnsAddr, mux),
		RouterConfig:    createRouter(),
	}

	return c.Build()
}

func createLog(verbose bool) *conf.LogConfig {
	logLevel := commlog.Severity_Error
	if verbose {
		logLevel = commlog.Severity_Debug
	}

	c := &conf.LogConfig{
		LogLevel: logLevel.String(),
	}

	return c
}

func createDns(dnsAddr string) *conf.DNSConfig {
	c := &conf.DNSConfig{
		Tag: "dnsQuery",
		Servers: []*conf.NameServerConfig{{
			Address: &conf.Address{Address: commnet.ParseAddress(dnsAddr)},
		}},
	}

	return c
}

func createInbound(listens []*settings.Listen) []conf.InboundDetourConfig {
	var ins []conf.InboundDetourConfig

	for _, listen := range listens {
		c := conf.InboundDetourConfig{
			Tag:      listen.Protocol,
			Protocol: listen.Protocol,
			ListenOn: &conf.Address{
				Address: commnet.LocalHostIP,
			},
			PortList: &conf.PortList{Range: []conf.PortRange{{
				To:   listen.Port,
				From: listen.Port,
			}}},
			SniffingConfig: &conf.SniffingConfig{
				Enabled:      true,
				DestOverride: &conf.StringList{"http", "tls", "quic"},
			},
			Settings: func() *json.RawMessage {
				s := json.RawMessage([]byte(`{
					"auth": "noauth",
					"udp": true,
					"userLevel": 0
				}`))
				return &s
			}(),
		}

		ins = append(ins, c)
	}

	return ins
}

func createOutbound(lks []*subscription.Link, dnsAddr string, mux int16) []conf.OutboundDetourConfig {
	var outs []conf.OutboundDetourConfig

	for _, lk := range lks {
		if len(lk.Tag) == 0 {
			lk.Tag = "proxy"
		}

		var out conf.OutboundDetourConfig
		switch lk.Protocol {
		case "vmess":
			out = subscription.VmessAsOutbound(lk, mux)
		}

		outs = append(outs, out)
	}

	outs = append(outs, createDnsOutbound(dnsAddr)...)

	return outs
}

func createDnsOutbound(dnsAddr string) []conf.OutboundDetourConfig {
	direct := conf.OutboundDetourConfig{
		Protocol: "freedom",
		Tag:      "direct",
		Settings: func() *json.RawMessage {
			s := json.RawMessage([]byte(`{
				"userLevel": 0
			}`))
			return &s
		}(),
	}

	reject := conf.OutboundDetourConfig{
		Protocol: "blackhole",
		Tag:      "reject",
	}

	dnsout := conf.OutboundDetourConfig{
		Protocol: "dns",
		Tag:      "dnsOut",
		StreamSetting: &conf.StreamConfig{
			SocketSettings: &conf.SocketConfig{
				DialerProxy: "proxy",
			},
		},
		Settings: func() *json.RawMessage {
			s := json.RawMessage([]byte(fmt.Sprintf(`{
				"address": "%s",
				"nonIpQuery": "skip",
				"userLevel": 0
			}`, dnsAddr)))
			return &s
		}(),
	}

	return []conf.OutboundDetourConfig{direct, reject, dnsout}
}

func createRouter() *conf.RouterConfig {
	r := &conf.RouterConfig{
		RuleList: []json.RawMessage{
			json.RawMessage([]byte(`{
				"outboundTag": "proxy",
				"inboundTag": ["dnsQuery"],
				"type": "field"
			}`)),
			json.RawMessage([]byte(`{
				"outboundTag": "dnsOut",
				"port": "53",
				"type": "field"
			}`)),
		},
	}

	return r
}

func CoreVersion() string {
	return core.Version()
}
