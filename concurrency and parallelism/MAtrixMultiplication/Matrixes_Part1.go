package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	matrixSize = 500
)

var (
	matrix = [matrixSize][matrixSize] int{}
	matrixA = [matrixSize][matrixSize] int{}
	matrixB = [matrixSize][matrixSize] int{}
	Result = [matrixSize][matrixSize] int{}
	rwLock = sync.RWMutex{}
	cond = sync.NewCond(rwLock.RLocker())
	waitGroup = sync.WaitGroup{}
)

func generateRandomMatrix(matrix *[matrixSize][matrixSize]int){
	for row := 0; row < matrixSize; row++{
		for col := 0; col < matrixSize; col++{
			matrix[row][col] += rand.Intn(10) - 5
		}
	}
}

func workOutRow (row int){
	rwLock.RLock()
	for {
		waitGroup.Done()
		cond.Wait()
		for col := 0; col < matrixSize; col++{
			for i := 0; i<matrixSize; i++{
				Result[row][col] += matrix[row][i] * matrixB[i][col]
			}
		}
	}
}

func main(){
	fmt.Println("Is Working Let him cook hh ...")
	waitGroup.Add(matrixSize)
	for row := 0; row < matrixSize; row ++{
		go workOutRow(row)
	}

	start := time.Now()
	for i := 0; i< 100; i++{
		waitGroup.Wait()
		rwLock.Lock()
		generateRandomMatrix(&matrixA)
		generateRandomMatrix(&matrixB)
		waitGroup.Add(matrixSize)
		rwLock.Unlock()
		cond.Broadcast()
	}
	elap := time.Since(start)
	fmt.Println("Done")
	fmt.Printf("Processing Took %s\n", elap)
}