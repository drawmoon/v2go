package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xtls/xray-core/infra/conf"
)

type VmessLink struct {
	Ver            string      `json:"v"`
	Ps             string      `json:"ps"`
	Add            string      `json:"add"`
	Port           interface{} `json:"port"`
	Id             string      `json:"id"`
	Aid            interface{} `json:"aid"`
	Security       string      `json:"security"`
	Scy            string      `json:"scy"` // Security 的别名
	Net            string      `json:"net"`
	Type           string      `json:"type"`
	Host           string      `json:"host"`
	Path           string      `json:"path"`
	Tls            string      `json:"tls"`
	Sni            string      `json:"sni"`
	Fp             string      `json:"fp"`
	Alpn           string      `json:"alpn"`
	SkipCertVerify bool        `json:"skip-cert-verify"`
}

func NewVmessLink(vmess string) (*VmessLink, error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		return nil, fmt.Errorf("vmess unreconized: %s", vmess)
	}

	b, err := base64.StdEncoding.DecodeString(vmess[8:])
	if err != nil {
		return nil, err
	}

	v := &VmessLink{}
	if err := json.Unmarshal(b, v); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *VmessLink) AsLink() *Link {
	lk := &Link{
		Protocol:       "vmess",
		Version:        v.Ver,
		Remarks:        v.Ps,
		Address:        v.Add,
		Port:           fmt.Sprintf("%v", v.Port),
		Id:             v.Id,
		AlterId:        fmt.Sprintf("%v", v.Aid),
		Security:       v.Security,
		Network:        v.Net,
		HeaderType:     v.Type,
		Host:           v.Host,
		Path:           v.Path,
		StreamSecurity: v.Tls,
		Sni:            v.Sni,
		Fingerprint:    v.Fp,
		Alpn:           v.Alpn,
		AllowInsecure:  v.SkipCertVerify,
	}

	lk.Security = func() string {
		if len(v.Scy) != 0 {
			return v.Scy
		}
		if len(v.Security) == 0 {
			return v.Security
		}
		return "auto"
	}()

	return lk
}

func VmessAsOutbound(lk *Link, mux int16) conf.OutboundDetourConfig {
	c := conf.OutboundDetourConfig{
		Tag:      lk.Tag,
		Protocol: lk.Protocol,
	}

	settings := json.RawMessage([]byte(fmt.Sprintf(`{
		"vnext": [
			{
				"address": "%s",
				"port": %v,
				"users": [{"id": "%s", "level": 0, "security": "%s"}]
			}
		]
	}`, lk.Address, lk.Port, lk.Id, lk.Security)))
	c.Settings = &settings

	c.StreamSetting = func() *conf.StreamConfig {
		network := conf.TransportProtocol(lk.Network)
		stream := &conf.StreamConfig{
			Network:  &network,
			Security: lk.StreamSecurity,
		}

		switch lk.Network {
		case "ws":
			stream.WSSettings = &conf.WebSocketConfig{
				Path: lk.Path,
				Headers: map[string]string{
					"host": lk.Host,
				},
			}
		}

		if lk.StreamSecurity == "tls" {
			stream.TLSSettings = &conf.TLSConfig{
				Insecure: lk.AllowInsecure,
			}
			if len(lk.Host) > 0 {
				stream.TLSSettings.ServerName = lk.Host
			}
		}

		return stream
	}()

	c.MuxSettings = &conf.MuxConfig{}
	if mux > 0 {
		c.MuxSettings.Enabled = true
		c.MuxSettings.Concurrency = mux
	}

	return c
}
