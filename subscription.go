package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type VMess struct {
	V             string `json:"v"`
	PS            string `json:"ps"`
	Add           string `json:"add"`
	Port          uint16 `json:"-"`
	Id            string `json:"id"`
	AlterId       uint32 `json:"-"`
	Security      string `json:"security"`
	Net           string `json:"net"`
	Type          string `json:"type"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	Tls           string `json:"tls"`
	Sni           string `json:"sni"`
	Fp            string `json:"fp"`
	ALPN          string `json:"alpn"`
	SkiCertVerify bool   `json:"skip-cert-verify"`
}

func (v *VMess) UnmarshalJSON(data []byte) error {
	type Alias VMess
	aux := &struct {
		Scy  string      `json:"scy"`
		Port json.Number `json:"port"`
		AId  json.Number `json:"aid"`
		*Alias
	}{
		Alias: (*Alias)(v),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	port, err := aux.Port.Int64()
	if err != nil {
		return err
	}
	v.Port = uint16(port)
	aid, err := aux.AId.Int64()
	if err != nil {
		return err
	}
	v.AlterId = uint32(aid)
	if len(aux.Scy) != 0 {
		v.Security = aux.Scy
	} else if len(aux.Alias.Security) == 0 {
		v.Security = "auto"
	}
	return nil
}

func (v *VMess) Into() *Server {
	return &Server{
		Protocol:       "vmess",
		Remarks:        v.PS,
		Address:        v.Add,
		Port:           v.Port,
		Id:             v.Id,
		AlterId:        v.AlterId,
		Security:       v.Security,
		Network:        v.Net,
		HeaderType:     v.Type,
		Host:           v.Host,
		Path:           v.Path,
		StreamSecurity: v.Tls,
		SNI:            v.Sni,
		Fingerprint:    v.Fp,
		ALPN:           v.ALPN,
		AllowInsecure:  v.SkiCertVerify,
	}
}

func fetch(urls []string) []*Server {
	log.Println("fetching subscriptions")
	var servers []*Server
	for _, url := range urls {
		req, _ := http.NewRequest("GET", url, nil)
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		res, err := client.Do(req)
		if err != nil || res.StatusCode != 200 {
			log.Fatalf("fetch subscription failed, url: %s", url)
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		decoded, _ := base64.StdEncoding.DecodeString(string(body))
		lines := strings.Split(string(decoded), "\n")
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			if strings.HasPrefix(line, "vmess://") {
				buf, _ := base64.StdEncoding.DecodeString(line[8:])
				var vmess VMess
				err := json.Unmarshal(buf, &vmess)
				if err != nil {
					log.Printf("unmarshal vmess failed: %s", string(buf))
					continue
				}
				servers = append(servers, vmess.Into())
			} else {
				log.Printf("skip protocol: %s", strings.Split(line, "://")[0])
			}
		}
	}
	log.Printf("found %d subscriptions", len(servers))
	return servers
}
