package main

import (
	"log"
	"os"
	"path/filepath"
)

func (p *Pipeline) readDir(dir string) (fnames chan string) {
	fnames = make(chan string)
	go func() {
		defer close(fnames)
		fileList := p.iter(dir)
		for _, f := range fileList {
			select {
			case fnames <- f:
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
	return fnames
}

func (p *Pipeline) iter(path string) (files []string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files
}
