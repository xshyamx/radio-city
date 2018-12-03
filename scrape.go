package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func feedFromUrls(urls []string) (RSS, error) {
	rss := NewRSS()
	for _, url := range urls {
		c, err := channelFromUrl(url)
		if err == nil {
			rss.Channels = append(rss.Channels, c)
		}
	}
	return rss, nil
}

func channelFromUrl(url string) (Channel, error) {
	channel := Channel{}
	buf, err := loadUrl(url)
	if err != nil {
		return channel, err
	}
	return getChannel(buf)
}

func loadUrl(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%s returned status code %d", url, res.StatusCode)
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body")
	}
	return buf, nil
}

func getChannel(buf []byte) (Channel, error) {
	channel := Channel{}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(buf))
	if err != nil {
		return channel, err
	}

	channel.Title = doc.Find("title").First().Text()
	channel.Description = doc.Find(`.pod_desc_txt p`).First().Text()
	channel.Link = doc.Find(`link[rel="canonical"]`).First().AttrOr("href", "")
	channel.PublishDate = time.Now()
	IST, _ := time.LoadLocation("Asia/Kolkata")
	doc.Find(".podcast_button a").Each(func(i int, pi *goquery.Selection) {
		descStr := pi.AttrOr("data-podname", "")
		link := strings.TrimSpace(pi.AttrOr("data-podcast", ""))
		pd := time.Now()
		title, desc := descStr, descStr
		if descStr != "" {
			di := strings.LastIndex(descStr, "-")
			if di != -1 {
				dateStr := strings.TrimSpace(descStr[di+1:])
				pd, _ = time.ParseInLocation("January 2, 2006", dateStr, IST)
				if pd.IsZero() {
					fmt.Println("Failed to parse", dateStr)
				}
			}
			fi := strings.Index(descStr, "-")
			if fi != -1 {
				title = strings.TrimSpace(descStr[0:fi])
				if fi < di {
					desc = strings.TrimSpace(descStr[fi+1 : di])
				} else {
					desc = title
				}
			}
		}
		item := Item{
			Title:       title,
			Description: desc,
			Link:        link,
			PublishDate: pd,
			GUID: GUID{
				Value: link,
			},
			Categories: []string{
				"podcast",
				"crime",
			},
		}
		if item.Description != "" && item.Link != "" {
			channel.Items = append(channel.Items, item)
		}
	})
	return channel, nil
}
