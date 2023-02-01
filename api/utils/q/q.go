package q

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
)

func qPrint(i any) {
	s := ""
	if str, ok := i.(string); ok {
		s = str
	} else {
		byt, _ := json.MarshalIndent(i, "", "\t")
		s = string(byt)
	}

	_, file, line, _ := runtime.Caller(2)

	fmt.Printf("%s:%d\t", filepath.Base(file), line)
	fmt.Println(s)
	fmt.Println("---")
}

func Q(i ...any) {
	fmt.Println("========================================================================================")
	for _, v := range i {
		qPrint(v)
	}
	fmt.Println("========================================================================================")
}
