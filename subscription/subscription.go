package subscription

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func Fetch(urls []string) ([]*Link, error) {
	log.Debugln("fetching subscriptions")

	var lks []*Link
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

		content, _ := io.ReadAll(res.Body)
		b, err := base64.StdEncoding.DecodeString(string(content))
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
	}

	log.Debugf("found %d subscriptions", len(lks))
	return lks, nil
}
