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
	meta      interface{}
	mask      uint64
	parent    *trieNode
	children  map[rune]*trieNode
	termCount int
}

// Trie for control
type Trie struct {
	mu   sync.Mutex
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

// Add adds the key to the Trie, including a value. The value
// is stored as `interface{}` and must be type cast by
// the caller.
func (t *Trie) Add(key string, value interface{}) {
	t.mu.Lock()

	t.size++
	runes := []rune(key)
	bitmask := maskruneslice(runes)
	node := t.root
	node.mask |= bitmask
	node.termCount++
	for i := range runes {
		r := runes[i]
		bitmask = maskruneslice(runes[i:])
		if n, ok := node.children[r]; ok {
			node = n
			node.mask |= bitmask
		} else {
			node = node.newChild(r, "", bitmask, nil, false)
		}
		node.termCount++
	}
	node = node.newChild(nul, key, 0, value, true)
	t.mu.Unlock()
}

// Find finds and returns a value associated with `key`.
func (t *Trie) Find(key string) (interface{}, bool) {
	node := findNode(t.root, []rune(key))
	if node == nil {
		return nil, false
	}

	node, ok := node.children[nul]
	if !ok || !node.term {
		return nil, false
	}

	return node.meta, true
}

// HasPrefix returns true if there is a key started with `pre(fix)`.
func (t *Trie) HasPrefix(pre string) bool {
	node := findNode(t.root, []rune(pre))
	return node != nil
}

// Remove removes a key from the trie, ensuring that
// all bitmasks up to root are appropriately recalculated.
func (t *Trie) Remove(key string) {
	var (
		i    int
		rs   = []rune(key)
		node = findNode(t.root, []rune(key))
	)
	t.mu.Lock()

	t.size--
	for n := node.parent; n != nil; n = n.parent {
		i++
		if len(n.children) > 1 {
			r := rs[len(rs)-i]
			n.removeChild(r)
			break
		}
	}
	t.mu.Unlock()
}

// Keys returns all the keys currently stored and started with `pre(fix)` in the trie.
func (t *Trie) Keys(pre string) []string {
	if t.size == 0 {
		return []string{}
	}

	return t.PrefixSearch(pre)
}

// Values returns all values that have a key started with `pre(fix)`.
func (t *Trie) Values(pre string) []interface{} {
	node := findNode(t.root, []rune(pre))
	if node == nil {
		return nil
	}

	return collectValues(node)
}

// All returns a map for all matched keys and values with `pre(fix)`.
func (t *Trie) All(pre string) map[string]interface{} {
	node := findNode(t.root, []rune(pre))
	if node == nil {
		return nil
	}

	return collectAll(node)
}

// FuzzySearch performs a fuzzy search against the keys in the trie.
func (t *Trie) FuzzySearch(pre string) []string {
	keys := fuzzycollect(t.root, []rune(pre))
	sort.Sort(byKeys(keys))
	return keys
}

// PrefixSearch performs a prefix search against the keys in the trie.
func (t *Trie) PrefixSearch(pre string) []string {
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
	node, ok := t.findLongestMatchedNode(key)
	if ok {
		return node.path, true
	}
	return "", false
}

// FindLongestMatchedPrefix finds a longest matched key with the input `key` in the trie and
// returns a matched key (that is a prefix of `key`) and a inserted value.
func (t *Trie) FindLongestMatchedPrefix(key string) (string, interface{}, bool) {
	node, ok := t.findLongestMatchedNode(key)
	if !ok {
		return "", nil, false
	}
	return node.path, node.meta, true
}

// FindPrefix finds all matched keys as the prefix against to the input `key` in the trie.
func (t *Trie) FindPrefix(key string) map[string]interface{} {
	m := make(map[string]interface{})
	nodes, ok := t.findMatchedNodes(key)
	if ok {
		for _, n := range nodes {
			m[n.path] = n.meta
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
	m := make(map[string]interface{})
	node := findNode(t.root, []rune(key))
	if node != nil {
		m = collectAll(node)
	}
	nodes, ok := t.findMatchedNodes(key)
	if ok {
		for _, n := range nodes {
			m[n.path] = n.meta
		}
	}
	return m
}

// Creates and returns a pointer to a new child for the node.
func (n *trieNode) newChild(val rune, path string, bitmask uint64, meta interface{}, term bool) *trieNode {
	node := &trieNode{
		val:      val,
		path:     path,
		mask:     bitmask,
		term:     term,
		meta:     meta,
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
	for nd := n.parent; nd != nil; nd = nd.parent {
		nd.mask ^= nd.mask
		nd.mask |= uint64(1) << uint64(nd.val-'a')
		for _, c := range nd.children {
			nd.mask |= c.mask
		}
	}
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
	// keys := make([]string, 0, node.termCount)
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
			values = append(values, n.meta)
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
	// keys := make([]string, 0, node.termCount)
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
			m[word] = n.meta
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
