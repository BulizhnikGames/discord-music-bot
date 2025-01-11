package internal

import (
	"context"
	"fmt"
	"iter"
	"log"
	"math/rand/v2"
	"strings"
	"sync"
)

type node[T any] struct {
	prev, next *node[T]
	val        *T
	handled    bool
}

type CycleQueue[T any] struct {
	firstNode, readNode, writeNode, handleNode *node[T]
	Len, Cap                                   int
	mutex                                      *sync.RWMutex
	WriteHandler                               func(ctx context.Context, val *T) (*T, error)
	ctx                                        context.Context
	stopHandlers                               context.CancelFunc
	tryHandleSignal                            chan struct{}
}

func CreateCycleQueue[T any](size int) *CycleQueue[T] {
	ctx, cancel := context.WithCancel(context.Background())
	firstNode := &node[T]{}
	prevNode := firstNode
	for i := 1; i < size; i++ {
		newNode := &node[T]{
			prev: prevNode,
		}
		prevNode.next = newNode
		prevNode = newNode
	}
	firstNode.prev = prevNode
	return &CycleQueue[T]{
		firstNode:       firstNode,
		readNode:        firstNode,
		writeNode:       firstNode,
		handleNode:      firstNode,
		Cap:             size,
		mutex:           &sync.RWMutex{},
		ctx:             ctx,
		stopHandlers:    cancel,
		tryHandleSignal: make(chan struct{}, size),
	}
}

func (queue *CycleQueue[T]) SetHandler(handler func(ctx context.Context, val *T) (*T, error)) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.WriteHandler = handler
}

func (queue *CycleQueue[T]) notNilValToHandle() bool {
	select {
	case <-queue.ctx.Done():
		return false
	default:
		queue.mutex.RLock()
		defer queue.mutex.RUnlock()
		return queue.handleNode.val != nil
	}
}

func (queue *CycleQueue[T]) Run() {
	for {
		select {
		case <-queue.ctx.Done():
			queue.mutex.Lock()
			queue.ctx, queue.stopHandlers = context.WithCancel(context.Background())
			queue.handleNode = queue.readNode
			queue.mutex.Unlock()
			queue.tryHandleSignal <- struct{}{}
		case <-queue.tryHandleSignal:
			for queue.notNilValToHandle() {
				queue.mutex.RLock()
				log.Printf("trying to handle: %+v", queue.handleNode.val)
				if !queue.handleNode.handled {
					handleNodeCopy := queue.handleNode
					queue.mutex.RUnlock()
					queue.HandleElement(queue.ctx, handleNodeCopy)
				} else {
					queue.mutex.RUnlock()
					queue.mutex.Lock()
					queue.handleNode = queue.handleNode.next
					queue.mutex.Unlock()
				}
			}
		}
	}
}

func (queue *CycleQueue[T]) HandleElement(ctx context.Context, listNode *node[T]) {
	queue.mutex.RLock()
	val := listNode.val
	queue.mutex.RUnlock()
	if val == nil {
		return
	}
	processed, err := queue.WriteHandler(ctx, val)
	if err != nil {
		processed = val
		log.Printf("couldn't handle new element: %v", err)
	}
	select {
	case <-ctx.Done():
		return
	default:
		break
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	listNode.val = processed
	listNode.handled = err == nil
}

func (queue *CycleQueue[T]) Write(v T) {
	log.Printf("write")
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.writeNode.val != nil {
		return
	}
	queue.writeNode.val = &v
	log.Printf("written")
	queue.Len++
	queue.writeNode = queue.writeNode.next
	queue.tryHandleSignal <- struct{}{}
}

func (queue *CycleQueue[T]) ReadHandled() *T {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.readNode.val == nil || !queue.readNode.handled {
		return nil
	}
	v := *queue.readNode.val
	queue.readNode.val = nil
	queue.Len--
	queue.readNode = queue.readNode.next
	return &v
}

func (queue *CycleQueue[T]) Clear() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.writeNode = queue.firstNode
	queue.handleNode = queue.firstNode
	queue.stopHandlers()
	queue.Len = 0
}

func (queue *CycleQueue[T]) All() iter.Seq[T] {
	return func(yield func(v T) bool) {
		queue.mutex.Lock()
		defer queue.mutex.Unlock()
		reader := *queue.readNode
		for reader.val != nil {
			yield(*reader.val)
			reader = *reader.next
		}
	}
}

func (queue *CycleQueue[T]) Shuffle() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	nodes := make([]*node[T], 0, queue.Len)
	reader := queue.readNode
	builder := strings.Builder{}
	builder.WriteString("Queue:")
	for reader.val != nil {
		builder.WriteString(fmt.Sprintf(" %v", reader.val))
		nodes = append(nodes, reader)
		reader = reader.next
	}
	//log.Println(builder.String())
	rand.Shuffle(queue.Len, func(i, j int) {
		tmpVal, tmpHandled := nodes[i].val, nodes[i].handled
		nodes[i].val = nodes[j].val
		nodes[i].handled = nodes[j].handled
		nodes[j].val = tmpVal
		nodes[j].handled = tmpHandled
	})
	queue.readNode = nodes[0]
	queue.handleNode = queue.readNode
	queue.stopHandlers()
}
