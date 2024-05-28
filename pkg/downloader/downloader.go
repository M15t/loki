package downloader

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"loki/pkg/parser"
	"loki/pkg/tools"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// Start starts a new download task
func (d *Downloader) Start(task *Task) error {
	parserResult, err := parser.Parse(task.M3U8URL)
	if err != nil {
		return err
	}

	// Determine if the output is a file or a directory
	outputFilePath, outputFileName, tsFolder, err := d.setupOutputPaths(task)
	if err != nil {
		return err
	}

	d.outputFilePath = outputFilePath
	d.outputFileName = outputFileName

	d.tsFolder = tsFolder
	d.segLen = len(parserResult.M3U8.Segments)
	d.result = parserResult

	if err := d.downloadSegments(task); err != nil {
		return err
	}

	// divider for downloading and merging
	fmt.Print("\n")

	return d.merge()
}

func (d *Downloader) downloadSegments(task *Task) error {
	var wg sync.WaitGroup
	limitChan := make(chan struct{}, task.Concurrency) // Adjust the buffer size as needed

	d.queue = tools.StartQueue(d.segLen)

	for {
		tsIdx, end, err := d.next()
		if err != nil {
			if end {
				break
			}
			// log.Printf("Error fetching next segment: %s", err)
			continue
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := d.process(idx); err != nil {
				log.Printf("[failed] %s", err)
				if err := d.back(idx); err != nil {
					log.Printf("Error sending segment back to queue: %s", err)
				}
			}
			<-limitChan
		}(tsIdx)
		limitChan <- struct{}{}
	}

	wg.Wait()

	return nil
}

func (d *Downloader) process(segIndex int) error {
	tsFilename := tools.ResolveTSFilename(segIndex)
	tsURL := d.resolveTSURL(segIndex)

	body, err := tools.Get(tsURL)
	if err != nil {
		return fmt.Errorf("request %d failed: %w", segIndex, err)
	}
	defer body.Close()

	fPath := filepath.Join(d.tsFolder, tsFilename)
	if _, err := os.Stat(fPath); err == nil {
		// If the file exists, skip processing
		// log.Printf("File already exists, skipping download: %s", fPath)
		return nil
	}

	fTemp := fPath + tsTempFileSuffix
	f, err := os.Create(fTemp)
	if err != nil {
		return fmt.Errorf("create file %s: %w", tsFilename, err)
	}
	defer f.Close()

	bytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read bytes from %s: %w", tsURL, err)
	}

	// Additional processing (e.g., decryption, trimming) can be refactored into separate functions.
	bytes, err = d.decrytpData(bytes, segIndex)
	if err != nil {
		return err
	}

	if _, err := f.Write(bytes); err != nil {
		return fmt.Errorf("write to %s: %w", fTemp, err)
	}
	defer f.Close()

	if err = os.Rename(fTemp, fPath); err != nil {
		return fmt.Errorf("rename file %s to %s: %w", fTemp, fPath, err)
	}

	// +1 flag until finish
	atomic.AddInt32(&d.finish, 1)

	// fmt.Printf("[loading %6.2f%%] %d %s\n", float32(d.finish)/float32(d.segLen)*100, segIndex, tsFilename)

	// Drawing progress bar
	tools.DrawProgressBar("downloading", float32(d.finish)/float32(d.segLen), progressWidth, "complete")

	return nil
}

func (d *Downloader) next() (segIndex int, end bool, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if len(d.queue) == 0 {
		err = fmt.Errorf("queue empty")
		if d.finish == int32(d.segLen) {
			end = true
			return
		}
		// Some segment indexes are still running.
		end = false
		return
	}

	segIndex = d.queue[0]
	d.queue = d.queue[1:]
	return
}

func (d *Downloader) back(segIndex int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if segIndex < 0 || segIndex >= d.segLen {
		return fmt.Errorf("invalid segment index: %d", segIndex)
	}

	d.queue = append(d.queue, segIndex)
	return nil
}

func (d *Downloader) merge() error {
	missingCount := 0
	for idx := 0; idx < d.segLen; idx++ {
		tsFilename := tools.ResolveTSFilename(idx)
		fPath := filepath.Join(d.tsFolder, tsFilename)
		if _, err := os.Stat(fPath); os.IsNotExist(err) {
			missingCount++
		}
	}

	if missingCount > 0 {
		log.Printf("[warning] %d files missing", missingCount)
	}

	// Create a file to merge all segments into
	mFilePath := filepath.Join(d.outputFilePath, d.outputFileName)
	mFile, err := os.Create(mFilePath)
	if err != nil || mFile == nil {
		return fmt.Errorf("create main TS file failed: %w", err)
	}
	defer mFile.Close()

	writer := bufio.NewWriter(mFile)
	defer writer.Flush()

	mergedCount := 0
	for segIndex := 0; segIndex < d.segLen; segIndex++ {
		tsFilename := tools.ResolveTSFilename(segIndex)
		segmentPath := filepath.Join(d.tsFolder, tsFilename)
		bytes, err := os.ReadFile(segmentPath)
		if err != nil {
			log.Printf("Failed to read file %s: %s", tsFilename, err)
			continue
		}

		if _, err = writer.Write(bytes); err != nil {
			log.Printf("Failed to write to main TS file: %s", err)
			continue
		}
		mergedCount++
		tools.DrawProgressBar("merging", float32(mergedCount)/float32(d.segLen), progressWidth, "complete")
	}

	// Remove temporary TS folder
	if err = os.RemoveAll(d.tsFolder); err != nil {
		fmt.Printf("[warning] Failed to remove temporary folder %s: %s\n", d.tsFolder, err.Error())
	}

	if mergedCount != d.segLen {
		fmt.Printf("[warning] %d files merge failed\n", d.segLen-mergedCount)
	}

	fmt.Printf("\n[output] %s\n", mFilePath)

	return nil
}

func (d *Downloader) decrytpData(data []byte, segIndex int) ([]byte, error) {
	// Decrypt the data if necessary
	sf := d.result.M3U8.Segments[segIndex]
	if sf == nil {
		return nil, fmt.Errorf("invalid segment index: %d", segIndex)
	}
	key, ok := d.result.Keys[sf.KeyIndex]
	if ok && key != "" {
		var err error
		data, err = tools.AES128Decrypt(data, []byte(key), []byte(d.result.M3U8.Keys[sf.KeyIndex].IV))
		if err != nil {
			return nil, fmt.Errorf("decrypt: %s, %s", d.resolveTSURL(segIndex), err.Error())
		}
	}

	// Check for and handle MPEG-TS Sync Byte
	syncByte := uint8(0x47) // MPEG-TS Sync Byte is 0x47
	for i, b := range data {
		if b == syncByte {
			return data[i:], nil // Return the data starting from the Sync Byte
		}
	}

	return data, nil // Return the original data if no Sync Byte is found
}

func (d *Downloader) resolveTSURL(segIndex int) string {
	seg := d.result.M3U8.Segments[segIndex]
	return tools.ResolveURL(d.result.URL, seg.URI)
}

func (d *Downloader) setupOutputPaths(task *Task) (outputFilePath, outputFileName, tsFolder string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", fmt.Errorf("Error: %s", err.Error())
	}

	outputFilePath = task.OutputFilePath
	outputFileName = task.OutputFileName

	if outputFilePath == "" {
		outputFilePath = filepath.Join(homeDir, "Downloads")
	}

	switch {
	case outputFileName == "":
		outputFileName = "output.mp4"
	default:
		if !tools.IsExistedExt(outputFileName) {
			outputFileName = outputFileName + ".mp4"
		}
	}

	tsFolder = filepath.Join(outputFilePath, tsFolderName)

	// Remove temporary TS folder if existed
	if err = os.RemoveAll(tsFolder); err != nil {
		fmt.Printf("[warning] Failed to remove temporary folder %s: %s\n", tsFolder, err.Error())
	}

	// Create output TS folder
	if err := os.MkdirAll(tsFolder, os.ModePerm); err != nil {
		return "", "", "", fmt.Errorf("create storage folder failed: %s", err.Error())
	}

	return outputFilePath, outputFileName, tsFolder, nil
}
