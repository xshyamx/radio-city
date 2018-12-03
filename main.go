package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func serveRSS(w http.ResponseWriter, r *http.Request) {

	urls := []string{
		"https://www.radiocity.in/radiocity/show-podcasts-tamil/Crime-Diary/153",
		"https://www.radiocity.in/radiocity/show-podcasts-hindi/Kissa-Crime-Ka/82",
	}

	rss, err := feedFromUrls(urls)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to scrape feed data"))
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(rss); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func setupShutdownHandler(srv *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				if err := srv.Shutdown(nil); err != nil {
					panic(err) // failure/timeout shutting down the server gracefully
				}
				return
			}
		}
	}()
}

func main() {
	log.Printf("main: starting HTTP server")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port}

	http.HandleFunc("/", serveRSS)
	go setupShutdownHandler(srv)
	fmt.Printf("Server listening on :%s\n", port)
	if err := srv.ListenAndServe(); err != nil {
		// cannot panic, because this probably is an intentional close
		log.Printf("Httpserver: ListenAndServe() error: %s", err)
	}

}
