package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// extractItems extracts a list of items from a parsed document
func extractItems(doc *goquery.Document, imgUrl URL, categories []string) ([]Item, error) {
	var items []Item
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
		length := 0
		linkUrl, err := parseURL(link)
		if err != nil {
			fmt.Printf("Failed to parse link %s", link)
		}
		if link != "" {
			res, err := http.Head(linkUrl.String())
			if err != nil {
				fmt.Printf("Failed to load media url %s\n%q\n", link, err)
			}
			if res.StatusCode == http.StatusOK {
				ls := res.Header.Get("Content-Length")
				length, _ = strconv.Atoi(ls)
			}
		}

		item := Item{
			Title:       title,
			Description: desc,
			Link:        linkUrl,
			Enclosure: Enclosure{
				URL:    linkUrl,
				Type:   mime.TypeByExtension(path.Ext(link)),
				Length: length,
			},
			ItunesImage: ItunesImage{URL: imgUrl},
			PublishDate: XMLDate(pd),
			GUID: GUID{
				Value: link,
			},
			Categories: categories,
		}
		if item.Description != "" && item.Link.RequestURI() != "" {
			items = append(items, item)
		}
	})
	return items, nil
}

// getChannel builds a channel from scraped podcast url buffer
func getChannel(podcast Podcast, selfLink AtomLink, buf []byte) (Channel, error) {
	channel := Channel{}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(buf))
	if err != nil {
		return channel, err
	}
	channel.Title = doc.Find(".pod_desc_txt h1").First().Text()
	channel.Description = doc.Find(`.pod_desc_txt p`).First().Text()
	urlStr := doc.Find(`link[rel="canonical"]`).First().AttrOr("href", "")
	if strings.HasPrefix(urlStr, "//") {
		urlStr = "http:" + urlStr
	}
	channelLink, err := url.Parse(urlStr)
	if err != nil {
		return channel, errors.Wrapf(err, "Failed to parse url %s", urlStr)
	}
	channel.Link = URL(*channelLink)
	channel.AtomLink = selfLink
	channel.LastBuildDate = XMLDate(time.Now())
	channel.PublishDate = XMLDate(time.Now())

	if imgUrl, ok := doc.Find(".pod_desc_img img").First().Attr("src"); ok {
		img, err := url.Parse(imgUrl)
		if err != nil {
			return channel, errors.Wrapf(err, "Failed to parse image url %s", imgUrl)
		}
		channel.Image = Image{
			Title: channel.Title,
			Link:  channel.Link,
			URL:   URL(*img),
		}
		channel.ItunesImage = ItunesImage{URL: channel.Image.URL}
	}
	if channel.Items, err = extractItems(doc, channel.Image.URL, podcast.Categories); err != nil {
		return channel, err
	}
	return channel, nil
}

// getItems returns a list of items from the given podcast from a buffer
func getItems(podcast Podcast, buf []byte) ([]Item, error) {
	var items []Item
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(buf))
	if err != nil {
		return items, err
	}
	imgUrl, err := parseURL(podcast.Image)
	if err != nil {
		fmt.Printf("Failed to parse image url %s", podcast.Image)
	}
	return extractItems(doc, imgUrl, podcast.Categories)
}

// scrapeChannel builds a new channel with the items scraped from the podcast
func scrapeChannel(podcast Podcast, selfLink AtomLink) (Channel, error) {
	channel := Channel{}
	buf, err := loadUrl(podcast.URL)
	if err != nil {
		return channel, errors.Wrap(err, "Failed to load podcast url")
	}
	return getChannel(podcast, selfLink, buf)
}

// scrapeItem builds a list of items by scraping the podcast url
func scrapeItems(podcast Podcast) ([]Item, error) {
	buf, err := loadUrl(podcast.URL)
	if err != nil {
		return []Item{}, errors.Wrap(err, "Failed to load podcast url")
	}
	return getItems(podcast, buf)
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

type FeedBuilder func(Podcast, AtomLink) (RSS, error)
type MasterFeedBuilder func([]Podcast, AtomLink) (RSS, error)
