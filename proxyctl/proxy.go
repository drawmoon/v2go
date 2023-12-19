package proxyctl

import (
	"main/settings"
	"main/subscription"

	log "github.com/sirupsen/logrus"
)

func Start(lks []*subscription.Link, setting *settings.Setting, verbose bool) (*Xray, error) {
	log.Debugf("starting service, choose the %d fastest servers", len(lks))

	listens := setting.Listens
	if len(listens) == 0 {
		listens = append(listens, &settings.Listen{
			Protocol: "http",
			Port:     pickFreeTcpPort(),
		})
	}
	for _, l := range listens {
		log.Debugf("listening on %s %s:%d", l.Protocol, "127.0.0.1", l.Port)
	}

	x, err := NewXray(lks, listens, verbose, setting.UseLocalDns, -1)
	if err != nil {
		return nil, err
	}

	if err := x.Start(); err != nil {
		return nil, err
	}

	return x, nil
}
