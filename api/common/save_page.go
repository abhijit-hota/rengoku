package common

import (
	"bufio"
	"context"
	"fmt"
	"mime"
	"os"
	"sort"
	"time"

	"github.com/abhijit-hota/rengoku/server/utils"
	"github.com/go-shiori/obelisk"
)

func SavePage(url string, name string) error {
	req := obelisk.Request{URL: url}
	arc := obelisk.Archiver{
		EnableVerboseLog: true,
		UserAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		RequestTimeout:   time.Duration(5) * time.Second,
	}
	arc.Validate()

	fmt.Println("Started archiving.")

	result, contentType, err := arc.Archive(context.Background(), req)
	if err != nil {
		return err
	}
	fmt.Println("Done archiving.")

	// Choose an extension based on the contentType
	// The longest extension wins
	extensions := utils.MustGet(mime.ExtensionsByType(contentType))
	sort.Slice(extensions, func(i, j int) bool { return len(extensions[i]) > len(extensions[j]) })
	extension := ""

	f, err := os.OpenFile(os.Getenv("SAVE_OFFLINE_PATH")+"/"+name+extension, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	_, err = writer.Write(result)
	if err != nil {
		return err
	}
	fmt.Println("Saved to file.")

	return nil
}
