package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Pipeline struct {
	n                 int
	done              chan struct{}
	sortingFile       int
	sortingFileRevers bool
	ignoredHeaderLine bool
}

func NewPipeline(n int, done chan struct{}, sortingFile int, sortingFileRevers, ignoredHeaderLine bool) *Pipeline {
	return &Pipeline{
		n:                 n,
		done:              done,
		sortingFile:       sortingFile,
		sortingFileRevers: sortingFileRevers,
		ignoredHeaderLine: ignoredHeaderLine,
	}
}

func (p *Pipeline) sortContent(content chan string) (res chan string) {
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

func (p *Pipeline) Run(dirName string) (sortedContent chan string) {
	fnChan := p.readDir(dirName)
	contChan := p.fileReadingStage(fnChan, 3)
	return p.sortContent(contChan)
}

func (p *Pipeline) WaitSignal(done chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		close(done)
	}()
}
