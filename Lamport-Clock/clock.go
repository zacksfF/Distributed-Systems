package lamportclock

import (
	"sync"
)

var Default = Clock{}

type Clock struct {
	val uint64
	mut sync.Mutex
}

func (c *Clock) Lamp(L uint64) uint64 {
	c.mut.Lock()
	if L > c.val {
		c.val = L + 1
		c.mut.Unlock()
		return L + 1
	} else {
		c.val++
		L = c.val
		c.mut.Unlock()
		return v
	}
}
