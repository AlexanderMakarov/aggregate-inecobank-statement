package main

import (
	"bufio"
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
		scanner:  bufio.NewScanner(transform.NewReader(reader, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder())),
		startKey: startKey,
		endKey:   endKey,
		inRange:  false,
	}
}

func (lr *LineReader) Read(p []byte) (n int, err error) {
	if !lr.inRange {
		for lr.scanner.Scan() {
			line := lr.scanner.Text()
			if strings.HasPrefix(line, lr.startKey) {
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
		line := lr.scanner.Text()
		if strings.HasPrefix(line, lr.endKey) {
			lr.inRange = false
			break
		}
		buffer = append(buffer, line+"\n")
	}

	if err := lr.scanner.Err(); err != nil {
		return n, err
	}

	if len(buffer) == 0 {
		return 0, io.EOF
	}

	joinedLines := strings.Join(buffer, "\n")
	n += copy(p, joinedLines)

	return n, nil
}
