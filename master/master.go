package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const BASE_URL = "https://www.radiocity.in"

type Category struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Podcast struct {
	Path  string `json:"prefix"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Image string `json:"imageUrl"`
}

func scrapeChannelImage(in <-chan Podcast, out chan<- Podcast, errs chan<- error, done chan<- bool) {
	for pod := range in {
		res, err := http.Get(pod.URL)
		if err != nil {
			errs <- errors.Wrapf(err, "Failed to load podcast detail page from %s", pod.URL)
		}
		if res.StatusCode != http.StatusOK {
			errs <- fmt.Errorf("Expected status code %d got %d", http.StatusOK, res.StatusCode)
		}
		defer res.Body.Close()
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if imgUrl, ok := doc.Find(".pod_desc_img img").First().Attr("src"); ok {
			pod.Image = imgUrl
		}
		out <- pod
	}
	done <- true
}
func scrapeCategory(cats <-chan Category, pods chan<- Podcast, errs chan<- error, done chan<- bool) {
	for category := range cats {
		res, err := http.Get(category.URL)
		if err != nil {
			errs <- errors.Wrapf(err, "Failed to load %s", category.URL)
		}
		if res.StatusCode != http.StatusOK {
			errs <- fmt.Errorf("Expected status code %d got %d", http.StatusOK, res.StatusCode)
		}
		defer res.Body.Close()
		doc, err := goquery.NewDocumentFromReader(res.Body)
		/*
		   file := "testdata/tamil.html"
		   buf, err := ioutil.ReadFile(file)
		   if err != nil {
		     errs <- errors.Wrapf(err, "Failed to read from %s", file)
		   }
		   doc, err := goquery.NewDocumentFromReader(bytes.NewReader(buf))
		   if err != nil {
		     errs <- errors.Wrapf(err, "Failed to parse html file")
		   }
		*/

		if err != nil {
			errs <- errors.Wrap(err, "Failed to parse response html")
		}
		doc.Find(".podcast_button").Each(func(i int, s *goquery.Selection) {
			pods <- Podcast{
				Name: s.Find("p").Text(),
				URL:  s.Find("a").First().AttrOr("href", ""),
				Path: makePath(s.Find("p").Text()),
			}
		})
	}
	done <- true
}

func makePath(name string) string {
	var b strings.Builder
	b.WriteString("/")
	for _, c := range strings.Split(strings.TrimSpace(name), " ") {
		b.WriteByte(c[0])
	}
	return strings.ToLower(b.String())
}

func getLandingPages(reader io.Reader, cats chan<- Category, pods chan<- Podcast, errs chan<- error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		errs <- errors.Wrap(err, "Failed to parse response html")
	}
	li := doc.Find(`li a:matchesOwn(^Podcast$)`).Parent()
	if li.Length() == 0 {
		errs <- fmt.Errorf("Failed to match root element")
	}
	//fmt.Println(li.Html())
	li.Find("a").Each(func(i int, s *goquery.Selection) {
		if strings.HasPrefix(s.AttrOr("href", ""), "http") {
			link, _ := url.Parse(s.AttrOr("href", ""))
			lastFrag := link.Path[strings.LastIndex(link.Path, "/")+1:]
			if _, err := strconv.Atoi(lastFrag); err != nil {
				cats <- Category{s.Text(), link.String()}
			} else {
				pods <- Podcast{
					Name: s.Text(),
					URL:  link.String(),
					Path: makePath(s.Text()),
				}
			}
		}
	})
	close(cats)
}

func getLandingPageFromUrl(baseUrl string, cats chan<- Category, pods chan<- Podcast, errs chan<- error) {
	res, err := http.Get(baseUrl)
	if err != nil {
		errs <- errors.Wrapf(err, "Failed to load %s", baseUrl)
	}
	if res.StatusCode != http.StatusOK {
		errs <- fmt.Errorf("Expected status code %d got %d", http.StatusOK, res.StatusCode)
	}
	defer res.Body.Close()
	getLandingPages(res.Body, cats, pods, errs)
}

func main() {
	cats := make(chan Category)
	pods := make(chan Podcast)
	ipod := make(chan Podcast)
	errs := make(chan error)
	done := make(chan bool)
	idone := make(chan bool)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	go getLandingPageFromUrl(BASE_URL, cats, pods, errs)
	// spawn category workers
	doneCount, idoneCount, workerCount := 0, 0, 3
	for i := 0; i < workerCount; i++ {
		go scrapeCategory(cats, pods, errs, done)
		go scrapeChannelImage(pods, ipod, errs, idone)
	}
	podMap := make(map[string]Podcast)
	for {
		select {
		case podcast, ok := <-ipod:
			if ok {
				podMap[podcast.URL] = podcast
			} else {
				//completed
				podcasts := make([]Podcast, len(podMap))
				i := 0
				for _, pod := range podMap {
					podcasts[i] = pod
					i++
				}
				enc.Encode(podcasts)
				return
			}
			//enc.Encode(podcast)
		case <-idone:
			idoneCount++
			if idoneCount >= workerCount {
				close(ipod)
			}
		case <-done:
			doneCount++
			if doneCount >= workerCount {
				close(pods)
			}
		case err := <-errs:
			fmt.Println(err)
			return
		}
	}
}
