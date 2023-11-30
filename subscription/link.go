package subscription

import (
	"fmt"
	"strings"
)

type Link struct {
	Protocol       string `json:"protocol"`       // 代理协议，vmess
	Version        string `json:"version"`        // 版本
	Remarks        string `json:"remarks"`        // 别名
	Address        string `json:"address"`        // 地址
	Port           string `json:"port"`           // 端口
	Id             string `json:"id"`             // 用户ID
	AlterId        string `json:"alterId"`        // 额外ID
	Security       string `json:"security"`       // 加密方式，aes-128-gcm，chacha20-poly1305，auto，none，zero
	Network        string `json:"network"`        // 传输协议，tcp，kcp，ws，h2，quic，grpc
	HeaderType     string `json:"headerType"`     // 伪装类型，none
	Host           string `json:"host"`           // 伪装域名
	Path           string `json:"path"`           // 路径
	StreamSecurity string `json:"streamSecurity"` // 传输层安全，tls
	Sni            string `json:"sni"`            // 服务器名称指示
	Fingerprint    string `json:"fingerprint"`    // TSL指纹
	Alpn           string `json:"alpn"`           // 应用层协议，h2，http/1.1
	AllowInsecure  bool   `json:"allowInsecure"`  // 跳过证书验证
	Delay          int32  `json:"delay"`          // 延迟
	Tag            string `json:"tag"`            // 标签
}

func NewLink(s string) (*Link, error) {
	if strings.HasPrefix(s, "vmess://") {
		v, err := NewVmessLink(s)
		if err != nil {
			return nil, err
		}

		return v.AsLink(), nil
	} else {
		return nil, fmt.Errorf("unsupported protocol: %s", s)
	}
}
