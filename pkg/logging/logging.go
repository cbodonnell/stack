package logging

import (
	"bufio"
	"io"
	"log"
)

func LogReaderWithPrefix(r io.Reader, prefix string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("%s%s", prefix, scanner.Text())
	}
}
