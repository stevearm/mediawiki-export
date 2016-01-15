package main

import (
	"fmt"
	"regexp"

	"github.com/stevearm/mediawiki-export/mediawiki"
)

func export(client mediawiki.Client, exportDir string, fs fileSystem) error {
	titles, err := client.ListArticleTitles()
	if err != nil {
		return err
	}
	var scrubber scrubber
	err = scrubber.Init()
	if err != nil {
		return err
	}
	uniqueTitles := make(map[string]struct{})
	for _, title := range titles {
		scrubbedTitle := scrubber.Scrub(title)
		if _, found := uniqueTitles[scrubbedTitle]; found {
			return fmt.Errorf("Found duplicate title: %s", scrubbedTitle)
		}
		uniqueTitles[scrubbedTitle] = struct{}{}
	}
	for title := range uniqueTitles {
		article, err := client.GetArticle(title)
		if err != nil {
			return err
		}
		filename := fmt.Sprintf("%s.txt", title)
		articleBytes := []byte(article)
		err = fs.WriteFile(filename, articleBytes, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

type scrubber struct {
	regex *regexp.Regexp
}

func (s *scrubber) Init() error {
	regex, err := regexp.Compile("[^A-Za-z0-9_-]")
	if err != nil {
		return err
	}
	s.regex = regex
	return nil
}

func (s scrubber) Scrub(title string) string {
	return s.regex.ReplaceAllString(title, "_")
}
