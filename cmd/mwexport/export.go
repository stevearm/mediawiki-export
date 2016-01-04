package main

import (
	"fmt"

	"github.com/stevearm/mediawiki-export/mediawiki"
)

func export(host, username, password, exportDir string) error {
	client := &mediawiki.Client{
		Host:     host,
		Username: username,
		Password: password,
	}
	err := client.Login()
	if err != nil {
		return err
	}
	titles, err := client.ListArticleTitles()
	if err != nil {
		return err
	}
	for _, title := range titles {
		fmt.Printf("Found title: %v\n", title)
	}
	return nil
}
