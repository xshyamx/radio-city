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
		t.Errorf("Failed to load feed from files")
	}
	if len(rss.Channels) != 2 {
		t.Errorf("Loaded 2 files but 2 channels not created")
	}
}

func TestChannel(t *testing.T) {
	c, err := channelFromFile("testdata/cd.html")
	if err != nil {
		t.Errorf("Failed to channel from file")
	}
	if len(c.Items) == 0 {
		t.Errorf("No items loaded")
	}
}
