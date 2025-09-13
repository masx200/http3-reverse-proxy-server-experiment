package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world\n")
		log.Println("Request:", r.Method, r.URL.Path, r.Proto)
		fmt.Fprintf(w, "Protocol: %s\n", r.Proto)
	})
	h2s := &http2.Server{
		// ...
	}
	h1s := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(handler, h2s),
	}
	log.Println("http server Listening on :8080")
	log.Fatal(h1s.ListenAndServe())

}
