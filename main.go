package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func RSSScrapeHandler(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rss, err := feedFromUrls([]string{
			url,
		})
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
}

func IndexHandler(podcasts []Podcast) http.HandlerFunc {
	tmpl, err := template.New("index.html").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8"/>
    <title>RadioCity Podcasts</title>
  </head>
  <body>
    <h1>RadioCity Podcasts</h1>
    <ul>
      {{ range . }}
      <li><a href="{{.Path}}">{{.Name}}</a></li>
      {{ end }}
    </ul>
  </body>
</html>`)
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, podcasts); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to process template"))
		}
	}
}

type Podcast struct {
	Path string
	Name string
	URL  string
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8080"
	}
	podcasts := []Podcast{
		Podcast{"/cd", "Crime Diary", "https://www.radiocity.in/radiocity/show-podcasts-tamil/Crime-Diary/153"},
		Podcast{"/kck", "Kissa Crime Ka", "https://www.radiocity.in/radiocity/show-podcasts-hindi/Kissa-Crime-Ka/82"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler(podcasts))
	for _, podcast := range podcasts {
		mux.HandleFunc(podcast.Path, RSSScrapeHandler(podcast.URL))
	}
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
