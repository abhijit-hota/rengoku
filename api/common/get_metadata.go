package common

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	DB "github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"

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
	contentType := resp.Header["Content-Type"]
	if !strings.HasPrefix(resp.Status, "2") || len(contentType) == 0 {
		return nil, errors.New("can't fetch")
	}
	if !strings.HasPrefix(contentType[0], "text/html") {
		return nil, errors.New("not html")
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
	if len(data) <= 0 {
		return nil, errors.New("meta not found")
	}
	headNode := utils.MustGet(html.Parse(bytes.NewReader(data)))

	hm := &DB.Meta{}
	crawl(headNode, hm)

	hm.Title = strings.TrimSpace(strings.ReplaceAll(hm.Title, "\n", ""))
	hm.Description = strings.TrimSpace(hm.Description)

	return hm, nil
}
