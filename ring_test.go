package ring

import (
	"sync"
	"testing"
)

func TestSimple(t *testing.T) {
	r := NewRing[int](2)
	ok := r.Offer(10)
	if !ok {
		t.Fail()
	}

	it := r.Snapshot()
	defer it.Close()
	ret, ok := it.Next()
	if !ok {
		t.Fail()
	}

	if ret != 10 {
		t.Fail()
	}
}

func TestConcurrentWrite(t *testing.T) {
	r := NewRing[int](100)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			r.Offer(idx)
			wg.Done()
		}(i)
	}

	wg.Wait()

	it := r.Snapshot()
	defer it.Close()

	cnt := 0
	for _, ok := it.Next(); ok; _, ok = it.Next() {
		cnt++
	}

	if cnt != 10 {
		t.Fatalf("cnt = %d\n", cnt)
	}
}
