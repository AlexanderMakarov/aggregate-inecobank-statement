package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type LineReader struct {
	scanner  *bufio.Scanner
	startKey string
	endKey   string
	inRange  bool
}

func InsideCSVReader(reader io.Reader, startKey, endKey string) *LineReader {
	return &LineReader{
		scanner: bufio.NewScanner(transform.NewReader(
			reader,
			unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder(),
		)),
		startKey: startKey,
		endKey:   endKey,
		inRange:  false,
	}
}

func (lr *LineReader) Read(p []byte) (n int, err error) {
	i := 0
	if !lr.inRange {
		for lr.scanner.Scan() {
			i++
			line := lr.scanner.Text()
			if strings.HasPrefix(line, lr.startKey) {
				fmt.Println("Found start of data on", i, "line.")
				lr.inRange = true
				break
			}
		}
		if !lr.inRange {
			return 0, io.EOF
		}
	}

	buffer := []string{}
	for lr.scanner.Scan() {
		i++
		line := lr.scanner.Text()
		if strings.HasPrefix(line, lr.endKey) {
			fmt.Println("Found end of data on", i, "line.")
			lr.inRange = false
			break
		}
		buffer = append(buffer, line)
	}

	if err := lr.scanner.Err(); err != nil {
		return n, err
	}

	if len(buffer) == 0 {
		return 0, io.EOF
	}

	joinedLines := strings.Join(buffer, "\n")
	n = copy(p, joinedLines)

	return n, nil
}
