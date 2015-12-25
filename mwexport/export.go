package main

import (
    "errors"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/stevearm/mediawiki-export/mediawiki"
)

func run() error {
    log.Fatal("asdf")
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
        host = flag.Arg(0)
        username = flag.Arg(1)
        password = flag.Arg(2)
        exportDir = flag.Arg(3)
    )
    return export(host, username, password, exportDir)
}

func export(host, username, password, exportDir string) error {
    fmt.Printf("Export from %v:%v@%v to %v\n", username, password, host, exportDir)
    // mediawiki.DoThing(host, username, password, exportDir)
    client := mediawiki.NewClient(host, username, password)
    client.Login()
    client.Login()
    return nil
}

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }
}
