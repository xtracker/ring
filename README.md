# ring
a lock free buffer for multi-producer and single-consumer pattern

## Usage Example

```go
package main

func main() {
    r := ring.NewRing[int](1000)

    for i := 0; i < 10; i++ {
        go func() {
            r.Offer(1) // concurrent producer
        }()
    }
    
    it := r.Snapshot()
    defer it.Close()
    for v, ok := it.Next(); ok; v, ok = it.Next() {
        // process v
    }
}

```
