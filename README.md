# v2go

Automatically select the fastest server as [Xray](https://github.com/XTLS/Xray-core) Outbounds and update the config file.

```
               _                    
 __      _____| |_ ___   __ _  ___  
 \ \ /\ / / _ \ __/ _ \ / _` |/ _ \ 
  \ V  V /  __/ || (_) | (_| | (_) |
   \_/\_/ \___|\__\___/ \__, |\___/ 
                        |___/       

INFO[0000] fetching subscriptions
INFO[0001] found 27 subscriptions
INFO[0001] ping with 12 threads
INFO[0011] ping 'github.com/freefq - 美国加利福尼亚州圣何塞PEG TECH数据中心 9' average elapsed 282ms
INFO[0014] ping 'github.com/freefq - 美国加利福尼亚州圣何塞MULTACOM机房 6' average elapsed 624ms
INFO[0015] ping 'github.com/freefq - 香港  3' average elapsed 662ms
...
INFO[0029] selected proxy: 'cf', the fastest server is 'github.com/freefq - 美国CloudFlare公司CDN节点 21', latency: 452ms
INFO[0029] selected proxy: 'hk', the fastest server is 'github.com/freefq - 香港  12', latency: 599ms
INFO[0029] starting service, choose the 2 fastest servers
INFO[0029] listening on http 127.0.0.1:10809
INFO[0029] listening on socks 127.0.0.1:10808
```

## TODO

- [x] Automatically selects the optimal server based on a selector
- [ ] Configure timed tasks to test server latency
- [ ] Manually or automatically check and update the core
- [ ] Support the SSR

## Credits

- [Xray](https://github.com/XTLS/Xray-core)
- [Clash](https://github.com/Dreamacro/clash)
- [Shadowsocks](https://github.com/shadowsocks/go-shadowsocks2)
- [v2ray-maid](https://github.com/mokeyish/v2ray-maid)
- [freefq](https://github.com/freefq/free)
