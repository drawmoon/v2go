package main

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

func ping(remarks string, proxy string, times int, timeout uint64) int32 {
	proxyUrl, _ := url.Parse(proxy)
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
	req, _ := http.NewRequest("GET", "https://www.google.com/generate_204", nil)
	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
	}
	total := 0
	skip := false
	for i := 0; i < times; i++ {
		elapsedMillis, err := realPing(client, req)
		if err != nil || elapsedMillis == -1 {
			skip = true
			break
		}
		total += int(elapsedMillis)
	}
	if skip {
		log.Printf("ping '%s' timeout", remarks)
		return -1
	}
	elapsedMillis := total / times
	log.Printf("ping '%s' average elapsed %dms", remarks, elapsedMillis)
	return int32(elapsedMillis)
}

func realPing(client *http.Client, req *http.Request) (int64, error) {
	now := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return -1, errors.New("ping failed")
	}
	defer res.Body.Close()
	elapsedMillis := time.Since(now).Milliseconds()
	if res.StatusCode == 204 && res.ContentLength == 0 {
		return elapsedMillis, nil
	}
	return -1, nil
}
