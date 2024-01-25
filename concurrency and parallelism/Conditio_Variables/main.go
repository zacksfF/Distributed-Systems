package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	num_money = 100
	lock = sync.Mutex{}
	moneyDeposited  = sync.NewCond(&lock)
)

func stingy(){
	for i := 1; i <= 10000; i++{
		lock.Lock()
		num_money += 100
		fmt.Println("Stingy sees balance of", num_money)
		moneyDeposited.Signal()
		lock.Unlock()
		time.Sleep(1 * time.Millisecond)
	}
	println("stingy Done")
}
func spendy(){
	for i := 1; i <= 10000; i++{
		lock.Lock()
		for num_money-20 < 0{
			moneyDeposited.Wait()
		}
		num_money -= 20
		fmt.Println("Spendy sees balance of", num_money)
		lock.Unlock()
		time.Sleep(1 * time.Millisecond)
	}
	println("stingy Done")
}

func main(){
	go stingy()
	go spendy()
	time.Sleep(3000 * time.Millisecond)
	print(num_money)
}

