package tunnel

import (
	"container/list"
	"io"
	"sync"
)

type Queue struct {
	lock   sync.Mutex
	ev     *sync.Cond
	queue  *list.List
	closed bool
}

func NewQueue() (q *Queue) {
	q = &Queue{
		queue:  list.New(),
		closed: false,
	}
	q.ev = sync.NewCond(&q.lock)
	return
}

func (q *Queue) Push(v interface{}) (err error) {
	logger.Debugf("push queue: %p", q)
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return io.ErrClosedPipe
	}
	q.queue.PushBack(v)
	q.ev.Signal()
	return
}

func (q *Queue) Pop(block bool) (v interface{}, err error) {
	logger.Debugf("pop queue: %p, block: %t", q, block)
	q.lock.Lock()
	defer q.lock.Unlock()
	var e *list.Element
	for e = q.queue.Front(); e == nil; e = q.queue.Front() {
		if q.closed {
			return nil, io.EOF
		}
		if !block {
			return
		}
		q.ev.Wait()
	}
	v = e.Value
	q.queue.Remove(e)
	return
}

func (q *Queue) Close() (err error) {
	logger.Debugf("close queue: %p", q)
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	q.ev.Broadcast()
	return
}
