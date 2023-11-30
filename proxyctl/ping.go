package proxyctl

import (
	"errors"
	"main/settings"
	"main/subscription"
	"net/http"
	"sort"
	"time"

	"github.com/sourcegraph/conc/pool"

	log "github.com/sirupsen/logrus"
)

func ParallelMeasureDelay(lks []*subscription.Link, conc int, times int, timeout uint64) []*subscription.Link {
	log.Debugf("ping with %d threads", conc)

	p := pool.New().WithMaxGoroutines(conc)
	for _, lk := range lks {
		lk := lk

		p.Go(func() {
			port := pickFreeTcpPort()
			listen := &settings.Listen{
				Protocol: "http",
				Port:     port,
			}
			delay, err := MeasureDelay(lk, listen, times, timeout)
			if err != nil {
				log.Error(err)
				delay = -1
			}

			lk.Delay = delay
		})
	}
	p.Wait()

	var r []*subscription.Link
	for _, lk := range lks {
		if lk.Delay > -1 {
			r = append(r, lk)
		}
	}

	return sortByDelay(r)
}

func MeasureDelay(lk *subscription.Link, listen *settings.Listen, times int, timeout uint64) (int32, error) {
	x, err := NewXray([]*subscription.Link{lk}, []*settings.Listen{listen}, false, false, -1)
	if err != nil {
		return -1, err
	}

	if err := x.Start(); err != nil {
		return -1, err

	}
	defer x.Close()

	c := x.NewHttpClient(timeout)
	return ping(c, times, lk.Remarks), nil
}

func ping(c *http.Client, times int, remarks string) int32 {
	req, _ := http.NewRequest("GET", "https://www.google.com/generate_204", nil)

	total := 0
	skip := false
	for i := 0; i < times; i++ {
		elapsedMillis, err := func() (int64, error) {
			now := time.Now()
			res, err := c.Do(req)
			if err != nil {
				return -1, errors.New("ping failed")
			}
			defer res.Body.Close()
			elapsedMillis := time.Since(now).Milliseconds()
			if res.StatusCode == 204 && res.ContentLength == 0 {
				return elapsedMillis, nil
			}
			return -1, nil
		}()
		if err != nil || elapsedMillis == -1 {
			skip = true
			break
		}
		total += int(elapsedMillis)
	}
	if skip {
		log.Debugf("ping '%s' timeout", remarks)
		return -1
	}
	elapsedMillis := total / times
	log.Debugf("ping '%s' average elapsed %dms", remarks, elapsedMillis)
	return int32(elapsedMillis)
}

func sortByDelay(lks []*subscription.Link) []*subscription.Link {
	sort.Slice(lks, func(a, b int) bool {
		return lks[a].Delay < lks[b].Delay
	})

	return lks
}
