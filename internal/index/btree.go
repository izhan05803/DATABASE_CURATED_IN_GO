package index

// BTree implements a B-Tree index structure.
// The order is the maximum number of children per node.
type BTree struct {
	root      *Node
	order     int
	minDegree int
	size      int
}

// Node represents a B-Tree node.
type Node struct {
	keys     []string
	values   []uint32
	children []*Node
	isLeaf   bool
}

// NewBTree creates a new B-Tree with the given order.
func NewBTree(order int) *BTree {
	if order < 3 {
		order = 3
	}
	minDegree := (order + 1) / 2

	return &BTree{
		order:     order,
		minDegree: minDegree,
	}
}

// Search finds a key and returns its page ID.
func (bt *BTree) Search(key string) (uint32, bool) {
	node := bt.root
	for node != nil {
		i := bt.findKey(node, key)
		if i < len(node.keys) && node.keys[i] == key {
			return node.values[i], true
		}
		if node.isLeaf {
			return 0, false
		}
		node = node.children[i]
	}
	return 0, false
}

// Insert adds or updates a key-value pair in the B-Tree.
func (bt *BTree) Insert(key string, pageID uint32) error {
	if bt.root == nil {
		bt.root = &Node{
			keys:   []string{key},
			values: []uint32{pageID},
			isLeaf: true,
		}
		bt.size = 1
		return nil
	}

	if bt.updateIfExists(bt.root, key, pageID) {
		return nil
	}

	if len(bt.root.keys) == bt.maxKeys() {
		newRoot := &Node{
			isLeaf:   false,
			children: []*Node{bt.root},
		}
		bt.splitChild(newRoot, 0)
		bt.root = newRoot
	}

	bt.insertNonFull(bt.root, key, pageID)
	bt.size++
	return nil
}

func (bt *BTree) insertNonFull(node *Node, key string, pageID uint32) {
	i := len(node.keys) - 1

	if node.isLeaf {
		node.keys = append(node.keys, "")
		node.values = append(node.values, 0)

		for i >= 0 && key < node.keys[i] {
			node.keys[i+1] = node.keys[i]
			node.values[i+1] = node.values[i]
			i--
		}

		node.keys[i+1] = key
		node.values[i+1] = pageID
		return
	}

	for i >= 0 && key < node.keys[i] {
		i--
	}
	i++

	if len(node.children[i].keys) == bt.maxKeys() {
		bt.splitChild(node, i)
		if key > node.keys[i] {
			i++
		}
	}

	bt.insertNonFull(node.children[i], key, pageID)
}

func (bt *BTree) splitChild(parent *Node, childIndex int) {
	fullChild := parent.children[childIndex]
	newSibling := &Node{isLeaf: fullChild.isLeaf}

	t := bt.minDegree
	midKey := fullChild.keys[t-1]
	midValue := fullChild.values[t-1]

	newSibling.keys = append(newSibling.keys, fullChild.keys[t:]...)
	newSibling.values = append(newSibling.values, fullChild.values[t:]...)

	fullChild.keys = fullChild.keys[:t-1]
	fullChild.values = fullChild.values[:t-1]

	if !fullChild.isLeaf {
		newSibling.children = append(newSibling.children, fullChild.children[t:]...)
		fullChild.children = fullChild.children[:t]
	}

	parent.keys = append(parent.keys, "")
	parent.values = append(parent.values, 0)
	copy(parent.keys[childIndex+1:], parent.keys[childIndex:])
	copy(parent.values[childIndex+1:], parent.values[childIndex:])
	parent.keys[childIndex] = midKey
	parent.values[childIndex] = midValue

	parent.children = append(parent.children, nil)
	copy(parent.children[childIndex+2:], parent.children[childIndex+1:])
	parent.children[childIndex+1] = newSibling
}

// Delete removes a key from the B-Tree.
func (bt *BTree) Delete(key string) error {
	if bt.root == nil {
		return nil
	}

	if !bt.deleteFromNode(bt.root, key) {
		return nil
	}

	bt.size--

	if len(bt.root.keys) == 0 {
		if bt.root.isLeaf {
			bt.root = nil
		} else {
			bt.root = bt.root.children[0]
		}
	}

	return nil
}

func (bt *BTree) deleteFromNode(node *Node, key string) bool {
	idx := bt.findKey(node, key)

	if idx < len(node.keys) && node.keys[idx] == key {
		if node.isLeaf {
			bt.removeFromLeaf(node, idx)
			return true
		}
		return bt.removeFromInternal(node, idx)
	}

	if node.isLeaf {
		return false
	}

	childIndex := idx
	if len(node.children[childIndex].keys) < bt.minDegree {
		bt.fill(node, childIndex)
		if childIndex > len(node.keys) {
			childIndex = len(node.keys)
		}
	}

	return bt.deleteFromNode(node.children[childIndex], key)
}

func (bt *BTree) removeFromLeaf(node *Node, idx int) {
	node.keys = append(node.keys[:idx], node.keys[idx+1:]...)
	node.values = append(node.values[:idx], node.values[idx+1:]...)
}

func (bt *BTree) removeFromInternal(node *Node, idx int) bool {
	key := node.keys[idx]

	if len(node.children[idx].keys) >= bt.minDegree {
		predKey, predVal := bt.getPredecessor(node.children[idx])
		node.keys[idx] = predKey
		node.values[idx] = predVal
		return bt.deleteFromNode(node.children[idx], predKey)
	}

	if len(node.children[idx+1].keys) >= bt.minDegree {
		succKey, succVal := bt.getSuccessor(node.children[idx+1])
		node.keys[idx] = succKey
		node.values[idx] = succVal
		return bt.deleteFromNode(node.children[idx+1], succKey)
	}

	bt.merge(node, idx)
	return bt.deleteFromNode(node.children[idx], key)
}

func (bt *BTree) getPredecessor(node *Node) (string, uint32) {
	current := node
	for !current.isLeaf {
		current = current.children[len(current.children)-1]
	}
	last := len(current.keys) - 1
	return current.keys[last], current.values[last]
}

func (bt *BTree) getSuccessor(node *Node) (string, uint32) {
	current := node
	for !current.isLeaf {
		current = current.children[0]
	}
	return current.keys[0], current.values[0]
}

func (bt *BTree) fill(node *Node, idx int) {
	if idx > 0 && len(node.children[idx-1].keys) >= bt.minDegree {
		bt.borrowFromPrev(node, idx)
		return
	}

	if idx < len(node.children)-1 && len(node.children[idx+1].keys) >= bt.minDegree {
		bt.borrowFromNext(node, idx)
		return
	}

	if idx < len(node.keys) {
		bt.merge(node, idx)
	} else {
		bt.merge(node, idx-1)
	}
}

func (bt *BTree) borrowFromPrev(node *Node, idx int) {
	child := node.children[idx]
	sibling := node.children[idx-1]

	child.keys = append([]string{node.keys[idx-1]}, child.keys...)
	child.values = append([]uint32{node.values[idx-1]}, child.values...)

	if !child.isLeaf {
		moved := sibling.children[len(sibling.children)-1]
		sibling.children = sibling.children[:len(sibling.children)-1]
		child.children = append([]*Node{moved}, child.children...)
	}

	last := len(sibling.keys) - 1
	node.keys[idx-1] = sibling.keys[last]
	node.values[idx-1] = sibling.values[last]
	sibling.keys = sibling.keys[:last]
	sibling.values = sibling.values[:last]
}

func (bt *BTree) borrowFromNext(node *Node, idx int) {
	child := node.children[idx]
	sibling := node.children[idx+1]

	child.keys = append(child.keys, node.keys[idx])
	child.values = append(child.values, node.values[idx])

	node.keys[idx] = sibling.keys[0]
	node.values[idx] = sibling.values[0]

	sibling.keys = sibling.keys[1:]
	sibling.values = sibling.values[1:]

	if !child.isLeaf {
		moved := sibling.children[0]
		sibling.children = sibling.children[1:]
		child.children = append(child.children, moved)
	}
}

func (bt *BTree) merge(node *Node, idx int) {
	child := node.children[idx]
	sibling := node.children[idx+1]

	child.keys = append(child.keys, node.keys[idx])
	child.values = append(child.values, node.values[idx])

	child.keys = append(child.keys, sibling.keys...)
	child.values = append(child.values, sibling.values...)

	if !child.isLeaf {
		child.children = append(child.children, sibling.children...)
	}

	node.keys = append(node.keys[:idx], node.keys[idx+1:]...)
	node.values = append(node.values[:idx], node.values[idx+1:]...)
	node.children = append(node.children[:idx+1], node.children[idx+2:]...)
}

func (bt *BTree) updateIfExists(node *Node, key string, value uint32) bool {
	idx := bt.findKey(node, key)
	if idx < len(node.keys) && node.keys[idx] == key {
		node.values[idx] = value
		return true
	}
	if node.isLeaf {
		return false
	}
	return bt.updateIfExists(node.children[idx], key, value)
}

func (bt *BTree) findKey(node *Node, key string) int {
	i := 0
	for i < len(node.keys) && node.keys[i] < key {
		i++
	}
	return i
}

func (bt *BTree) maxKeys() int {
	return bt.order - 1
}

// Size returns the number of keys in the tree.
func (bt *BTree) Size() int {
	return bt.size
}
