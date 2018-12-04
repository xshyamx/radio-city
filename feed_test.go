package main

import (
	"io/ioutil"
	"testing"
)

func channelFromFile(file string) (Channel, error) {
	channel := Channel{}
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return channel, err
	}
	return getChannel(buf)
}

func feedFromFiles(files []string) (RSS, error) {
	rss := NewRSS()
	for _, file := range files {
		c, err := channelFromFile(file)
		if err == nil {
			rss.Channels = append(rss.Channels, c)
		}
	}
	return rss, nil
}

func TestFeed(t *testing.T) {
	rss, err := feedFromFiles([]string{
		"testdata/cd.html",
		"testdata/kck.html",
	})
	if err != nil {
		t.Fatalf("Failed to load feed from files")
	}
	if len(rss.Channels) != 2 {
		t.Fatalf("Loaded 2 files but 2 channels not created")
	}
}

func TestChannel(t *testing.T) {
	c, err := channelFromFile("testdata/cd.html")
	if err != nil {
		t.Fatalf("Failed to channel from file")
	}
	if c.Link == "" {
		t.Fatalf("link must be a full and valid URL")
	}
	if len(c.Items) == 0 {
		t.Fatalf("No items loaded")
	}
	item := c.Items[0]
	if item.Enclosure.Type != "audio/mpeg" {
		t.Fatalf("Enclosure has wrong type")
	}
}

func TestFeedXML(t *testing.T) {
	c, err := channelFromFile("testdata/cd.html")
	if err != nil {
		t.Fatalf("Failed to channel from file")
	}
	rss := NewRSS()
	rss.Channels = []Channel{c}
	buf, err := writeFeed(rss)
	if err != nil {
		t.Fatalf("Failed to write RSS feed\n%q\n", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("Failed to write XML")
	}
}
