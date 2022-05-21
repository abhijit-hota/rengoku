package common

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"

	DB "github.com/abhijit-hota/rengoku/server/db"

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
	foundValue := ""
	for _, attr := range attributes {
		if (attr.Key == key || (lookingFor != "icon" && attr.Key == "name")) && strings.Contains(attr.Val, lookingFor) {
			if foundValue != "" {
				return foundValue
			}
			foundProp = true
		}
		if attr.Key == content {
			if foundProp {
				return attr.Val
			}
			foundValue = attr.Val
		}
	}
	return ""
}

var headRegex = regexp.MustCompile("<head.*>((.|\n|\r\n)+)</head>")

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
func GetMetadata(link string) (*DB.Meta, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}

	data, _ := io.ReadAll(resp.Body)
	head := headRegex.Find(data)
	headNode, _ := html.Parse(bytes.NewReader(head))

	hm := &DB.Meta{}
	crawl(headNode, hm)

	return hm, nil
}
