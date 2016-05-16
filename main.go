// Just a simple reverse proxy.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

var suppliers = map[string]func()map[string]*url.URL{

	"fileSupplier": func () map[string]*url.URL {
		configFile, err := os.Open("/jasprox.conf")
		if err != nil {
			log.Fatal(err)
		}
		defer configFile.Close()

		splitStatement := func (s string) (string, string) {
			statementSlice := strings.Split(s, " ")
			return statementSlice[0], statementSlice[1]
		}

		configScanner := bufio.NewScanner(configFile)
		upstreamMap := make(map[string]*url.URL)
		for configScanner.Scan() {
			d, u := splitStatement(configScanner.Text())
			upstreamMap[d], err = url.Parse(u)
			if err != nil {
				log.Fatal(err)
			}
		}

		if err := configScanner.Err(); err != nil {
			log.Fatal(err)
		}

		return upstreamMap
	},
}

// makeJasprox makes a new proxy handler.
func makeJasprox() func (http.ResponseWriter, *http.Request) {

	// supplyUpstreams makes an upstream map every second and sends it to
	// the upstreamMaps channel.
	supplyUpstreams := func(upstreamMaps chan map[string]*url.URL,
		                supplierFunc func()map[string]*url.URL) {
		for {
			upMap := supplierFunc()
			upstreamMaps <- upMap
			time.Sleep(3 * time.Second)
		}
	}

	upstreamMaps := make(chan map[string]*url.URL)
	supplierFunc := suppliers["fileSupplier"]
	go supplyUpstreams(upstreamMaps, supplierFunc)

	// proxyRequests recieves requests and sends proxies based on the
	// curent upstream map
	proxyRequests := func(requests chan *http.Request,
		              upMaps chan map[string]*url.URL,
		              proxies chan *httputil.ReverseProxy) {
		upMap := <-upMaps
		for {
			select {
			case r := <-requests:
				u := upMap[r.Host]
				proxy := httputil.NewSingleHostReverseProxy(u)
				proxies <- proxy
			case newUpMap := <-upMaps:
				upMap = newUpMap
			}
		}
	}

	requests := make(chan *http.Request)
	proxies := make(chan *httputil.ReverseProxy)
	go proxyRequests(requests, upstreamMaps, proxies)

	// handler puts requests onto the requests channel, pulls proxies off
	// the proxies channel
	return func (w http.ResponseWriter, r *http.Request) {
		requests <- r
		proxy := <-proxies
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	fmt.Println("This is Jasprox.")
	jasproxHandler := http.HandlerFunc(makeJasprox())
	log.Fatal(http.ListenAndServe(":80", jasproxHandler))
}
