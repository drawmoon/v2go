xrc
===

```sh
$ xrc --help

__  ___ __ ___
\ \/ / '__/ __|
 >  <| | | (__
/_/\_\_|  \___|
A xray client.

Usage:
  xrc [flags]
  xrc [command]

Available Commands:
  start         Start proxy
  ls            View configuration details
  rmu           Remove subscription
  rmf           Remove filter

Flags:
  -url          Append subscription address
  -sub          Retrieve subscriptions
  -ping         Test node delay
  -f            Filter test nodes
  -http         Set http proxy port
  -socks        Set socks proxy port
  -v            Show more logs

Example:
  xrc -url https://your-subscription-url -sub -ping
  xrc -socks 1088 -http 1099 start
```

Usage examples
--------------

Set up a subscription and measure the delay

```sh
$ xrc -url https://your-subscription-url -sub -ping

the fastest server is 'the-xxx-server-node', latency: 89ms
```

Set up a subscription and filter nodes

```sh
$ xrc -url https://your-subscription-url -sub -ping -f abc

the fastest server is 'the-abc-server-node', latency: 89ms
```

View configuration details

```sh
$ xrc ls

CORE    VERSION
Xray    1.8.4

TAG     SUBSCRIPTION
0       https://your-subscription-url

TAG     NAME                    ADDR                    PING
0       the-xxx-server-node     xxx.cn                  89

TAG     SELECTOR
```
