package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
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

type Channel struct {
	XMLName       xml.Name `xml:"channel"`
	Title         string   `xml:"title"`
	Description   string   `xml:"description"`
	Link          URL      `xml:"link"`
	AtomLink      AtomLink
	Image         Image
	LastBuildDate XMLDate `xml:"lastBuildDate,omitempty"`
	PublishDate   XMLDate `xml:"pubDate,omitempty"`
	Items         []Item  `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        URL    `xml:"link"`
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
	Type    string   `xml:"type,attr"`
	URL     URL      `xml:"url,attr"`
	Length  int      `xml:"length,attr"`
}

type Image struct {
	XMLName xml.Name `xml:"image"`
	Link    URL      `xml:"link"`
	URL     URL      `xml:"url"`
	Title   string   `xml:"title"`
}

type AtomLink struct {
	XMLName xml.Name `xml:"atom:link"`
	URL     URL      `xml:"href,attr"`
	Rel     string   `xml:"rel,attr"`
	Type    string   `xml:"type,attr"`
}

func parseURL(link string) (URL, error) {
	var ur URL
	u, err := url.Parse(link)
	if err != nil {
		return ur, err
	}
	ur = URL(*u)
	return ur, nil
}
func NewAtomLink(link string) AtomLink {
	u, err := parseURL(link)
	if err != nil {
		fmt.Println("Failed to parse", link, err)
	}
	return AtomLink{
		URL:  u,
		Rel:  "self",
		Type: "application/rss+xml",
	}
}

type XMLDate time.Time
type URL url.URL

func (u URL) String() string {
	ux := url.URL(u)
	up := &ux
	return up.String()
}

func (u URL) RequestURI() string {
	ux := url.URL(u)
	//  up := &ux
	return ux.RequestURI()
}

func (u URL) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	un := url.URL(u)
	up := &un
	return e.EncodeElement(up.String(), start)
}

func (u *URL) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var urlStr string
	if err := d.DecodeElement(&urlStr, &start); err != nil {
		return err
	}
	ux, err := parseURL(urlStr)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse %s as url", urlStr)
	}
	u.Scheme = ux.Scheme
	u.Opaque = ux.Opaque
	u.User = ux.User
	u.Host = ux.Host
	u.Path = ux.Path
	u.RawPath = ux.RawPath
	u.ForceQuery = ux.ForceQuery
	u.RawQuery = ux.RawQuery
	u.Fragment = ux.Fragment
	return nil
}

func (u URL) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	un := url.URL(u)
	up := &un
	return xml.Attr{
		Name:  name,
		Value: up.String(),
	}, nil
}
func (u *URL) UnmarshalXMLAttr(attr xml.Attr) error {
	ux, err := parseURL(attr.Value)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse %s as url", attr.Value)
	}
	u.Scheme = ux.Scheme
	u.Opaque = ux.Opaque
	u.User = ux.User
	u.Host = ux.Host
	u.Path = ux.Path
	u.RawPath = ux.RawPath
	u.ForceQuery = ux.ForceQuery
	u.RawQuery = ux.RawQuery
	u.Fragment = ux.Fragment
	return nil
}

func (d XMLDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(time.Time(d).Format(time.RFC1123Z), start)
}

func (dd *XMLDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var xmlStr string
	if err := d.DecodeElement(&xmlStr, &start); err != nil {
		return err
	}
	dt, err := time.Parse(time.RFC1123Z, xmlStr)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse %s as RFC1123Z date", xmlStr)
	}
	*dd = XMLDate(dt)
	return nil
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
