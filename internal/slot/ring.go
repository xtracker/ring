package slot

import "sync/atomic"

// lock free Ring buffer
// push & pop will not be access in same thread

func NewRing[T any](sz int) *Ring[T] {
	return &Ring[T]{
		dps: make([]T, sz),
	}
}

type Ring[T any] struct {
	head, tail uint64
	size       uint64
	dps        []T
	Nil        T
}

func (r *Ring[T]) increment(cur uint64) uint64 {
	cur++
	if r.size&(r.size-1) == 0 {
		return (cur) & (r.size - 1)
	}

	return cur % r.size
}

func (r *Ring[T]) Len() int {
	tail := atomic.LoadUint64(&r.tail)
	head := atomic.LoadUint64(&r.head)

	return int((tail + r.size - head) % (r.size))
}

func (r *Ring[T]) Offer(dp T) bool {
	tail := atomic.LoadUint64(&r.tail)
	nextTail := r.increment(tail)
	if nextTail == atomic.LoadUint64(&r.head) {
		return false // full
	}

	r.dps[tail] = dp
	atomic.StoreUint64(&r.tail, nextTail)
	return true
}

func (r *Ring[T]) Poll() (T, bool) {
	head := atomic.LoadUint64(&r.head)
	tail := atomic.LoadUint64(&r.tail)
	if head == tail {
		return r.Nil, false // empty, direct return
	}

	nextHead := r.increment(head)
	dp := r.dps[head]
	atomic.StoreUint64(&r.tail, nextHead)
	return dp, true
}

// use Iterator to fetch ring data
// poll will set mem barrier each element
type Iterator[T any] interface {
	Next() (T, bool)
	Close()
}

type snapshotIterator[T any] struct {
	*Ring[T]
	target  uint64
	current uint64
}

func (si *snapshotIterator[T]) Next() (T, bool) {
	if si.current == si.target {
		return si.Nil, false
	}

	dp := si.dps[si.current]
	si.current = si.increment(si.current)
	return dp, true
}

func (si *snapshotIterator[T]) Close() {
	atomic.StoreUint64(&si.head, si.current)
}

func (r *Ring[T]) Snapshot() Iterator[T] {
	return &snapshotIterator[T]{
		Ring:    r,
		target:  atomic.LoadUint64(&r.tail),
		current: atomic.LoadUint64(&r.head),
	}
}
