package input

import (
	"bufio"
	"fmt"
	"os"
)

// ReadFile reads log file line by line
func ReadFile(filePath string) ([]string, error) {
	var lines []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// PrintLines prints logs with line numbers
func PrintLines(lines []string) {
	fmt.Println("\n📄 Reading logs...")

	for i, line := range lines {
		fmt.Printf("%d: %s\n", i+1, line)
	}
}
