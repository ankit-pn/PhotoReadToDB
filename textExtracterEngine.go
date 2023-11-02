package main

import (
	"fmt"
	"github.com/otiai10/gosseract/v2"
)

func extractText(path string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(path)
	client.SetLanguage("eng", "hin", "urd")
	text, err := client.Text()
	if err != nil {
		fmt.Println("Error: ", err)
		return "", err
	}
	return text, nil
}


