package main

import (
	"github.com/otiai10/gosseract/v2"
)

func extractTextWithClient(client *gosseract.Client, path string) (string, error) {
	client.SetImage(path)
	text, err := client.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}


