package index

// BTree implements a B-Tree index structure
type BTree struct {
	root  *Node
	order int // maximum number of children per node
	size  int
}

// Node represents a B-Tree node
type Node struct {
	keys     []string
	values   []uint32 // page IDs
	children []*Node
	isLeaf   bool
}

// NewBTree creates a new B-Tree with the given order
func NewBTree(order int) *BTree {
	return &BTree{
		root:  nil,
		order: order,
		size:  0,
	}
}

// Search finds a key and returns its page ID
func (bt *BTree) Search(key string) (uint32, bool) {
	if bt.root == nil {
		return 0, false
	}
	return bt.searchNode(bt.root, key)
}

func (bt *BTree) searchNode(node *Node, key string) (uint32, bool) {
	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}

	if i < len(node.keys) && key == node.keys[i] {
		return node.values[i], true
	}

	if node.isLeaf {
		return 0, false
	}

	return bt.searchNode(node.children[i], key)
}

// Insert adds a key-value pair to the B-Tree
func (bt *BTree) Insert(key string, pageID uint32) error {
	if bt.root == nil {
		bt.root = &Node{
			keys:   []string{key},
			values: []uint32{pageID},
			isLeaf: true,
		}
		bt.size++
		return nil
	}

	// Simple insert for leaf-only tree (MVP)
	node := bt.root
	for !node.isLeaf {
		i := 0
		for i < len(node.keys) && key > node.keys[i] {
			i++
		}
		node = node.children[i]
	}

	// Insert into leaf
	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}

	// Check for duplicate
	if i < len(node.keys) && node.keys[i] == key {
		node.values[i] = pageID // update existing
		return nil
	}

	// Insert at position i
	node.keys = append(node.keys, "")
	node.values = append(node.values, 0)
	copy(node.keys[i+1:], node.keys[i:])
	copy(node.values[i+1:], node.values[i:])
	node.keys[i] = key
	node.values[i] = pageID
	bt.size++

	return nil
}

// Delete removes a key from the B-Tree
func (bt *BTree) Delete(key string) error {
	if bt.root == nil {
		return nil
	}

	node := bt.root
	for !node.isLeaf {
		i := 0
		for i < len(node.keys) && key > node.keys[i] {
			i++
		}
		node = node.children[i]
	}

	for i, k := range node.keys {
		if k == key {
			node.keys = append(node.keys[:i], node.keys[i+1:]...)
			node.values = append(node.values[:i], node.values[i+1:]...)
			bt.size--
			return nil
		}
	}

	return nil
}

// Size returns the number of keys in the tree
func (bt *BTree) Size() int {
	return bt.size
}
