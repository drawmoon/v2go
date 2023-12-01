package cmd

import (
	"encoding/json"
	"errors"
	"main/proxyctl"
	"main/settings"
	"main/subscription"
	"os"

	log "github.com/sirupsen/logrus"
)

func Retest(setting *settings.Setting) {
	_, err := remeasureDelay(setting)
	if err != nil {
		log.Fatal(err)
	}
}

func remeasureDelay(setting *settings.Setting) ([]*subscription.Link, error) {
	lks, err := subscription.Fetch(setting.Urls)
	if err != nil {
		return nil, err
	}

	outlks := proxyctl.ParallelMeasureDelay(lks, setting.Concurrency, setting.Times, setting.Timeout)
	final := matchSelector(outlks, setting.Proxies)
	if len(final) == 0 {
		return nil, errors.New("no server available")
	}

	// 尝试将节点写到本地
	b, err := json.Marshal(final)
	if err == nil {
		os.WriteFile(settings.GetUserProfilePath(), b, 0644)
	}

	return final, nil
}
