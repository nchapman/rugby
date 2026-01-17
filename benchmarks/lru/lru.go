package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	A uint32 = 1103515245
	C uint32 = 12345
	M uint32 = 1 << 31
)

type LCG struct {
	seed uint32
}

func NewLCG(seed uint32) *LCG {
	return &LCG{seed: seed}
}

func (l *LCG) Next() uint32 {
	l.seed = (A*l.seed + C) % M
	return l.seed
}

type LRUNode struct {
	key   uint32
	value uint32
	prev  *LRUNode
	next  *LRUNode
}

type LRU struct {
	size  int
	cache map[uint32]*LRUNode
	head  *LRUNode
	tail  *LRUNode
	count int
}

func NewLRU(size int) *LRU {
	return &LRU{
		size:  size,
		cache: make(map[uint32]*LRUNode),
	}
}

func (l *LRU) Get(key uint32) (uint32, bool) {
	node, ok := l.cache[key]
	if !ok {
		return 0, false
	}
	l.moveToEnd(node)
	return node.value, true
}

func (l *LRU) Put(key, value uint32) {
	if node, ok := l.cache[key]; ok {
		node.value = value
		l.moveToEnd(node)
		return
	}

	if l.count == l.size {
		l.evict()
	}

	node := &LRUNode{key: key, value: value}
	l.cache[key] = node
	l.addToEnd(node)
	l.count++
}

func (l *LRU) moveToEnd(node *LRUNode) {
	l.removeNode(node)
	l.addToEnd(node)
}

func (l *LRU) removeNode(node *LRUNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}
}

func (l *LRU) addToEnd(node *LRUNode) {
	node.prev = l.tail
	node.next = nil
	if l.tail != nil {
		l.tail.next = node
	}
	l.tail = node
	if l.head == nil {
		l.head = node
	}
}

func (l *LRU) evict() {
	if l.head != nil {
		delete(l.cache, l.head.key)
		l.removeNode(l.head)
		l.count--
	}
}

func main() {
	size := 100
	if len(os.Args) > 1 {
		if s, err := strconv.Atoi(os.Args[1]); err == nil {
			size = s
		}
	}
	n := 10000
	if len(os.Args) > 2 {
		if s, err := strconv.Atoi(os.Args[2]); err == nil {
			n = s
		}
	}

	mod := uint32(size) * 10
	rng0 := NewLCG(0)
	rng1 := NewLCG(1)
	hit := 0
	missed := 0
	lru := NewLRU(size)

	for i := 0; i < n; i++ {
		n0 := rng0.Next() % mod
		lru.Put(n0, n0)

		n1 := rng1.Next() % mod
		if _, ok := lru.Get(n1); ok {
			hit++
		} else {
			missed++
		}
	}

	fmt.Printf("%d\n%d\n", hit, missed)
}
