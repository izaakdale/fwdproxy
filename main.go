package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	srv := &http.Server{
		Handler: timekeeper(mux),
		Addr:    fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")),
	}

	log.Fatal(srv.ListenAndServe())
}

func handleRequest(w http.ResponseWriter, inReq *http.Request) {
	outReq, err := http.NewRequest(inReq.Method, inReq.URL.String(), inReq.Body)
	if err != nil {
		http.Error(w, "proxy is failing to create requests", http.StatusInternalServerError)
		return
	}

	for key, values := range inReq.Header {
		for _, value := range values {
			outReq.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(outReq)
	if err != nil {
		http.Error(w, "failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, values := range inReq.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func timekeeper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()

		next.ServeHTTP(w, r)
		delta := time.Since(timeStart)

		log.Printf("%s %+v\n", r.Host, delta)
	})
}
