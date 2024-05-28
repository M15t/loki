package downloader

import (
	"loki/pkg/parser"
	"sync"
)

// Downloader model
type Downloader struct {
	lock  sync.Mutex
	queue []int

	tsFolder string

	outputFilePath string
	outputFileName string

	finish int32
	segLen int

	result *parser.Result
}

// Task model
type Task struct {
	M3U8URL        string
	OutputFilePath string
	OutputFileName string
	Concurrency    int
}
