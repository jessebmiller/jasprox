package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/http/httptest"
	"net/url"
)


func jasHandler(w http.ResponseWriter, r *http.Request) {
	echoRequestServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, r)
			}))
	defer echoRequestServer.Close()

	echoHostServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, r.Host)
			}))
	defer echoHostServer.Close()

	requestURL, err := url.Parse(echoRequestServer.URL)
	if err != nil {
		log.Fatal(err)
	}
	hostURL, err := url.Parse(echoHostServer.URL)
	if err != nil {
		log.Fatal(err)
	}

	requestProxy := httputil.NewSingleHostReverseProxy(requestURL)
	hostProxy := httputil.NewSingleHostReverseProxy(hostURL)

	fmt.Println(r.URL)
	if r.URL.Path == "/host" {
		hostProxy.ServeHTTP(w, r)
	} else {
		requestProxy.ServeHTTP(w, r)
	}
}


func main() {
	fmt.Println("This is Jasprox.")

	log.Fatal(http.ListenAndServe(":8000", http.HandlerFunc(jasHandler)))

	fmt.Println("Goodbye.")
}
