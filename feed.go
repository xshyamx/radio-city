package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

type RSS struct {
	XMLName  xml.Name `xml:"rss"`
	Version  string   `xml:"version,attr"`
	Channels []Channel
}

func NewRSS() RSS {
	return RSS{
		Version: "2.0",
	}
}

type Channel struct {
	XMLName       xml.Name  `xml:"channel"`
	Title         string    `xml:"title"`
	Description   string    `xml:"description"`
	Link          string    `xml:"link"`
	LastBuildDate time.Time `xml:"lastBuildDate,omitempty"`
	PublishDate   time.Time `xml:"pubDate,omitempty"`
	Items         []Item
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	GUID        GUID
	Categories  []string  `xml:"category"`
	PublishDate time.Time `xml:"pubDate"`
}

type GUID struct {
	XMLName   xml.Name `xml:"guid"`
	Value     string   `xml:",chardata"`
	PermaLink bool     `xml:"PermaLink,attr"`
}

type XMLDate struct {
	time.Time
}

func (c *XMLDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(c.Time.Format(time.RFC1123Z), start)
}

func writeFeed(rss RSS) (*bytes.Buffer, error) {
	out := bytes.NewBufferString(xml.Header)
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")
	if err := enc.Encode(rss); err != nil {
		fmt.Printf("error: %v\n", err)
		return out, err
	}
	return out, nil
}
