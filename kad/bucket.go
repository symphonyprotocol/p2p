package kad

import (
	"github.com/symphonyprotocol/p2p/node"
)

type BucketNode struct {
	prev *BucketNode
	node *node.RemoteNode
	next *BucketNode
}

func (b *BucketNode) Prev() *BucketNode {
	return b.prev
}

func (b *BucketNode) Node() *node.RemoteNode {
	return b.node
}

func (b *BucketNode) Next() *BucketNode {
	return b.next
}

type BucketQueue struct {
	head *BucketNode
	tail *BucketNode
	size int
}

func (q *BucketQueue) Size() int {
	return q.size
}

func (q *BucketQueue) Tail() *BucketNode {
	return q.tail
}

func (q *BucketQueue) Head() *BucketNode {
	return q.head
}

func (q *BucketQueue) Peek() *node.RemoteNode {
	if q.head == nil {
		return nil
	}
	return q.head.node
}

func (q *BucketQueue) SearchNode(nodeID string) *BucketNode {
	if q.size == 0 {
		return nil
	}
	bnode := q.head
	for {
		if bnode.node.GetID() == nodeID {
			return bnode
		}
	}
	return nil
}

func (q *BucketQueue) Add(remoteNode *node.RemoteNode) {
	bnode := &BucketNode{
		prev: q.tail,
		node: remoteNode,
		next: nil,
	}
	if q.tail == nil {
		q.head = bnode
		q.tail = bnode
	} else {
		q.tail.next = bnode
		q.tail = bnode
	}
	q.size++
	bnode = nil
}

func (q *BucketQueue) Pop() *node.RemoteNode {
	if q.head == nil {
		return nil
	}
	firstNode := q.head
	q.head = firstNode.next
	q.size--
	return firstNode.node
}

func (q *BucketQueue) Search(nodeID string) *node.RemoteNode {
	if q.tail == nil {
		return nil
	}
	tmpNode := q.head
	for {
		if tmpNode == nil {
			break
		}
		if nodeID == tmpNode.node.GetID() {
			return tmpNode.node
		}
		tmpNode = tmpNode.next
	}
	return nil
}

type KBucket struct {
	nodes *BucketQueue
}

func NewKBucket() *KBucket {
	return &KBucket{
		nodes: &BucketQueue{},
	}
}

func (b *KBucket) Add(remoteNode *node.RemoteNode) bool {
	if b.nodes.Size() < BUCKETS_SIZE {
		b.nodes.Add(remoteNode)
		return true
	}
	return false
}

func (b *KBucket) Peek() *node.RemoteNode {
	return b.nodes.Peek()
}

func (b *KBucket) Pop() *node.RemoteNode {
	return b.nodes.Pop()
}

func (b *KBucket) Search(nodeID string) *node.RemoteNode {
	return b.nodes.Search(nodeID)
}

func (b *KBucket) Size() int {
	return b.nodes.Size()
}

func (b *KBucket) MoveToTail(nodeID string) {
	if b.nodes.Size() <= 1 {
		return
	}
	bnode := b.nodes.SearchNode(nodeID)
	if bnode == b.nodes.tail {
		return
	}
	if bnode == b.nodes.head {
		remote := b.Pop()
		b.Add(remote)
	} else {
		prev := bnode.prev
		next := bnode.next
		prev.next = next
		next.prev = prev
		tail := b.nodes.tail
		tail.next = bnode
		bnode.prev = tail
		bnode.next = nil
		b.nodes.tail = bnode
	}
}
