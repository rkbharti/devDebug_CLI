package input

import (
	"bufio"
	"os"
)

// ProcessFile reads a file line-by-line and calls handle() for each line.
// Each line is parsed (JSON or plain text) before being passed to handle.
func ProcessFile(filepath string, handle func(ParsedLine, int)) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// increase buffer size — JSON log lines can be long
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineNum := 1

	for scanner.Scan() {
		parsed := ParseLine(scanner.Text())
		handle(parsed, lineNum)
		lineNum++
	}

	return scanner.Err()
}
