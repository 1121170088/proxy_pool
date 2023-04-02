package rotateproxy

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var socksUrls []string

func readFile(fileName string)  {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)

	socksUrls = nil
	var list []string
	for {
		line, err := rd.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		line = strings.Trim(line, "\n")
		line = strings.Trim(line, "\r")
		line = strings.Trim(line, " ")
		if line == "" {
			continue
		}
		list = append(list, line)
	}
	socksUrls = list
}

func SelectProxy(checkUrl , socksListFile string)  {
	readFile(socksListFile)
	var minTime int64
	var proxy string
	var duar time.Duration
	for _, p := range socksUrls {
		time, avail := checkProxyWithCheckURL(p, checkUrl)
		intTime := int64(time)
		if avail {
			if proxy == "" {
				proxy = p
				minTime = intTime
				duar = time
			} else {
				if intTime < minTime {
					proxy = p
					minTime = intTime
					duar = time
				}
			}
		}
	}
	log.Printf("good proxy  %s , %f s", proxy, duar.Seconds())
	ProxyURL = proxy
}

func checkProxyWithCheckURL(proxyURL string, checkURL string) (timeout time.Duration, avail bool) {
	log.Printf("check %s： %s\n", proxyURL, checkURL)
	proxy, _ := url.Parse("socks5://" + proxyURL)
	httpclient := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyURL(proxy),
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		},
		Timeout: 5 * time.Second,
	}
	startTime := time.Now()
	resp, err := httpclient.Get(checkURL)
	if err != nil {
		log.Printf("%v", err)
		return 0, false
	}
	defer resp.Body.Close()
	since := time.Since(startTime)

	// TODO: support regex
	if resp.StatusCode != 200 {
		log.Printf("status code err %d", resp.StatusCode)
		return 0, false
	}
	log.Printf("ok, elapsed time %f s", since.Seconds())
	return since, true
}