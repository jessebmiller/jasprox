package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func fixedJbmMicroBackend(r *http.Request) *url.URL {
	switch r.Host {
	case "general.show":
		r.URL.Host = "localhost:8000"
	case "jessebmiller.com":
		r.URL.Host = "localhost:8001"
	default:
		r.URL.Host = "localhost:8001"
	}

	return r.URL
}

func jasHandler(w http.ResponseWriter, r *http.Request) {
	proxyURL := fixedJbmMicroBackend(r)
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	proxy.ServeHTTP(w, r)
	fmt.Println(r.URL, proxyURL)
}

func main() {
	fmt.Println("This is Jasprox.")
	log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(jasHandler)))
}
