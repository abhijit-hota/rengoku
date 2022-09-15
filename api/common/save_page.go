package common

import (
	"context"
	"fmt"
	"os"

	"github.com/go-shiori/obelisk"
)

func SavePage(url string, name string) error {
	req := obelisk.Request{URL: url}
	arc := obelisk.Archiver{
		EnableVerboseLog: true,
		UserAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		// RequestTimeout:   time.Duration(5) * time.Second,
	}
	arc.Validate()

	fmt.Println("Started archiving.")

	result, _, err := arc.Archive(context.Background(), req)
	if err != nil {
		return err
	}
	fmt.Println("Done archiving.")

	f, err := os.OpenFile(os.Getenv("SAVE_OFFLINE_PATH")+"/"+name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(result)
	if err != nil {
		return err
	}
	fmt.Println("Saved to file.")

	return nil
}
