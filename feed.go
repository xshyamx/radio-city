package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	AtomNS  string   `xml:"xmlns:atom,attr"`
	Channel Channel
}

func NewRSS() RSS {
	return RSS{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
	}
}

//https://www.radiocity.in/images/menu-images/logo.png
type Channel struct {
	XMLName       xml.Name `xml:"channel"`
	Title         string   `xml:"title"`
	Description   string   `xml:"description"`
	Link          string   `xml:"link"`
	AtomLink      AtomLink
	Image         Image
	LastBuildDate XMLDate `xml:"lastBuildDate,omitempty"`
	PublishDate   XMLDate `xml:"pubDate,omitempty"`
	Items         []Item
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	GUID        GUID
	Enclosure   Enclosure
	Categories  []string `xml:"category"`
	PublishDate XMLDate  `xml:"pubDate"`
}

type GUID struct {
	XMLName   xml.Name `xml:"guid"`
	Value     string   `xml:",chardata"`
	PermaLink bool     `xml:"isPermaLink,attr"`
}

type Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Type    string   `xml:"type,attr"`
	Length  int      `xml:"length,attr"`
}

type Image struct {
	XMLName xml.Name `xml:"image"`
	Link    string   `xml:"link"`
	URL     string   `xml:"url"`
	Title   string   `xml:"title"`
}

type AtomLink struct {
	XMLName xml.Name `xml:"atom:link"`
	URL     string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr"`
	Type    string   `xml:"type,attr"`
}

func NewAtomLink(link string) AtomLink {
	return AtomLink{
		URL:  link,
		Rel:  "self",
		Type: "application/rss+xml",
	}
}

type XMLDate time.Time

func (d XMLDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(time.Time(d).Format(time.RFC1123Z), start)
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
