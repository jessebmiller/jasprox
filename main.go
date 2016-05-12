// Just a simple reverse proxy.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)


const USE_STRAT string = "file"
var STRATS = map[string]func(*http.Request) (*url.URL, error){

	"environ": environStrategy,
	"file": fileStrategy,

}


// fileStrategy reads proxy rules from /jasprox.conf
func fileStrategy(r *http.Request) (*url.URL, error) {
	configFile, err := os.Open("/jasprox.conf")
	if err != nil {
		fmt.Println("ERROR: Could not open file /jasprox.conf")
		return nil, err
	}
	defer configFile.Close()

	splitStatement := func (s string) (string, string) {
		statementSlice := strings.Split(s, " ")
		return statementSlice[0], statementSlice[1]
	}

	configScanner := bufio.NewScanner(configFile)
	for configScanner.Scan() {
		downstream, upstream := splitStatement(configScanner.Text())
		if downstream == r.Host {
			fmt.Println(
				"found proxy rule:",
				downstream,
				"->",
				upstream,
			)
			return url.Parse(upstream)
		}
	}

	if err := configScanner.Err(); err != nil {
		fmt.Println("ERROR: Could not scan config file")
		return nil, err
	}

	return nil, errors.New(
		"ERROR: fileStrategy could not find matching upstream")
}


// environStrategy reads proxy rules from environment variables
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
