package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/atotto/clipboard"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func ListenForHotKey() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key(0x2f))
	err := hk.Register()
	if err != nil {
		log.Fatal("Couldn't register hotkey")
	}

	for i := range hk.Keydown() {
		fmt.Println(i)
		url, err := clipboard.ReadAll()
		handle(err)
		http.Post("http://localhost:8080/add", "text/plain", strings.NewReader(url))
	}
}

func main() {
	mainthread.Init(ListenForHotKey)
}
