package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	KL "github.com/MarinX/keylogger"
	CB "github.com/atotto/clipboard"
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

func getKeylogger() *KL.KeyLogger {
	// Find keyboard
	keyboard := KL.FindKeyboardDevice()
	if len(keyboard) <= 0 {
		log.Fatal("No keyboard found. You will need to provide manual input path.")
	} else {
		log.Println(keyboard)
	}

	// Create a keylogger
	keylogger, err := KL.New(keyboard)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	return keylogger
}

func main() {

	go dataserver()

	keylogger := getKeylogger()
	defer keylogger.Close()

	shortcuts := map[string]bool{"CTRL": false, "SHIFT": false, "/": false}

	in := keylogger.Read()
	for i := range in {
		// Listen to only key press & release events
		if i.Type == KL.EvKey && (i.KeyPress() || i.KeyRelease()) {
			gen := strings.Split(i.KeyString(), "_")
			key := gen[len(gen)-1]

			if _, ok := shortcuts[key]; ok {
				shortcuts[key] = i.KeyPress()
			}

			allPressed := true
			for _, v := range shortcuts {
				allPressed = (allPressed && v)
			}
			if allPressed {
				str, err := CB.ReadAll()
				if err != nil {
					panic(err)
				}

				fmt.Println("Shortcut!")
				fmt.Println(str)
			}
		}
	}
}
