package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

var (
	initialString string
	finalString   string
	stringLength  int
)

func capitalize(letterChannel chan string, currentLetter string, wg *sync.WaitGroup) {
	thisLetter := strings.ToUpper(currentLetter)
	wg.Done()
	letterChannel <- thisLetter
}

func addToFinalStack(letterChannel chan string, wg *sync.WaitGroup) {
	letter := <-letterChannel
	finalString += letter
	wg.Done()
}

func main() {
	runtime.GOMAXPROCS(2)
	var wg sync.WaitGroup
	initialString = `Always check for the latest developments and consider the specific
	 needs of your project when selecting a framework for distributed systems, as the 
	 field is dynamic, and new technologies may emerge over time.`
	initialBytes := []byte(initialString)
	var letterChannel = make(chan string)
	stringLength = len(initialBytes)
	for i := 0; i < stringLength; i++ {
		wg.Add(2)
		go capitalize(letterChannel, string(initialBytes[i]), &wg)
		go addToFinalStack(letterChannel, &wg)
		wg.Wait()
	}
	fmt.Println(finalString)
}
