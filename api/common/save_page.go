package common

import (
	"context"
	"os"

	"github.com/abhijit-hota/rengoku/server/utils/log"
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

	log.Info.Printf("Started archiving: (%s)", url)

	result, _, err := arc.Archive(context.Background(), req)
	if err != nil {
		return err
	}
	log.Info.Printf("Done archiving: (%s)", url)

	f, err := os.OpenFile(os.Getenv("SAVE_OFFLINE_PATH")+name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(result)
	if err != nil {
		return err
	}
	log.Info.Printf("Saved to file: (%s)", url)

	return nil
}
