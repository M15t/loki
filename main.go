package main

import (
	"flag"
	"fmt"
	"loki/pkg/downloader"
	"os"
	"time"
)

// Declare
var (
	url    string
	output string
	name   string
)

const (
	concurrency = 100
)

// sample
// https://surrit.com/78f7dda9-2d95-448b-9ce5-d09b458c8fdd/842x480/video.m3u8

func init() {
	flag.StringVar(&url, "u", "", "URL to fetch")
	flag.StringVar(&output, "o", "", "Output path")
	flag.StringVar(&name, "n", "output", "File name")
}

func main() {
	fmt.Println("Well well, here we go again")
	start := time.Now()

	// Parse command-line flags
	flag.Parse()

	// Validate required flags
	if err := validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		flag.Usage()
		os.Exit(1)
	}

	dl := downloader.New()
	if err := dl.Start(&downloader.Task{
		M3U8URL:        url,
		OutputFilePath: output,
		OutputFileName: name,
		Concurrency:    concurrency,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("[elapsed] %v\n", time.Since(start))
}

func validate() error {
	if url == "" {
		return fmt.Errorf("parameter '-u' (M3U8 URL) is required")
	}

	return nil
}
