package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

func (p *Pipeline) fileReadingStage(fnames chan string, n int) (allLines chan string) {
	lines := make([]chan string, n)
	allLines = make(chan string)
	for i := 0; i < n; i++ {
		lines[i] = make(chan string)
		readFiles(fnames, lines[i])
	}
	go func() {
		defer close(allLines)
		wg := &sync.WaitGroup{}
		for i := range lines {
			wg.Add(1)
			go func(ch chan string) {
				defer wg.Done()
				for line := range ch {
					select {
					case allLines <- line:
						{
							continue
						}
					case <-p.done:
						{
							break
						}
					}
				}
			}(lines[i])
		}
		wg.Wait()
	}()

	return allLines
}

func readFiles(fnames, lines chan string) {
	go func() {
		defer close(lines)
		for fname := range fnames {
			file, err := os.Open(fname)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
				//log.Fatal(err)
			}
			defer file.Close()
			fileScanner := bufio.NewScanner(file)
			fileScanner.Split(bufio.ScanLines)
			for fileScanner.Scan() {
				lines <- fileScanner.Text()
			}
		}
	}()
}
