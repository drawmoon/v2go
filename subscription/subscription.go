package subscription

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var tmpFile = filepath.Join(os.TempDir(), "xrc_subscription.tmp")

func Fetch(urls []string) ([]*Link, error) {
	var lks []*Link
	var err error

	// 尝试从缓存文件中读取订阅内容
	cacheFile, err := os.Open(tmpFile)
	if err == nil {
		defer cacheFile.Close()
		b, err := io.ReadAll(cacheFile)
		if err == nil {
			lines := strings.Split(string(b), "\n")
			for _, s := range lines {
				if len(s) == 0 {
					continue
				}
				t, err := parseSubscriptionContent(s)
				if err != nil {
					break
				}

				lks = append(lks, t...)
			}
		}
	}

	if len(lks) == 0 {
		lks, err = Resubscribe(urls)
		if err != nil {
			return nil, err
		}
	}

	return lks, nil
}

func Resubscribe(urls []string) ([]*Link, error) {
	log.Debugln("fetching subscriptions")
	var lks []*Link
	var b64StrArr []string

	for _, url := range urls {
		req, _ := http.NewRequest("GET", url, nil)
		c := &http.Client{
			Timeout: 10 * time.Second,
		}

		res, err := c.Do(req)
		if err != nil || res.StatusCode != 200 {
			return nil, fmt.Errorf("fetch subscription failed, url: %s", url)
		}
		defer res.Body.Close()

		b, _ := io.ReadAll(res.Body)
		b64Str := string(b)
		t, err := parseSubscriptionContent(b64Str)
		if err != nil {
			return nil, err
		}

		lks = append(lks, t...)
		b64StrArr = append(b64StrArr, b64Str)
	}

	len := len(lks)
	log.Debugf("found %d subscriptions", len)

	// 尝试将订阅内容写入缓存文件
	if len > 0 {
		os.WriteFile(tmpFile, []byte(strings.Join(b64StrArr, "\n")), 0644)
	}

	return lks, nil
}

func parseSubscriptionContent(content string) ([]*Link, error) {
	var lks []*Link

	b, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, errors.New("decode subscription failed")
	}

	lines := strings.Split(string(b), "\n")
	for _, s := range lines {
		if len(s) == 0 {
			continue
		}
		lk, err := NewLink(s)
		if err != nil {
			log.Warn(err.Error())
		}
		lks = append(lks, lk)
	}

	return lks, nil
}
