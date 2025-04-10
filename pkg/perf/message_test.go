package perf

import (
	"strconv"
	"sync"
	"testing"
)

func TestUniqSeq(t *testing.T) {
	messageGetter, err := NewDefaultMessageGetter(`{{UniqSeq "test" 10}}`)
	if err != nil {
		t.Error(err)
		return
	}

	uniqElems := map[int]int{}
	mux := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 50000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, release := messageGetter.Get(nil)
			defer release()

			if err != nil {
				t.Error(err)
				return
			}

			intVal, err := strconv.Atoi(string(val))
			if err != nil {
				t.Error(err)
				return
			}

			mux.Lock()
			defer mux.Unlock()

			uniqElems[intVal]++
		}()
	}

	wg.Wait()

	for k, v := range uniqElems {
		if v > 1 {
			t.Error(k, v)
		}
	}
}
