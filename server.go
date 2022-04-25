package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/atotto/clipboard"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func dataserver() {
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello")
		})

	log.Println("Starting data server")
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

func fn() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key(0x2f))
	err := hk.Register()
	if err != nil {
		return
	}

	for range hk.Keydown() {
		text, err := clipboard.ReadAll()
		if err != nil {
			panic(err)
		}
		fmt.Println(text)
	}
}

func main() {
	go dataserver()
	mainthread.Init(fn)
}
