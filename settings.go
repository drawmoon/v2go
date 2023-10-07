package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Settings struct {
	LogLevel        string      `json:"logLevel"`        // 日志级别，debug, info, warn, error
	Urls            []string    `json:"urls"`            // 订阅地址
	Core            string      `json:"core"`            // 核心，xray
	Profile         string      `json:"config"`          // 配置文件路径，只填文件名默认与核心在同一个位置
	Times           int         `json:"times"`           // 测试次数
	Timeout         uint64      `json:"timeout"`         // 测试超时等待时间
	Concurrency     int         `json:"concurrency"`     // 测试使用的线程数
	Proxies         []*Proxy    `json:"proxies"`         // 代理配置
	Listeners       []*Listener `json:"listeners"`       // 监听配置
	ApplicationHome string      `json:"applicationHome"` // 核心所在的目录位置，/path/to/
}

type Proxy struct {
	Selector string `json:"selector"`
	Tag      string `json:"tag"`
}

type Listener struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     uint16 `json:"port"`
}

type Server struct {
	Protocol       string // 代理协议，vmess
	Remarks        string // 别名
	Address        string // 地址
	Port           uint16 // 端口
	Id             string // 用户ID
	AlterId        uint32 // 额外ID
	Security       string // 加密方式，aes-128-gcm，chacha20-poly1305，auto，none，zero
	Network        string // 传输协议，tcp，kcp，ws，h2，quic，grpc
	HeaderType     string // 伪装类型，none
	Host           string // 伪装域名
	Path           string // 路径
	StreamSecurity string // 传输层安全，tls
	SNI            string // 服务器名称指示
	Fingerprint    string // TSL指纹
	ALPN           string // 应用层协议，h2，http/1.1
	AllowInsecure  bool   // 跳过证书验证
	Latency        int32  // 延迟
	Tag            string // 标签
}

func loadSettings(filename string) *Settings {
	dir, _ := os.Getwd()
	path := filepath.Join(dir, filename)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal("not found app settings file")
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("error reading app settings file")
	}

	var settings Settings
	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		log.Fatal("error parsing app settings file")
	}
	if len(settings.ApplicationHome) == 0 {
		settings.ApplicationHome = filepath.Join(dir, "bin")
	}

	return &settings
}
