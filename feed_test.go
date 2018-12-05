package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const configFile = "testdata/config.json"

var configMap = map[string]string{
	"/cd":  "testdata/cd.html",
	"/kck": "testdata/kck.html",
}

func feedFromFile(podcast Podcast, file string) (RSS, error) {
	rss := NewRSS()
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return rss, errors.Wrapf(err, "Failed to read file %s", file)
	}
	selfLink := NewAtomLink("http://localhost:8080" + podcast.Path)
	rss.Channel, err = getChannel(podcast, selfLink, buf)
	if err != nil {
		return rss, errors.Wrapf(err, "Failed to scrape channel from file")
	}
	return rss, nil
}

func loadPodcasts() ([]Podcast, error) {
	podcasts := []Podcast{}
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return podcasts, errors.Wrapf(err, "Failed read config file %s", configFile)
	}
	if err := json.Unmarshal(buf, &podcasts); err != nil {
		return podcasts, errors.Wrapf(err, "Failed to read ")
	}
	return podcasts, nil
}

func testBuilder(podcast Podcast) FeedBuilder {
	htmlFile, ok := configMap[podcast.Path]
	if !ok {
		fmt.Printf("Failed to find htmlFile for %s\n", podcast.Path)
	}
	return func(podcast Podcast, selfLink AtomLink) (RSS, error) {
		return feedFromFile(podcast, htmlFile)
	}
}

func validateSelfLink(selfLink AtomLink, t *testing.T) {
	if selfLink.URL == "" {
		t.Errorf("Channel must have a non-empty self link")
	}
	if selfLink.Rel != "self" {
		t.Errorf("Channel atom link rel must be 'self' instead was %s", selfLink.Rel)
	}
	if selfLink.Type != "application/rss+xml" {
		t.Errorf("Channel atom link rel must be 'application/rss+xml' instead was %s", selfLink.Type)
	}

}
func validateFeed(rss RSS, t *testing.T) {
	if rss.Version != "2.0" {
		t.Errorf("Expected to have version 2.0 but was %s", rss.Version)
	}
	channel := rss.Channel
	if channel.Title == "" {
		t.Errorf("Channel must have a non-emtpy title")
	}
	if channel.Link == "" {
		t.Errorf("Channel must have a non-empty link")
	}
	if channel.Image.URL == "" {
		t.Errorf("Channel image url must be a non-empty string")
	}
	validateSelfLink(channel.AtomLink, t)
	if len(channel.Items) == 0 {
		t.Errorf("Channel must have at least one item")
	}
	for _, item := range channel.Items {
		validateItem(item, t)
	}
}
func validateItem(item Item, t *testing.T) {
	if item.Title == "" {
		t.Errorf("Item must have a non-empty title")
	}
	if item.Link == "" {
		t.Errorf("Item must have a non-empty link")
	}
	if item.Enclosure.URL == "" {
		t.Errorf("Item enclosure must have a non-empty href")
	}
	if len(item.Categories) == 0 {
		t.Errorf("Item must have atleast one category")
	}
}

func printFeed(rss RSS) {
	buf, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal valid rss feed\n%q", err)
	}
	fmt.Println(string(buf))
}
func TestFeed(t *testing.T) {
	podcasts, err := loadPodcasts()
	if err != nil {
		t.Fatalf("Failed to load podcast config\n%q", err)
	}
	if len(podcasts) != 2 {
		t.Fatalf("Expected to load 2 podcasts but, loaded %d", len(podcasts))
	}
	podcast := podcasts[0]
	htmlFile, ok := configMap[podcast.Path]
	if !ok {
		t.Fatalf("Test file not found for path %s", podcast.Path)
	}
	rss, err := feedFromFile(podcast, htmlFile)
	if err != nil {
		t.Fatalf("Failed scrape podcast info\n%q", err)
	}
	validateFeed(rss, t)

}

func TestIndex(t *testing.T) {
	podcasts, err := loadPodcasts()
	if err != nil {
		t.Fatalf("Failed to load podcasts\n%q", err)
	}
	r := httptest.NewRequest("GET", "http://localhost:8080", nil)
	w := httptest.NewRecorder()
	IndexHandler(podcasts)(w, r)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Did not respond with success")
	}
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Expected 'text/html' mime type but got %s", contentType)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Fatalf("Failed to load queryable html document")
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		t.Run(podcasts[i].Name, func(t *testing.T) {
			if s.AttrOr("href", "") != podcasts[i].Path {
				t.Errorf("Expected href to be %s but was %s", podcasts[i].Path, s.AttrOr("href", ""))
			}
			if s.Text() != podcasts[i].Name {
				t.Errorf("Expected href to be %s but was %s", podcasts[i].Name, s.Text())
			}
		})
	})
}
func testPodcast(podcast Podcast, t *testing.T) func(*testing.T) {
	return func(t *testing.T) {
		url := "http://localhost:8080/" + podcast.Path
		r := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		RSSScrapeHandler(podcast, testBuilder(podcast))(w, r)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Did not respond with success")
		}
		contentType := res.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/rss+xml") {
			t.Errorf("Expected 'text/html' mime type but got %s", contentType)
		}
		body, _ := ioutil.ReadAll(res.Body)
		if len(body) == 0 {
			t.Fatalf("Response was empty")
		}
		// get linkUrl using string manipulation as go will fail to unmarshal namespaced elements
		xmlStr := string(body)
		start, end := strings.Index(xmlStr, "<link>"), strings.Index(xmlStr, "</link>")
		var linkUrl string
		if start < end && start != -1 {
			linkUrl = xmlStr[start+len("<link>") : end]
		}
		var rss RSS
		if err := xml.Unmarshal(body, &rss); err != nil {
			t.Fatalf("Failed to unmarshal feed xml\n%q", err)
		}
		// fix up the namespaced elments
		//    rss.AtomNS = NewRSS().AtomNS
		rss.Channel.AtomLink = NewAtomLink(url)
		rss.Channel.Link = linkUrl
		validateFeed(rss, t)
	}
}

func TestRSS(t *testing.T) {
	podcasts, err := loadPodcasts()
	if err != nil {
		t.Fatalf("Failed to load podcasts\n%q", err)
	}
	for _, podcast := range podcasts {
		t.Run(podcast.Name, testPodcast(podcast, t))
	}

}
