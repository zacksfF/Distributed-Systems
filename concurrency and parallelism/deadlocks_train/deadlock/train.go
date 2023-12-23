package deadlock

import (
	"log"
	"time"

	. "github.com/cutajarj/multithreadingingo/deadlocks_train/common"
)

func MoveTrain(train *Train, distance int, crossing[]*Crossing){
	for train.Front < distance{
		train.Front += 1
		for _, crossing := range crossing{
			if train.Front == crossing.Position{
				crossing.Intersection.Mutex.Lock()
				crossing.Intersection.LockedBy = train.Id
			}
			back := train.Front - train.TrainLength
			if back == crossing.Position{
				crossing.Intersection.LockedBy = -1
				crossing.Intersection.Mutex.Unlock()
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
}