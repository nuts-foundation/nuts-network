package concurrency

import (
	"context"
)

type Queue struct {
	c chan interface{}
}

func (q *Queue) Init(size int) {
	q.c = make(chan interface{}, size)
}

// Get gets an item from the queue. It blocks until:
// - There's a document to return
// - The context is cancelled or expires
// - The queue is closed by the producer
// When the queue is closed this function an error and no document.
func (q Queue) Get(cxt context.Context) (interface{}, error) {
	select {
	case <-cxt.Done():
		return nil, cxt.Err()
	case item := <-q.c:
		return item, nil
	}
}

func (q Queue) Add(item interface{}) {
	q.c <- item
}