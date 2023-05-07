package ring

import (
	"runtime"
	_ "unsafe"

	"github.com/xtracker/ring/internal/slot"
)

func NewRing[T any](sz int) *Ring[T] {
	r := &Ring[T]{
		slots: make([]*slot.Ring[T], runtime.GOMAXPROCS(0)),
	}

	for i := range r.slots {
		r.slots[i] = slot.NewRing[T](sz)
	}

	return r
}

type Ring[T any] struct {
	slots []*slot.Ring[T]
}

func (r *Ring[T]) Offer(dp T) bool {
	pid := procPin()
	defer procUnPin()

	if pid >= len(r.slots) {
		return false
	}

	return r.slots[pid].Offer(dp)
}

func (r *Ring[T]) Snapshot() Iterator[T] {
	its := iterators[T]{}
	for _, s := range r.slots {
		its = append(its, s.Snapshot())
	}

	return its
}

type Iterator[T any] interface {
	slot.Iterator[T]
}

type iterators[T any] []slot.Iterator[T]

func (i iterators[T]) Next() (T, bool) {
	for _, it := range i {
		v, ok := it.Next()
		if ok {
			return v, true
		}
	}

	var zero T
	return zero, false
}

func (i iterators[T]) Close() {
	for _, it := range i {
		it.Close()
	}
}

//go:linkname procPin runtime.procPin
func procPin() int

//go:linkname procUnPin runtime.procUnpin
func procUnPin()
