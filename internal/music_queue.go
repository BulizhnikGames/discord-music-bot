package internal

import (
	"context"
	"iter"
	"log"
	"math/rand/v2"
	"sync"
)

type node struct {
	prev, next *node
	val        *Song
	handled    bool
	//idx        int
}

type MusicQueue struct {
	firstNode, readNode, writeNode, handleNode *node
	Len, Cap                                   int
	mutex                                      *sync.RWMutex
	WriteHandler                               func(ctx context.Context, val *Song) (*Song, error)
	topCtx                                     context.Context
	ctx                                        context.Context
	stopHandlers                               context.CancelFunc
	tryHandleSignal                            chan struct{}
	NewHandled                                 chan struct{}
}

func CreateCycleQueue(ctx context.Context, size int) *MusicQueue {
	lowCtx, cancel := context.WithCancel(ctx)
	//firstNode := &node{idx: 0}
	firstNode := &node{}
	prevNode := firstNode
	for i := 1; i < size; i++ {
		newNode := &node{
			prev: prevNode,
			//idx:  i,
		}
		prevNode.next = newNode
		prevNode = newNode
	}
	firstNode.prev = prevNode
	return &MusicQueue{
		firstNode:       firstNode,
		readNode:        firstNode,
		writeNode:       firstNode,
		handleNode:      firstNode,
		Cap:             size,
		mutex:           &sync.RWMutex{},
		topCtx:          ctx,
		ctx:             lowCtx,
		stopHandlers:    cancel,
		tryHandleSignal: make(chan struct{}, size),
		NewHandled:      make(chan struct{}, size),
	}
}

func (queue *MusicQueue) SetHandler(handler func(ctx context.Context, val *Song) (*Song, error)) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.WriteHandler = handler
}

func (queue *MusicQueue) askHandle() {
	go func() {
		queue.tryHandleSignal <- struct{}{}
	}()
}

func (queue *MusicQueue) notify() {
	go func() {
		queue.NewHandled <- struct{}{}
	}()
}

func (queue *MusicQueue) notNilValToHandle() bool {
	select {
	case <-queue.ctx.Done():
		return false
	default:
		queue.mutex.RLock()
		defer queue.mutex.RUnlock()
		//log.Printf("checking at %d (%p): %+v", queue.handleNode.idx, queue.handleNode, queue.handleNode.val)
		return queue.handleNode.val != nil
	}
}

func (queue *MusicQueue) Run() {
	for {
		select {
		case <-queue.ctx.Done():
			queue.mutex.Lock()
			select {
			case <-queue.topCtx.Done():
				queue.mutex.Unlock()
				return
			default:
				queue.ctx, queue.stopHandlers = context.WithCancel(queue.topCtx)
				queue.handleNode = queue.readNode
				queue.mutex.Unlock()
				queue.askHandle()
			}
		case <-queue.tryHandleSignal:
			for queue.notNilValToHandle() {
				queue.mutex.RLock()
				handled := queue.handleNode.handled
				handleNodeCopy := queue.handleNode
				queue.mutex.RUnlock()
				//log.Printf("trying to handle: %+v", queue.handleNode.val)
				if !handled {
					queue.handleElement(queue.ctx, handleNodeCopy)
				}
				queue.mutex.Lock()
				queue.handleNode = queue.handleNode.next
				queue.mutex.Unlock()
			}
		}
	}
}

func (queue *MusicQueue) handleElement(ctx context.Context, listNode *node) {
	queue.mutex.RLock()
	val := listNode.val
	queue.mutex.RUnlock()
	if val == nil {
		return
	}
	processed, err := queue.WriteHandler(ctx, val)
	if err != nil {
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
	if processed != nil {
		listNode.val = processed
	}
	listNode.handled = true
	queue.notify()
}

func (queue *MusicQueue) Write(v Song) {
	//log.Printf("write")
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.writeNode.val != nil {
		return
	}
	queue.writeNode.val = &v
	//log.Printf("written at %d", queue.writeNode.idx)
	queue.Len++
	queue.writeNode = queue.writeNode.next
	queue.askHandle()
}

func (queue *MusicQueue) ReadHandled() *Song {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.readNode.val == nil {
		return nil
	}
	valCopy := *queue.readNode.val
	queue.readNode.val = nil
	queue.readNode.handled = false
	queue.Len--
	queue.readNode = queue.readNode.next
	return &valCopy
}

func (queue *MusicQueue) Clear() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.writeNode = queue.firstNode
	queue.readNode = queue.firstNode
	queue.handleNode = queue.firstNode
	remover := queue.firstNode
	for i := 0; i < queue.Cap; i++ {
		remover.val = nil
		remover.handled = false
		remover = remover.next
	}
	queue.stopHandlers()
	queue.Len = 0
}

func (queue *MusicQueue) All() iter.Seq[Song] {
	return func(yield func(v Song) bool) {
		queue.mutex.Lock()
		defer queue.mutex.Unlock()
		reader := *queue.readNode
		for reader.val != nil {
			if reader.handled {
				if reader.val.FileURL != "" {
					yield(*reader.val)
				}
			} else {
				yield(*reader.val)
			}
			reader = *reader.next
		}
	}
}

func (queue *MusicQueue) Shuffle() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if queue.Len == 0 {
		return
	}

	nodes := make([]*node, 0, queue.Len)
	reader := queue.readNode

	for reader.val != nil {
		nodes = append(nodes, reader)
		reader = reader.next
	}
	if len(nodes) == 0 {
		return
	}

	rand.Shuffle(len(nodes), func(i, j int) {
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
