package common

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/go-shiori/obelisk"
)

func SavePage(url string, id int) {
	req := obelisk.Request{URL: url}
	arc := obelisk.Archiver{EnableVerboseLog: true}
	arc.Validate()

	fmt.Println("Started archiving.")
	result, _, err := arc.Archive(context.Background(), req)
	fmt.Println("Done archiving.")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	f, err := os.OpenFile(os.Getenv("SAVE_OFFLINE_FOLDER")+"/"+fmt.Sprint(id)+".html", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	_, err = writer.Write(result)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Saved to file.")
}
