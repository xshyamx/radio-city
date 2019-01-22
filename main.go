package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Podcast struct {
	Path       string   `json:"prefix"`
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	Image      string   `json:"imageUrl"`
	Categories []string `json:"categories"`
}

var podcasts = []Podcast{
	Podcast{
		Path:       "/cd",
		Name:       "Crime Diary",
		URL:        "https://www.radiocity.in/radiocity/show-podcasts-tamil/Crime-Diary/153",
		Image:      "https://www.radiocity.in//images/other-channels/other-podcast/CrimeDiary Podcast40kb1493819764.jpg",
		Categories: []string{"crime", "podcast"},
	},
	Podcast{
		Path:       "/kck",
		Name:       "Kissa Crime Ka",
		URL:        "https://www.radiocity.in/radiocity/show-podcasts-hindi/Kissa-Crime-Ka/82",
		Image:      "https://www.radiocity.in//images/other-channels/other-podcast/kisacrimeka1490279213.jpg",
		Categories: []string{"crime", "podcast"},
	},
}

func buildFeed(podcasts []Podcast, selfLink AtomLink) (RSS, error) {
	logger := log.New(os.Stderr, "[main][buildFeed ", 0)
	start := time.Now()
	rss := NewRSS()
	masterImage := "https://www.radiocity.in/images/menu-images/logo.png"
	imgUrl, err := parseURL(masterImage)
	if err != nil {
		fmt.Printf("Failed to parse image url %s", masterImage)
	}
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
	for _, podcast := range podcasts {
		pitems, err := scrapeItems(podcast)
		if err != nil {
			return rss, err
		}
		rss.Channel.Items = append(rss.Channel.Items, pitems...)
	}
	logger.Printf("Built master feed in %s", time.Since(start).String())
	return rss, nil

}
func main() {

	rss, err := buildFeed(podcasts, NewAtomLink("http://localhost:8080/master"))
	if err != nil {
		panic(err)
	}
	out, err := writeFeed(rss)
	if err != nil {
		panic(err)
	}
	fmt.Println(out.String())
}
