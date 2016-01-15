/*
The mwexport tool exports the whole mediawiki to a local folder. It saves the raw wikitext of each article into a file named after the article's title.

Usage:

  mwexport [OPTIONS] host username password exportDir
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/stevearm/mediawiki-export/mediawiki"
)

func run() error {
	flag.Set("logtostderr", "true")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] host username password exportDir\n", os.Args[0])
		flag.PrintDefaults()
	}
	var flagVersion = flag.Bool("version", false, "show version")
	flag.Parse()
	if *flagVersion {
		fmt.Fprintf(os.Stderr, "mwexport version: 1.0\n")
		return nil
	}
	if flag.NArg() != 4 {
		flag.Usage()
		return errors.New("")
	}
	var (
		host      = flag.Arg(0)
		username  = flag.Arg(1)
		password  = flag.Arg(2)
		exportDir = flag.Arg(3)
	)
	return export(mediawiki.GetClient(host, username, password), exportDir, localFileSystem{})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
