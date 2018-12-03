package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

func feedFromUrls(urls []string) (RSS, error) {
	rss := NewRSS()
	for _, url := range urls {
		c, err := channelFromUrl(url)
		//fmt.Println("feedFromUrls", c, err)
		if err == nil {
			rss.Channels = append(rss.Channels, c)
		} else {
			fmt.Printf("Failed to load feed from %s\n", url)
		}
	}
	return rss, nil
}

func channelFromUrl(url string) (Channel, error) {
	channel := Channel{}
	buf, err := loadUrl(url)
	if err != nil {
		return channel, errors.Wrapf(err, "Failed GET %s", url)
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
	//fmt.Println("loadUrl", buf, err)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read response body")
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
	urlStr := doc.Find(`link[rel="canonical"]`).First().AttrOr("href", "")
	if strings.HasPrefix(urlStr, "//") {
		urlStr = "http:" + urlStr
	}
	channelLink, err := url.Parse(urlStr)
	if err != nil {
		return channel, errors.Wrapf(err, "Failed to parse url %s", urlStr)
	}
	channel.Link = channelLink.String()
	channel.PublishDate = XMLDate(time.Now())
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
		linkUrl, err := url.Parse(link)
		if err != nil {
			fmt.Printf("Failed to parse link %s", link)
		}
		item := Item{
			Title:       title,
			Description: desc,
			Link:        linkUrl.String(),
			Enclosure: Enclosure{
				URL:  linkUrl.String(),
				Type: mime.TypeByExtension(path.Ext(link)),
			},
			PublishDate: XMLDate(pd),
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
