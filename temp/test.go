package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func main() {

	http.HandleFunc(`/read`, func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set(`Content-Type`, `application/json`)

		err := req.ParseForm()
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(rw).Encode(map[string]string{"error": err.Error()})
			return
		}

		link := req.FormValue(`link`)
		if link == "" {
			rw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(rw).Encode(map[string]string{"error": `empty value of link`})
			return
		}

		if _, err := url.Parse(link); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(rw).Encode(map[string]string{"error": err.Error()})
			return
		}

		resp, err := http.Get(link)
		if err != nil {
			//proxy status and err
			rw.WriteHeader(resp.StatusCode)
			json.NewEncoder(rw).Encode(map[string]string{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		meta := extract(resp.Body)
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(meta)
		return
	})

	// little help %)
	println("call like: \n$ curl -XPOST 'http://localhost:4567/read' -d link='https://github.com/golang/go'")

	err := http.ListenAndServe(`:4567`, nil)
	if err != nil {
		panic(err)
	}

}

type HTMLMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

func attrValue(attributes []html.Attribute, lookingFor string) string {
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

func crawler(node *html.Node, hm *HTMLMeta) {
	if node.Type == html.TextNode && node.Parent.Data == "title" {
		hm.Title = node.Data
	}
	if node.Type == html.ElementNode {
		if hm.Title == "" && node.Data == "meta" {
			hm.Title = attrValue(node.Attr, "title")
		}
		if hm.Description == "" && node.Data == "meta" {
			hm.Description = attrValue(node.Attr, "description")
		}
		if len(hm.Favicon) == 0 && node.Data == "link" {
			hm.Favicon = attrValue(node.Attr, "icon")
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		crawler(child, hm)
	}
}
func extract(resp io.Reader) *HTMLMeta {
	headRegex := regexp.MustCompile("<head>((.|\n|\r\n)+)</head>")
	hm := new(HTMLMeta)

	data, _ := io.ReadAll(resp)
	head := headRegex.Find(data)
	tokens, _ := html.Parse(bytes.NewReader(head))

	crawler(tokens, hm)

	return hm
}
