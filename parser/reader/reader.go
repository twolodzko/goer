package reader

import (
	"bufio"
	"io"
	"strings"
)

type Reader struct {
	*bufio.Reader
	cache string
}

func NewReader(in io.Reader) *Reader {
	return &Reader{bufio.NewReader(in), ""}
}

func (reader *Reader) Next() (string, error) {
	var out string
	isString := false
	for {
		line, err := reader.readLine()
		if err != nil && err != io.EOF {
			return out, err
		}

		isComment := false
		isEscaped := false
		for i, r := range line {
			switch r {
			case '"':
				if !isComment && !isEscaped {
					isString = !isString
				}
			case '%':
				isComment = true
			case '\\':
				isEscaped = true
				continue
			case '.':
				if !isComment && !isEscaped && !isString {
					if len(line) > i+1 {
						reader.cache = line[i+1:]
					}
					out += " " + line[:i+1]
					return strings.TrimSpace(out), nil
				}
			case '\n':
				out += "\n" + line
			}
			isEscaped = false
		}

		if err == io.EOF && len(reader.cache) == 0 {
			return out, err
		}
	}
}

// Read line from cache or input.
func (reader *Reader) readLine() (string, error) {
	if reader.cache != "" {
		cache := reader.cache
		reader.cache = ""
		return cache, nil
	}
	return reader.ReadString('\n')
}
