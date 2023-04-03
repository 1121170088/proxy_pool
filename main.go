package main

import (
	"flag"
	"proxy_pool/rotateproxy"
	"time"
)

var (
	ListenAddr string
	checkURL    string
	socksListFile string
)

func init() {
	flag.StringVar(&ListenAddr, "l", "127.0.0.1:7777", "listen address")
	flag.StringVar(&checkURL, "check", `https://github.com`, "check url")
	flag.StringVar(&rotateproxy.DefaultProxy, "defaultProxy", `127.0.0.1:7890`, "defaultProxy")
	flag.StringVar(&socksListFile, "socksListFile", `socks5.txt`, "socksListFile")
	flag.Parse()
}

func main() {
	go func() {
		for {
			rotateproxy.SelectProxy(checkURL, socksListFile)
			time.Sleep(5 * time.Minute)
		}
	}()
	c := rotateproxy.NewRedirectClient(ListenAddr)
	c.Serve()
	select {}
}
