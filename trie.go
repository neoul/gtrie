// Package gtrie is an implementation of an R-Way Trie data structure.
// This package supports more useful functions for the trie based on
// derekparker/trie (https://godoc.org/github.com/derekparker/trie).
//
// A Trie has a root trieNode which is the base of the tree.
// Each subsequent trieNode has a letter and children, which are
// nodes that have letter values associated with them.
package gtrie

import (
	"sort"
	"sync"
)

// node structure of the R-Way Trie
type trieNode struct {
	val       rune
	path      string
	term      bool
	depth     int
	data      interface{}
	mask      uint64
	parent    *trieNode
	children  map[rune]*trieNode
	termCount int
}

// Trie for control
type Trie struct {
	mu   sync.RWMutex
	root *trieNode
	size int
}

// byKeys for fuzzy search
type byKeys []string

func (a byKeys) Len() int           { return len(a) }
func (a byKeys) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byKeys) Less(i, j int) bool { return len(a[i]) < len(a[j]) }

const nul = 0x0

// New creates a new Trie with an initialized root trieNode.
func New() *Trie {
	return &Trie{
		root: &trieNode{children: make(map[rune]*trieNode), depth: 0},
		size: 0,
	}
}

// Size returns the size of the trie.
func (t *Trie) Size() int {
	return t.size
}

// Add adds the key to the Trie, including a data. The data
// is stored as `interface{}` and must be type cast by
// the caller.
func (t *Trie) Add(key string, data interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cnt := 1
	runes := []rune(key)
	// check the node exists
	if node := findNode(t.root, runes); node != nil {
		if node, ok := node.children[nul]; ok && node.term {
			cnt = 0
		}
	}

	t.size = t.size + cnt
	bitmask := maskruneslice(runes)
	node := t.root
	node.mask |= bitmask
	node.termCount = node.termCount + cnt
	for i := range runes {
		r := runes[i]
		bitmask = maskruneslice(runes[i:])
		if n, ok := node.children[r]; ok {
			node = n
			node.mask |= bitmask
		} else {
			node = node.newChild(r, "", bitmask, nil, false)
		}
		node.termCount = node.termCount + cnt
	}
	node = node.newChild(nul, key, 0, data, true)
}

// Find finds and returns a data associated with `key`.
func (t *Trie) Find(key string) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(key))
	if node == nil {
		return nil, false
	}

	node, ok := node.children[nul]
	if !ok || !node.term {
		return nil, false
	}

	return node.data, true
}

// HasPrefix returns true if there is a key started with `pre(fix)`.
func (t *Trie) HasPrefix(pre string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(pre))
	return node != nil
}

// Remove removes a key from the trie and return the data, ensuring that
// all bitmasks up to root are appropriately recalculated.
func (t *Trie) Remove(key string) interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	var (
		i    int
		r    rune
		data interface{}
		rs   = []rune(key)
		node = findNode(t.root, []rune(key))
	)
	if node == nil {
		return nil
	}
	target, ok := node.children[nul]
	if !ok || !target.term {
		return nil
	}
	data = target.data
	target.children = nil
	target.parent = nil
	target.data = nil
	t.size--
	node.removeChild(nul)
	for node.parent != nil {
		node.termCount--
		parent := node.parent
		if len(node.children) <= 0 {
			i++
			r = rs[len(rs)-i]
			parent.removeChild(r)
			node.parent = nil
			node.data = nil
			node.children = nil
		}
		// fmt.Printf("key %s, parent.val %c n.val %c r %c\n", target.path, n.parent.val, n.val, r)
		node = parent
	}
	node.termCount--
	updateMask(node)
	return data
}

// RemoveAll removes all internal data of the trie
func (t *Trie) RemoveAll() {
	t.mu.Lock()
	node := t.root
	for r, c := range node.children {
		delete(node.children, r)
		removeAll(c)
	}
	node.val = 0
	node.path = ""
	node.term = false
	node.depth = 0
	node.data = nil
	node.mask = uint64(0)
	node.parent = nil
	node.termCount = 0
	t.mu.Unlock()

	// keys := t.Keys("")
	// for i := range keys {
	// 	t.Remove(keys[i])
	// }
}

// Keys returns all the keys currently stored and started with `pre(fix)` in the trie.
func (t *Trie) Keys(pre string) []string {
	// RLock & RUnlock at PrefixSearch
	if t.size == 0 {
		return []string{}
	}

	return t.PrefixSearch(pre)
}

// Values returns all values that have a key started with `pre(fix)`.
func (t *Trie) Values(pre string) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(pre))
	if node == nil {
		return nil
	}

	return collectValues(node)
}

// All returns a map for all matched keys and values with `pre(fix)`.
func (t *Trie) All(pre string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(pre))
	if node == nil {
		return nil
	}

	return collectAll(node)
}

// FuzzySearch performs a fuzzy search against the keys in the trie.
func (t *Trie) FuzzySearch(pre string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	keys := fuzzycollect(t.root, []rune(pre))
	sort.Sort(byKeys(keys))
	return keys
}

// PrefixSearch performs a prefix search against the keys in the trie.
func (t *Trie) PrefixSearch(pre string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(pre))
	if node == nil {
		return nil
	}

	return collect(node)
}

// findLongestMatchedNode finds a longest matched key with `key` in the trie
func (t *Trie) findLongestMatchedNode(key string) (*trieNode, bool) {
	var found *trieNode
	node := t.root

	if node == nil {
		return nil, false
	}

	for _, r := range []rune(key) {
		n, ok := node.children[r]
		if !ok {
			break
		}
		t, ok := n.children[nul]
		if ok && t.term {
			found = t
		}
		node = n
	}
	if found == nil {
		return nil, false
	}
	return found, true
}

// FindLongestMatchedkey finds a longest matched key with `key` in the trie
func (t *Trie) FindLongestMatchedkey(key string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node, ok := t.findLongestMatchedNode(key)
	if ok {
		return node.path, true
	}
	return "", false
}

// FindLongestMatchedPrefix finds a longest matched key with the input `key` in the trie and
// returns a matched key (that is a prefix of `key`) and a inserted data.
func (t *Trie) FindLongestMatchedPrefix(key string) (string, interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node, ok := t.findLongestMatchedNode(key)
	if !ok {
		return "", nil, false
	}
	return node.path, node.data, true
}

// FindPrefix finds all matched keys as the prefix against to the input `key` in the trie.
func (t *Trie) FindPrefix(key string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m := make(map[string]interface{})
	nodes, ok := t.findMatchedNodes(key)
	if ok {
		for _, n := range nodes {
			m[n.path] = n.data
		}
	}
	return m
}

// findMatchedNodes finds all matched nodes in the trie.
// The key of each node is a prefix of the input `key`.
func (t *Trie) findMatchedNodes(key string) ([]*trieNode, bool) {
	found := false
	node := t.root
	if node == nil {
		return nil, false
	}
	if t.size <= 0 {
		return nil, false
	}
	nodes := make([]*trieNode, 0, t.size)
	for _, r := range []rune(key) {
		n, ok := node.children[r]
		if !ok {
			break
		}
		t, ok := n.children[nul]
		if ok && t.term {
			nodes = append(nodes, t)
			found = true
		}
		node = n
	}
	if found {
		return nodes, true
	}
	return nil, false
}

// FindMatchedKey finds all matched prefix keys against to the input `key` in the trie.
func (t *Trie) FindMatchedKey(key string) ([]string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	nodes, ok := t.findMatchedNodes(key)
	if ok {
		keys := make([]string, 0, len(nodes))
		for _, n := range nodes {
			keys = append(keys, n.path)
		}
		return keys, true
	}
	return nil, false
}

// FindAll finds all relative prefix keys against to the input `key` and
// all matched keys that starts with the input `key`.
func (t *Trie) FindAll(key string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m := make(map[string]interface{})
	node := findNode(t.root, []rune(key))
	if node != nil {
		m = collectAll(node)
	}
	nodes, ok := t.findMatchedNodes(key)
	if ok {
		for _, n := range nodes {
			m[n.path] = n.data
		}
	}
	return m
}

// Creates and returns a pointer to a new child for the node.
func (n *trieNode) newChild(val rune, path string, bitmask uint64, data interface{}, term bool) *trieNode {
	node := &trieNode{
		val:      val,
		path:     path,
		mask:     bitmask,
		term:     term,
		data:     data,
		parent:   n,
		children: make(map[rune]*trieNode),
		depth:    n.depth + 1,
	}
	n.children[node.val] = node
	n.mask |= bitmask
	return node
}

// removeChild removes the child
func (n *trieNode) removeChild(r rune) {
	delete(n.children, r)
	updateMask(n.parent)
	// for nd := n.parent; nd != nil; nd = nd.parent {
	// 	nd.mask ^= nd.mask
	// 	nd.mask |= uint64(1) << uint64(nd.val-'a')
	// 	for _, c := range nd.children {
	// 		nd.mask |= c.mask
	// 	}
	// }
}

// updateMask updates n.mask
func updateMask(node *trieNode) {
	for ; node != nil; node = node.parent {
		node.mask ^= node.mask
		node.mask |= uint64(1) << uint64(node.val-'a')
		for _, c := range node.children {
			node.mask |= c.mask
		}
	}
}

func removeAll(node *trieNode) {
	for r, c := range node.children {
		delete(node.children, r)
		removeAll(c)
	}
	node.parent = nil
	node.children = nil
	node.data = nil
}

func findNode(node *trieNode, runes []rune) *trieNode {
	if node == nil {
		return nil
	}

	if len(runes) == 0 {
		return node
	}

	n, ok := node.children[runes[0]]
	if !ok {
		return nil
	}

	var nrunes []rune
	if len(runes) > 1 {
		nrunes = runes[1:]
	} else {
		nrunes = runes[0:0]
	}

	return findNode(n, nrunes)
}

func maskruneslice(rs []rune) uint64 {
	var m uint64
	for _, r := range rs {
		m |= uint64(1) << uint64(r-'a')
	}
	return m
}

func collect(node *trieNode) []string {
	var (
		n *trieNode
		i int
	)
	keys := make([]string, 0, node.termCount)
	nodes := make([]*trieNode, 1, len(node.children)+1)
	nodes[0] = node
	for l := len(nodes); l != 0; l = len(nodes) {
		i = l - 1
		n = nodes[i]
		nodes = nodes[:i]
		for _, c := range n.children {
			nodes = append(nodes, c)
		}
		if n.term {
			word := n.path
			keys = append(keys, word)
		}
	}
	return keys
}

func collectValues(node *trieNode) []interface{} {
	var (
		n *trieNode
		i int
	)
	values := make([]interface{}, 0, node.termCount)
	nodes := make([]*trieNode, 1, len(node.children)+1)
	nodes[0] = node
	for l := len(nodes); l != 0; l = len(nodes) {
		i = l - 1
		n = nodes[i]
		nodes = nodes[:i]
		for _, c := range n.children {
			nodes = append(nodes, c)
		}
		if n.term {
			values = append(values, n.data)
		}
	}
	return values
}

func collectAll(node *trieNode) map[string]interface{} {
	var (
		n *trieNode
		i int
	)
	m := make(map[string]interface{})
	nodes := make([]*trieNode, 1, len(node.children)+1)
	nodes[0] = node
	for l := len(nodes); l != 0; l = len(nodes) {
		i = l - 1
		n = nodes[i]
		nodes = nodes[:i]
		for _, c := range n.children {
			nodes = append(nodes, c)
		}
		if n.term {
			word := n.path
			m[word] = n.data
		}
	}
	return m
}

type potentialSubtree struct {
	idx  int
	node *trieNode
}

func fuzzycollect(node *trieNode, partial []rune) []string {
	if len(partial) == 0 {
		return collect(node)
	}

	var (
		m    uint64
		i    int
		p    potentialSubtree
		keys []string
	)

	potential := []potentialSubtree{potentialSubtree{node: node, idx: 0}}
	for l := len(potential); l > 0; l = len(potential) {
		i = l - 1
		p = potential[i]
		potential = potential[:i]
		m = maskruneslice(partial[p.idx:])
		if (p.node.mask & m) != m {
			continue
		}

		if p.node.val == partial[p.idx] {
			p.idx++
			if p.idx == len(partial) {
				keys = append(keys, collect(p.node)...)
				continue
			}
		}

		for _, c := range p.node.children {
			potential = append(potential, potentialSubtree{node: c, idx: p.idx})
		}
	}
	return keys
}
