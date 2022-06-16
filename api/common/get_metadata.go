package common

import (
	"bufio"
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

var delim = []byte("</head>")

func getHTMLUptoHead(link string) (result []byte, err error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	streamedReader := bufio.NewReader(resp.Body)
	for {
		var found []byte
		found, err = streamedReader.ReadBytes(delim[len(delim)-1])
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return
		}
		result = append(result, found...)
		if bytes.HasSuffix(result, delim) {
			return result[:len(result)-len(delim)], nil
		}
	}
}

func GetMetadata(link string) (*DB.Meta, error) {
	data, err := getHTMLUptoHead(link)
	if err != nil {
		return nil, err
	}
	headNode, _ := html.Parse(bytes.NewReader(data))

	hm := &DB.Meta{}
	crawl(headNode, hm)
	return hm, nil
}
