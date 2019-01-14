package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
)

func buildFeed(podcast Podcast, selfLink AtomLink) (RSS, error) {
	rss := NewRSS()
	var err error
	rss.Channel, err = scrapeChannel(podcast, selfLink)
	return rss, err
}

func buildMasterFeed(podcasts []Podcast, selfLink AtomLink) (RSS, error) {
	masterImage := "https://www.radiocity.in/images/menu-images/logo.png"
	imgUrl, err := parseURL(masterImage)
	if err != nil {
		fmt.Printf("Failed to parse image url %s", masterImage)
	}
	rss := NewRSS()

	rss.Channel = Channel{
		AtomLink:      selfLink,
		Title:         "RadioCity Master Feed",
		Link:          selfLink.URL,
		PublishDate:   XMLDate(time.Now()),
		LastBuildDate: XMLDate(time.Now()),
		Description:   "Generated master feed from a given set of podcasts",
		Image: Image{
			Link:  imgUrl,
			Title: "RadioCity Master Feed",
			URL:   selfLink.URL,
		},
		ItunesImage: ItunesImage{
			URL: imgUrl,
		},
	}
	for _, pod := range podcasts {
		pitems, err := scrapeItems(pod)
		if err != nil {
			return rss, err
		}
		rss.Channel.Items = append(rss.Channel.Items, pitems...)
	}
	return rss, err
}

func MasterScrapeHandler(podcasts []Podcast, builder MasterFeedBuilder) http.HandlerFunc {
	var rss RSS
	var lastBuildDate XMLDate
	return func(w http.ResponseWriter, r *http.Request) {
		reload := r.URL.Query().Get("refresh") == "true"
		//fmt.Println("reload", reload)
		if reload || rss.Version == "" {
			//fmt.Println("reloading...")
			if !time.Time(rss.Channel.PublishDate).IsZero() {
				lastBuildDate = rss.Channel.PublishDate
			}
			proto := r.Header.Get("X-Forwarded-Proto")
			if proto == "" {
				proto = "http"
			}
			selfLink := NewAtomLink(fmt.Sprintf("%s://%s%s", proto, r.Host, r.URL.Path))
			var err error
			rss, err = builder(podcasts, selfLink)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to build RSS feed"))
				return
			}
			if !time.Time(lastBuildDate).IsZero() {
				rss.Channel.LastBuildDate = lastBuildDate
			}
			// expire value after 1 day
			timer := time.NewTimer(24 * time.Hour)
			go func() {
				<-timer.C
				rss.Version = ""
			}()

		}
		w.Header().Set("Content-Type", "application/rss+xml")
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		if err := enc.Encode(rss); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}
}

func RSSScrapeHandler(podcast Podcast, builder FeedBuilder) http.HandlerFunc {
	var rss RSS
	var lastBuildDate XMLDate
	return func(w http.ResponseWriter, r *http.Request) {
		reload := r.URL.Query().Get("refresh") == "true"
		//fmt.Println("reload", reload)
		if reload || rss.Version == "" {
			//fmt.Println("reloading...")
			if !time.Time(rss.Channel.PublishDate).IsZero() {
				lastBuildDate = rss.Channel.PublishDate
			}
			proto := r.Header.Get("X-Forwarded-Proto")
			if proto == "" {
				proto = "http"
			}
			selfLink := NewAtomLink(fmt.Sprintf("%s://%s%s", proto, r.Host, r.URL.Path))
			var err error
			rss, err = builder(podcast, selfLink)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to build RSS feed"))
				return
			}
			if !time.Time(lastBuildDate).IsZero() {
				rss.Channel.LastBuildDate = lastBuildDate
			}
			// expire value after 1 day
			timer := time.NewTimer(24 * time.Hour)
			go func() {
				<-timer.C
				rss.Version = ""
			}()

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
      <li><a href="/master">Master Feed</a></li>
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
	Path       string   `json:"prefix"`
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	Image      string   `json:"imageUrl"`
	Categories []string `json:"categories"`
}

func getPodcasts() ([]Podcast, error) {
	configUrl := os.Getenv("CONFIG_URL")
	if configUrl == "" {
		configUrl = "https://api.myjson.com/bins/mn5mu"
	}
	var podcasts []Podcast
	buf, err := loadUrl(configUrl)
	if err != nil {
		return podcasts, errors.Wrapf(err, "Failed to retrieve url %s", configUrl)
	}
	if err := json.Unmarshal(buf, &podcasts); err != nil {
		return podcasts, errors.Wrapf(err, "Failed to parse %s as json", string(buf))
	}
	return podcasts, nil
}

func main() {
	logger := log.New(os.Stdout, "", 0)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8080"
	}
	podcasts, err := getPodcasts()
	if err != nil {
		logger.Fatalf("Failed to retrieve any podcast configurations\n%q", err)
	}
	logger.Printf("Loaded %d podcast configs\n", len(podcasts))

	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler(podcasts))
	for _, podcast := range podcasts {
		mux.HandleFunc(podcast.Path, RSSScrapeHandler(podcast, buildFeed))
	}
	mux.HandleFunc("/master", MasterScrapeHandler(podcasts, buildMasterFeed))
	h := &http.Server{Addr: addr, Handler: mux}

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
