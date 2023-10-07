package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Xray struct {
	Home    string
	Version string
}

func (x *Xray) Start(file string) (*exec.Cmd, error) {
	cmd := exec.Command(x.Home, "-c", file)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?i)(?:x|v2)ray \d+\.\d+\.\d+ started`)
	reader := bufio.NewReader(stdout)
	for {
		str, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("unable to get service status: %s", x.Home)
		}
		if strings.HasPrefix(str, "Failed to start:") {
			return nil, errors.New("failed to start")
		}
		if re.MatchString(str) {
			break
		}
	}
	return cmd, nil
}

func (x *Xray) GenerateFullProfile(servers []*Server, listeners []*Listener) (string, error) {
	conf := &XrayConfig{
		Log:             generateLogConfig(),
		InboundDetours:  generateInboundConfig(listeners),
		OutboundDetours: generateOutboundConfig(servers),
		Router:          generateRouterConfig(),
	}
	bytes, err := json.Marshal(conf)
	if err != nil {
		return "", err
	}
	str := string(bytes)
	return str, nil
}

func (x *Xray) Validate(link *Server) bool {
	return link.Protocol != "vmess"
}

func generateLogConfig() *LogConfig {
	return &LogConfig{
		Level: "warning",
	}
}

func generateInboundConfig(listeners []*Listener) []*InboundDetourConfig {
	var inboundDetours []*InboundDetourConfig
	for _, l := range listeners {
		inbound := &InboundDetourConfig{
			Tag:      l.Protocol,
			Protocol: l.Protocol,
			ListenOn: l.Address,
			Port:     l.Port,
			Sniffing: &SniffingConfig{
				Enabled:      true,
				DestOverride: []string{"http", "tls"},
				RouteOnly:    false,
			},
			Settings: &InboundSettingsConfig{
				Auth:             "noauth",
				UDP:              true,
				AllowTransparent: false,
			},
		}
		inboundDetours = append(inboundDetours, inbound)
	}
	return inboundDetours
}

func generateOutboundConfig(servers []*Server) []*OutboundDetourConfig {
	outboundDetours := make([]*OutboundDetourConfig, 0, len(servers)+2)
	for _, s := range servers {
		if len(s.Tag) == 0 {
			s.Tag = "proxy"
		}
		proxy := &OutboundDetourConfig{
			Tag:      s.Tag,
			Protocol: s.Protocol,
			Settings: &OutboundSettingsConfig{
				VNextConfigs: []*VNextConfig{
					{
						Address: s.Address,
						Port:    s.Port,
						Users: []*UserConfig{{
							Id:       s.Id,
							AlterId:  s.AlterId,
							Security: s.Security,
							Level:    8,
						}},
					},
				},
			},
			MuxSettings: &MuxSettingConfig{
				Enabled:     true,
				Concurrency: -1,
			},
		}
		proxy.StreamSettings = func() *StreamConfig {
			stream := &StreamConfig{
				Network:  s.Network,
				Security: s.StreamSecurity,
			}
			if s.Network == "ws" {
				webSocket := &WebSocketConfig{
					Path:    s.Path,
					Headers: make(map[string]string),
				}
				if len(s.Host) > 0 {
					webSocket.Headers["host"] = s.Host
					stream.TLSSettings = &TLSConfig{
						AllowInsecure: false,
						ServerName:    s.Host,
					}
				}
				stream.WSSettings = webSocket
			} else {
				// TODO: to be supported
				log.Printf("network protocol '%s' not supported", stream.Network)
			}
			return stream
		}()
		outboundDetours = append(outboundDetours, proxy)
	}
	// generate direct and block outbound detours config
	outboundDetours = append(outboundDetours, func() []*OutboundDetourConfig {
		return []*OutboundDetourConfig{
			{
				Tag:      "direct",
				Protocol: "freedom",
			},
			{
				Tag:      "block",
				Protocol: "blackhole",
				Settings: &OutboundSettingsConfig{
					Response: &ResponseConfig{
						Type: "http",
					},
				},
			},
		}
	}()...)
	return outboundDetours
}

func generateRouterConfig() *RouterConfig {
	return &RouterConfig{
		DomainStrategy: "IPOnDemand",
		Rules: []*RouterRuleConfig{{
			Type:    "field",
			IP:      []string{"geoip:private"},
			Tag:     "direct",
			Enabled: true,
		}},
	}
}

func createXray(src string) (*Xray, error) {
	log.Debugf("creating xray, src: %s", src)
	cmd := exec.Command(src, "--version")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New("unable to get version")
	}
	outputStr := string(output)
	re := regexp.MustCompile(`(?i)(?:x|v2)ray (\d+\.\d+\.\d+) `)
	matches := re.FindStringSubmatch(outputStr)
	version := matches[1]
	log.Debugf("created xray version is %s", version)
	return &Xray{src, version}, nil
}

type XrayConfig struct {
	Log             *LogConfig              `json:"log,omitempty"`
	Router          *RouterConfig           `json:"routing,omitempty"`
	DNS             *json.RawMessage        `json:"dns,omitempty"`
	InboundDetours  []*InboundDetourConfig  `json:"inbounds,omitempty"`
	OutboundDetours []*OutboundDetourConfig `json:"outbounds,omitempty"`
	Transport       *json.RawMessage        `json:"transport,omitempty"`
	Policy          *json.RawMessage        `json:"policy,omitempty"`
	API             *json.RawMessage        `json:"api,omitempty"`
	Metrics         *json.RawMessage        `json:"metrics,omitempty"`
	Stats           *json.RawMessage        `json:"stats,omitempty"`
	Reverse         *json.RawMessage        `json:"reverse,omitempty"`
	FakeDNS         *json.RawMessage        `json:"fakeDns,omitempty"`
	Observatory     *json.RawMessage        `json:"observatory,omitempty"`
}

type LogConfig struct {
	AccessLog string `json:"access,omitempty"`
	ErrorLog  string `json:"error,omitempty"`
	Level     string `json:"loglevel,omitempty"`
	DNSLog    bool   `json:"dnsLog"`
}

type RouterConfig struct {
	Rules          []*RouterRuleConfig `json:"rules,omitempty"`
	DomainStrategy string              `json:"domainStrategy,omitempty"`
	Balancers      []*json.RawMessage  `json:"balancers,omitempty"`
	DomainMatcher  string              `json:"domainMatcher,omitempty"`
}

type RouterRuleConfig struct {
	Id      string   `json:"id,omitempty"`
	Type    string   `json:"type,omitempty"`
	IP      []string `json:"ip,omitempty"`
	Tag     string   `json:"outboundTag,omitempty"`
	Domain  []string `json:"domain,omitempty"`
	Enabled bool     `json:"enabled"`
}

type InboundDetourConfig struct {
	Protocol       string                         `json:"protocol,omitempty"`
	Port           uint16                         `json:"port,omitempty"`
	ListenOn       string                         `json:"listen,omitempty"`
	Settings       *InboundSettingsConfig         `json:"settings,omitempty"`
	Tag            string                         `json:"tag,omitempty"`
	Allocation     *InboundDetourAllocationConfig `json:"allocate,omitempty"`
	StreamSetting  *json.RawMessage               `json:"streamSettings,omitempty"`
	DomainOverride *json.RawMessage               `json:"domainOverride,omitempty"`
	Sniffing       *SniffingConfig                `json:"sniffing,omitempty"`
}

type InboundSettingsConfig struct {
	Auth             string `json:"auth,omitempty"`
	UDP              bool   `json:"udp"`
	AllowTransparent bool   `json:"allowTransparent"`
}

type InboundDetourAllocationConfig struct {
	Strategy    string `json:"strategy,omitempty"`
	Concurrency uint32 `json:"concurrency,omitempty"`
	RefreshMin  uint32 `json:"refresh,omitempty"`
}

type SniffingConfig struct {
	Enabled         bool     `json:"enabled"`
	DestOverride    []string `json:"destOverride,omitempty"`
	DomainsExcluded []string `json:"domainsExcluded,omitempty"`
	MetadataOnly    bool     `json:"metadataOnly"`
	RouteOnly       bool     `json:"routeOnly"`
}

type OutboundDetourConfig struct {
	Protocol       string                  `json:"protocol,omitempty"`
	SendThrough    *json.RawMessage        `json:"sendThrough,omitempty"`
	Tag            string                  `json:"tag,omitempty"`
	Settings       *OutboundSettingsConfig `json:"settings,omitempty"`
	StreamSettings *StreamConfig           `json:"streamSettings,omitempty"`
	ProxySettings  *json.RawMessage        `json:"proxySettings,omitempty"`
	MuxSettings    *MuxSettingConfig       `json:"mux,omitempty"`
}

type OutboundSettingsConfig struct {
	VNextConfigs []*VNextConfig  `json:"vnext,omitempty"`
	Response     *ResponseConfig `json:"response,omitempty"`
}

type VNextConfig struct {
	Address string        `json:"address,omitempty"`
	Port    uint16        `json:"port,omitempty"`
	Users   []*UserConfig `json:"users,omitempty"`
}

type UserConfig struct {
	Id       string `json:"id,omitempty"`
	AlterId  uint32 `json:"alterId,omitempty"`
	Security string `json:"security,omitempty"`
	Level    uint16 `json:"level,omitempty"`
}

type ResponseConfig struct {
	Type string `json:"type,omitempty"`
}

type StreamConfig struct {
	Network     string           `json:"network,omitempty"`
	Security    string           `json:"security,omitempty"`
	WSSettings  *WebSocketConfig `json:"wsSettings,omitempty"`
	TLSSettings *TLSConfig       `json:"tlsSettings,omitempty"`
}

type WebSocketConfig struct {
	Path                string            `json:"path,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	AcceptProxyProtocol bool              `json:"acceptProxyProtocol,omitempty"`
}

type TLSConfig struct {
	AllowInsecure                        bool             `json:"allowInsecure"`
	ServerName                           string           `json:"serverName,omitempty"`
	ALPN                                 *json.RawMessage `json:"alpn,omitempty"`
	Certificates                         *json.RawMessage `json:"certificates,omitempty"`
	DisableSystemRoot                    bool             `json:"disableSystemRoot"`
	EnableSessionResumption              bool             `json:"enableSessionResumption"`
	MinVersion                           string           `json:"minVersion,omitempty"`
	MaxVersion                           string           `json:"maxVersion,omitempty"`
	CipherSuites                         string           `json:"cipherSuites,omitempty"`
	PreferServerCipherSuites             bool             `json:"preferServerCipherSuites"`
	Fingerprint                          string           `json:"fingerprint,omitempty"`
	RejectUnknownSNI                     bool             `json:"rejectUnknownSni"`
	PinnedPeerCertificateChainSha256     *[]string        `json:"pinnedPeerCertificateChainSha256,omitempty"`
	PinnedPeerCertificatePublicKeySha256 *[]string        `json:"pinnedPeerCertificatePublicKeySha256,omitempty"`
}

type MuxSettingConfig struct {
	Enabled     bool  `json:"enabled"`
	Concurrency int32 `json:"concurrency,omitempty"`
}
