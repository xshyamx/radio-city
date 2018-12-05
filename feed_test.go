package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"testing"

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
