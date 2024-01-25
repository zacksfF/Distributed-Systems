package common

import "sync"

type Train struct {
	Id          int
	TrainLength int
	Front       int
}

type Inytersection struct{
	Id int
	Mutex sync.Mutex
	LockedBy int
}

type Crossing struct {
	Position int
	Intersection *Inytersection
}


