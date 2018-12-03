package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveRSS)
	h := &http.Server{Addr: addr, Handler: mux}
	logger := log.New(os.Stdout, "", 0)

	go func() {
		logger.Printf("Listening on http://0.0.0.0%s\n", addr)

		if err := h.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	<-stop

	logger.Println("\nShutting down the server...")
	h.Shutdown(context.Background())
	logger.Println("Server gracefully stopped")
}
