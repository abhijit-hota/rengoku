package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"io"
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

func extract(resp io.Reader) *HTMLMeta {
	hm := new(HTMLMeta)
	data, _ := io.ReadAll(resp)
	headRegex := regexp.MustCompile("<head>((.|\n|\r\n)+)</head>")

	head := string(headRegex.Find(data))
	head = strings.ReplaceAll(head, "\n", "")

	titleRegex := regexp.MustCompile(`<title.*>(.+)<\/title>`)
	metaTitleRegex := regexp.MustCompile(`<meta.*?property="og:title".*?content="(.+?)".*?\/?>`)
	descriptionRegex := regexp.MustCompile(`<meta.*?(?:name="description"|property="og:description").*?content="(.*?)".*?\/>`)

	descMatches := descriptionRegex.FindStringSubmatch(head)
	titleMatches := titleRegex.FindStringSubmatch(head)
	if len(titleMatches) == 0 {
		titleMatches = metaTitleRegex.FindStringSubmatch(head)
	}

	if len(descMatches) == 0 {
		hm.Description = ""
	} else {
		hm.Description = descMatches[1]
	}

	if len(titleMatches) == 0 {
		hm.Title = ""
	} else {
		hm.Title = titleMatches[1]
	}

	return hm
}
