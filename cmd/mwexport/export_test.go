package main

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stevearm/mediawiki-export/mediawiki"
)

func TestExport(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := mediawiki.NewMockClient(mockCtrl)
	mockClient.EXPECT().ListArticleTitles().Return([]string{
		"FirstArticle",
		"SecondArticle",
	}, nil)

	mockClient.EXPECT().GetArticle("FirstArticle").Return("This is the first article", nil)
	mockClient.EXPECT().GetArticle("SecondArticle").Return("This is the second article", nil)

	mockFileSystem := NewMockfileSystem(mockCtrl)
	fileMode := os.FileMode(0644)
	mockFileSystem.EXPECT().WriteFile("FirstArticle.txt", []byte("This is the first article"), fileMode).Return(nil)
	mockFileSystem.EXPECT().WriteFile("SecondArticle.txt", []byte("This is the second article"), fileMode).Return(nil)

	err := export(mockClient, "outputFolder", mockFileSystem)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDuplicateNames(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := mediawiki.NewMockClient(mockCtrl)
	mockClient.EXPECT().ListArticleTitles().Return([]string{
		"FirstArticle",
		"SecondArticle",
		"Third One",
		"Third_One",
	}, nil)

	err := export(mockClient, "outputFolder", nil)
	if err == nil {
		t.Errorf("Should have failed on duplicate names")
	}
}

func TestScrubTitle(t *testing.T) {
	var scrubber scrubber
	err := scrubber.Init()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	assertScrubTitle(t, scrubber, "House", "House")
	assertScrubTitle(t, scrubber, "OtherHouse", "OtherHouse")
	assertScrubTitle(t, scrubber, "Other House", "Other_House")
	assertScrubTitle(t, scrubber, "Other_House", "Other_House")
	assertScrubTitle(t, scrubber, "Other's House", "Other_s_House")
}

func assertScrubTitle(t *testing.T, scrubber scrubber, source, expected string) {
	actual := scrubber.Scrub(source)
	if actual != expected {
		t.Errorf("Scrub %v to %v failed: %v", source, expected, actual)
	}
}
