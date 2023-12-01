package cmd

import (
	"fmt"
	"main/settings"
	"main/subscription"

	log "github.com/sirupsen/logrus"
)

func Resub(setting *settings.Setting) {
	lks, err := subscription.Resubscribe(setting.Urls)
	if err != nil {
		log.Fatal(err)
	}

	if len(lks) == 0 {
		fmt.Println("no subscription found")
	} else {
		for i, l := range lks {
			fmt.Printf("[%d] %s\n", i, l.Remarks)
		}
	}
}
