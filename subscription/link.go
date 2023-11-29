package subscription

import (
	"fmt"
	"strings"
)

type Link struct {
	Protocol       string // 代理协议，vmess
	Version        string // 版本
	Remarks        string // 别名
	Address        string // 地址
	Port           string // 端口
	Id             string // 用户ID
	AlterId        string // 额外ID
	Security       string // 加密方式，aes-128-gcm，chacha20-poly1305，auto，none，zero
	Network        string // 传输协议，tcp，kcp，ws，h2，quic，grpc
	HeaderType     string // 伪装类型，none
	Host           string // 伪装域名
	Path           string // 路径
	StreamSecurity string // 传输层安全，tls
	Sni            string // 服务器名称指示
	Fingerprint    string // TSL指纹
	Alpn           string // 应用层协议，h2，http/1.1
	AllowInsecure  bool   // 跳过证书验证
	Delay          int32  // 延迟
	Tag            string // 标签
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
