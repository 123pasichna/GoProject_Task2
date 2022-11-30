package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	fmt.Println("== Started ==")
	done := make(chan struct{})
	defer close(done)
	var (
		inputFileName     = flag.String("i", "", "Use a file with the name file-name as an input.")
		sortingFile       = flag.Int("f", 0, "Sort input lines by value number N")
		outputFileName    = flag.String("o", "", "Use a file with the name file-name as an output.")
		ignoredHeaderLine = flag.Bool("h", false, "The first line is a header that must be ignored during sorting but included in the output.")
		sortingFileRevers = flag.Bool("r", false, "Sort input lines in reverse order.")
		dirName           = flag.String("d", "", "Dir Name")
	)
	flag.Parse()
	sorter := NewPipeline(3, done, *sortingFile, *sortingFileRevers, *ignoredHeaderLine)
	sorter.WaitSignal(done)
	var content string

	if *inputFileName == "" && *dirName == "" {
		content = sorter.consolInput(*sortingFile, *sortingFileRevers, *ignoredHeaderLine)
	} else if *inputFileName != "" && *dirName != "" {
		log.Fatal("err")
	} else if *inputFileName != "" {
		content = sorter.fileInput(*sortingFile, *sortingFileRevers, *ignoredHeaderLine, *inputFileName)
	} else if *dirName != "" {
		res := sorter.Run(*dirName)
		var stringRes strings.Builder
		for i := range res {
			stringRes.WriteString(i + "\n")
		}
		content = stringRes.String()
	}
	if content != "" {
		fmt.Println("result:\n" + content)
		sorter.writeFile(content, *outputFileName)
	}
	if inputFileName != nil {
		fmt.Printf("Input file name: %s\n", *inputFileName)
	}
	if sortingFile != nil {
		fmt.Printf("sortingFile: %d\n", *sortingFile)
	}
	if outputFileName != nil {
		fmt.Printf("outputFileName: %s\n", *outputFileName)
	}
	if ignoredHeaderLine != nil {
		fmt.Printf("ignoredHeaderLine: %s\n", *inputFileName)
	}
	if sortingFileRevers != nil {
		fmt.Printf("sortingField: %v\n", *sortingFileRevers)
	}
	if dirName != nil {
		fmt.Printf("dirName: %v\n", *dirName)
	}
	fmt.Println("== Finished ==")

}

func (p *Pipeline) sorting(content chan string) (res chan string) {
	res = make(chan string)
	go func() {
		defer close(res)
		var buffer = make([]string, 0, 1000)
		for line := range content {
			buffer = append(buffer, line)
		}
		content := p.varParametr(p.sortingFile, p.sortingFileRevers, p.ignoredHeaderLine, buffer)
		for _, i := range content {
			select {
			case res <- i:
				{
					continue
				}
			case <-p.done:
				{
					break
				}
			}
		}
	}()
	return res
}

func (p *Pipeline) consolInput(sortingFieldIndex int, isReversedOrder, isNotIgnoreHeader bool) string {
	scanner := bufio.NewScanner(os.Stdin)
	return p.splitData(sortingFieldIndex, isReversedOrder, isNotIgnoreHeader, scanner)
}

func (p *Pipeline) fileInput(sortingFieldIndex int, isReversedOrder, isNotIgnoreHeader bool, inputFile string) string {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	content := p.splitData(sortingFieldIndex, isReversedOrder, isNotIgnoreHeader, fileScanner)
	return content
}

func (p *Pipeline) splitData(sortingFieldIndex int, isReversedOrder, isNotIgnoreHeader bool, scanner *bufio.Scanner) string {
	var header string
	n := 0
	table := [][]string{}
	for scanner.Scan() {
		line := scanner.Text()
		row := strings.Split(line, ",")
		if n == 0 {
			n = len(row)
			if isNotIgnoreHeader {
				header = line
				continue
			}
		}
		if line == "" {
			break
		}
		if n != len(row) {
			log.Fatalf("Error: row has %d columns, but must have %d\n", len(row), n)
		}
		table = append(table, row)
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
	sort.Slice(table, func(i, j int) bool {
		return p.process(table[i][sortingFieldIndex], table[j][sortingFieldIndex], isReversedOrder)
	})
	var result strings.Builder
	if header != "" {
		result.WriteString(header)
		result.WriteString("\n")
	}
	for _, row := range table {
		result.WriteString(strings.Join(row, ","))
		result.WriteString("\n")
	}
	return result.String()
}

func (p *Pipeline) varParametr(sortingFieldIndex int, isReversedOrder, isNotIgnoreHeader bool, buffer []string) []string {
	var header string
	n := 0
	table := [][]string{}
	for _, line := range buffer {
		row := strings.Split(line, ",")
		if n == 0 {
			n = len(row)
			if isNotIgnoreHeader {
				header = line
				continue
			}
		}
		if line == "" {
			break
		}
		if n != len(row) {
			log.Fatalf("Error: row has %d columns, but must have %d\n", len(row), n)
		}
		table = append(table, row)
	}
	sort.Slice(table, func(i, j int) bool {
		return p.process(table[i][sortingFieldIndex], table[j][sortingFieldIndex], isReversedOrder)
	})
	var result []string
	if header != "" {
		result = append(result, header)
	}
	for _, row := range table {
		result = append(result, strings.Join(row, ","))
	}
	return result
}

func (p *Pipeline) writeFile(content, fileName string) {
	if fileName != "" {
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		_, err = file.WriteString(content)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (p *Pipeline) process(first, next string, isReversed bool) bool {
	return first < next != isReversed
}
