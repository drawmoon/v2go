package settings

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Setting struct {
	Verbose     bool      `json:"verbose"`     // 输出详细日志
	Urls        []string  `json:"urls"`        // 订阅地址
	Core        string    `json:"core"`        // 核心，xray
	Times       int       `json:"times"`       // 测试次数
	Timeout     uint64    `json:"timeout"`     // 测试超时等待时间
	Concurrency int       `json:"concurrency"` // 测试使用的线程数
	UseLocalDns bool      `json:"useLocalDns"` // 使用本地 DNS
	Proxies     []*Proxy  `json:"proxies"`     // 代理配置
	Listens     []*Listen `json:"listens"`     // 监听配置
}

type Proxy struct {
	Selector string `json:"selector"` // 选择器，正则表达式
	Tag      string `json:"tag"`      // 标签
}

type Listen struct {
	Protocol string `json:"protocol"` // 监听协议，http, socks
	Port     uint32 `json:"port"`     // 监听端口
}

func LoadSettings(f string) (*Setting, error) {
	if !filepath.IsAbs(f) {
		workDir, _ := GetWorkDir()
		f = filepath.Join(workDir, "config.json")
	}

	file, err := os.Open(f)
	if err != nil {
		return nil, errors.New("not found app settings file")
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("error reading app settings file")
	}

	var s *Setting
	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, errors.New("error parsing app settings file")
	}

	return s, nil
}

func GetWorkDir() (string, error) {
	h, _ := os.UserHomeDir()
	workDir := filepath.Join(h, ".xrc")

	err := os.MkdirAll(workDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	return workDir, nil
}

func GetUserProfilePath() string {
	workDir, _ := GetWorkDir()
	return filepath.Join(workDir, "user_profile.json")
}
