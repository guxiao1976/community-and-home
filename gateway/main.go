package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	// Identity service proxy
	identityURL, _ := url.Parse("http://localhost:8888")
	identityProxy := httputil.NewSingleHostReverseProxy(identityURL)

	// Masterdata service proxy
	masterdataURL, _ := url.Parse("http://localhost:8889")
	masterdataProxy := httputil.NewSingleHostReverseProxy(masterdataURL)

	// Main handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path
		if strings.HasPrefix(r.URL.Path, "/api/identity/") {
			identityProxy.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/masterdata/") {
			masterdataProxy.ServeHTTP(w, r)
		} else if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	log.Println("API Gateway starting on :8080")
	log.Println("Proxying /api/identity/* to http://localhost:8888")
	log.Println("Proxying /api/masterdata/* to http://localhost:8889")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
