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
	Filters     []*Filter `json:"filters"`     // 代理配置
	Listens     []*Listen `json:"listens"`     // 监听配置
}

type Filter struct {
	Selector string `json:"selector"` // 选择器，正则表达式
	Tag      string `json:"tag"`      // 标签
}

type Listen struct {
	Protocol string `json:"protocol"` // 监听协议，http, socks
	Port     uint32 `json:"port"`     // 监听端口
}

func (s *Setting) Save() error {
	cp := GetAppSettingPath()
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(cp, b, 0644)
}

func LoadSettings() (*Setting, error) {
	cp := GetAppSettingPath()
	file, err := os.Open(cp)
	if err != nil {
		if os.IsNotExist(err) {
			s := LoadDefaultSettings()
			err := s.Save()
			if err != nil {
				return nil, err
			}
			return s, nil
		} else {
			return nil, err
		}
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

func LoadDefaultSettings() *Setting {
	return &Setting{
		Verbose:     false,
		Times:       10,
		Timeout:     5,
		Concurrency: 12,
		Listens:     []*Listen{{Protocol: "socks", Port: 10888}},
	}
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

func GetAppSettingPath() string {
	workDir, _ := GetWorkDir()
	return filepath.Join(workDir, "config.json")
}

func GetUserProfilePath() string {
	workDir, _ := GetWorkDir()
	return filepath.Join(workDir, "user_profile.json")
}
