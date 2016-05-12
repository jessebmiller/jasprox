// Just a simple reverse proxy.
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)


const USE_STRAT string = "environ"
var STRATS = map[string]func(*http.Request) (*url.URL, error){

	"environ": environStrategy,

	"justJesse": func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://jessebmiller.com:8000")
	},
}


func environStrategy(r *http.Request) (*url.URL, error) {
	return url.Parse(os.Getenv("PROXY_URL"))
}


// whichUrl chooses which url the request should proxy to, or returns an error
func whichUrl(r *http.Request) (*url.URL, error) {
	return STRATS[USE_STRAT](r)
}


// serveProxy chooses a url then proxies the request, or responds not found.
func serveProxy(w http.ResponseWriter, r *http.Request) {
	proxyUrl, err := whichUrl(r)
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
	} else {
		proxy := httputil.NewSingleHostReverseProxy(proxyUrl)
		proxy.ServeHTTP(w, r)
	}
}


func main() {
	fmt.Println("This is Jasprox.")
	proxyHandler := http.HandlerFunc(serveProxy)
	log.Fatal(http.ListenAndServe(":80", proxyHandler))
}
