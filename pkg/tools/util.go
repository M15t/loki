package tools

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	tsExt = ".ts"
)

// StartQueue returns a slice of integers from 0 to len-1
func StartQueue(len int) []int {
	q := make([]int, 0)
	for i := 0; i < len; i++ {
		q = append(q, i)
	}

	return q
}

// DrawProgressBar draws a progress bar with a prefix, a proportion indicating progress, a specified width, and optional suffixes.
func DrawProgressBar(prefix string, proportion float32, width int, suffix ...string) {
	if proportion > 1 {
		proportion = 1
	} else if proportion < 0 {
		proportion = 0
	}

	// Calculate the position of the progress indicator
	pos := int(proportion * float32(width))
	bar := fmt.Sprintf("%s%s", strings.Repeat("■", pos), strings.Repeat(" ", width-pos))

	// Join the suffixes with a space
	suffixStr := strings.Join(suffix, " ")

	// Calculate the total length of the prefix, bar, and suffix
	totalLength := utf8.RuneCountInString(prefix) + width + utf8.RuneCountInString(fmt.Sprintf("%6.2f%%", proportion*100)) + utf8.RuneCountInString(suffixStr) + 4

	// Adjust the width to fit the terminal window if necessary
	if totalLength > width {
		width = totalLength
	}

	// Construct the final progress bar string
	s := fmt.Sprintf("[%s] %s %6.2f%% %s", prefix, bar, proportion*100, suffixStr)

	// Print the progress bar
	fmt.Print("\r" + s)
}

// ResolveTSFilename returns ts filename
func ResolveTSFilename(ts int) string {
	return strconv.Itoa(ts) + tsExt
}

// IsExistedExt checks if the given file path has an extension.
func IsExistedExt(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext != ""
}
