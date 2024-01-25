package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

var (
	matches   []string
	waitGroup = sync.WaitGroup{}
	lock      = sync.Mutex{}
)

func FileSearch(root string, filename string) {
	fmt.Println("seraching in", root)
	files, _ := ioutil.ReadDir(root)
	for _, file := range files {
		if strings.Contains(file.Name(), filename) {
			lock.Lock()
			matches = append(matches, filepath.Join(root, file.Name()))
			lock.Unlock()
		}
		if file.IsDir() {
			waitGroup.Add(1)
			go FileSearch(filepath.Join(root, file.Name()), filename)
		}
	}
	waitGroup.Done()
}

func main() {
	waitGroup.Add(1)
	go FileSearch("/Users/zakariasaif", "Go")
	waitGroup.Wait()
	for _, file := range matches {
		fmt.Println("mathched", file)
	}
}
