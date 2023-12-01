package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"main/proxyctl"
	"main/settings"
	"main/subscription"
	"os"
)

func List(setting *settings.Setting) {
	lks, _ := getLocalSelectedNodes(setting, false)

	if len(lks) == 0 {
		fmt.Println("no selected nodes found")
	} else {
		for i, l := range lks {
			fmt.Printf("[%d] %s\n", i, l.Remarks)
		}
	}
}

func getLocalSelectedNodes(setting *settings.Setting, refresh bool) ([]*subscription.Link, bool) {
	var lks []*subscription.Link

	// 标记本地存储的节点测试状态，如果是 true 则需要重新测试所有节点
	dirty := false

	upf, err := os.Open(settings.GetUserProfilePath())
	if err == nil {
		defer upf.Close()
		b, err := io.ReadAll(upf)
		if err == nil {
			err = json.Unmarshal(b, &lks)

			// 重新再测试一次延迟
			if err == nil && refresh {
				lks = proxyctl.ParallelMeasureDelay(lks, setting.Concurrency, setting.Times, setting.Timeout)
				dirty = len(lks) == 0
			}
		}
	}

	return lks, dirty
}
