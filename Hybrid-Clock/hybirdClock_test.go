package hybridclock

import (
	"encoding/json"
	"fmt"
	hybridclock "hybirdclock"
	"testing"
)

func TestLogic(t *testing.T) {
	clock := hybridclock.New()

	ts := clock.Now()

	clock.Update(ts)

	if !ts.Less(clock.Now()) {
		t.Fatal("ts is less than current clock time")
	}
}

func TestJsonHLC(t *testing.T) {
	clock := hybridclock.New()

	value := struct {
		Timestamp *hybridclock.Timestamp `json:"timestamp"`
	}{
		Timestamp: clock.Now(),
	}

	fmt.Println(value)

	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))

	err = json.Unmarshal(b, &value)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(value)
}
