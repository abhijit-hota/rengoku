package common

import (
	DB "api/db"
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func searchAttributes(attributes []html.Attribute, lookingFor string) string {
	var key string

	var content string
	if lookingFor == "icon" {
		key = "rel"
		content = "href"
	} else {
		key = "property"
		content = "content"
	}

	foundProp := false
	for _, attr := range attributes {
		if (attr.Key == key || (lookingFor != "icon" && attr.Key == "name")) && strings.Contains(attr.Val, lookingFor) {
			foundProp = true
		}
		if foundProp && attr.Key == content {
			return attr.Val
		}
	}
	return ""
}

var headRegex *regexp.Regexp

func init() {
	headRegex = regexp.MustCompile("<head>((.|\n|\r\n)+)</head>")
}

func crawl(node *html.Node, hm *DB.Meta) {
	if node.Type == html.TextNode && node.Parent.Data == "title" {
		hm.Title = node.Data
	}
	if node.Type == html.ElementNode {
		if hm.Title == "" && node.Data == "meta" {
			hm.Title = searchAttributes(node.Attr, "title")
		}
		if hm.Description == "" && node.Data == "meta" {
			hm.Description = searchAttributes(node.Attr, "description")
		}
		if hm.Favicon == "" && node.Data == "link" {
			hm.Favicon = searchAttributes(node.Attr, "icon")
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		crawl(child, hm)
	}
}
func GetMetadata(link string, hm *DB.Meta) error {
	link = strings.TrimSpace(link)
	if !(strings.HasPrefix(link, "https://") || strings.HasPrefix(link, "https://")) {
		link = "https://" + link
	}

	if _, err := url.Parse(link); err != nil {
		return err
	}

	resp, err := http.Get(link)
	if err != nil {
		return err
	}

	data, _ := io.ReadAll(resp.Body)
	head := headRegex.Find(data)
	headNode, _ := html.Parse(bytes.NewReader(head))

	crawl(headNode, hm)

	return nil
}
