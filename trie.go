// Package gtrie is an implementation of an R-Way Trie value structure.
// This package supports more useful functions based on
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

// trieNode for the node structure of the R-Way Trie
type trieNode struct {
	rval      rune
	path      string
	term      bool
	depth     int
	value     interface{}
	mask      uint64
	parent    *trieNode
	children  map[rune]*trieNode
	termCount int
}

// Trie for R-Way Trie
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

// Size returns the number of nodes inserted to the trie.
func (t *Trie) Size() int {
	return t.size
}

// Add adds a key to the Trie, including a value. The value
// is stored as `interface{}` and must be type cast by the caller.
// Upon the Add(), the old value added with the same key is removed from the trie.
func (t *Trie) Add(key string, value interface{}) {
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
	node = node.newChild(nul, key, 0, value, true)
}

// Find finds the value of the key matching to the input `key` exactly.
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
	return node.value, true
}

// Remove removes the `key` from the trie and return the value,
// ensuring that all bitmasks up to root are appropriately recalculated.
func (t *Trie) Remove(key string) interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	var (
		i     int
		r     rune
		value interface{}
		rs    = []rune(key)
		node  = findNode(t.root, []rune(key))
	)
	if node == nil {
		return nil
	}
	target, ok := node.children[nul]
	if !ok || !target.term {
		return nil
	}
	value = target.value
	target.children = nil
	target.parent = nil
	target.value = nil
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
			node.value = nil
			node.children = nil
		}
		// fmt.Printf("key %s, parent.rval %c n.rval %c r %c\n", target.path, n.parent.rval, n.rval, r)
		node = parent
	}
	node.termCount--
	updateMask(node)
	return value
}

// Clear removes all the keys and values of the trie.
func (t *Trie) Clear() {
	t.mu.Lock()
	node := t.root
	for r, c := range node.children {
		delete(node.children, r)
		removeAll(c)
	}
	node.rval = 0
	node.path = ""
	node.term = false
	node.depth = 0
	node.value = nil
	node.mask = uint64(0)
	node.parent = nil
	node.termCount = 0
	t.mu.Unlock()

	// keys := t.FindByPrefix("")
	// for i := range keys {
	// 	t.Remove(keys[i])
	// }
}

// FindByFuzzy performs a fuzzy search (Approximate string matching) against the keys in the trie.
func (t *Trie) FindByFuzzy(key string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	keys := fuzzycollect(t.root, []rune(key))
	sort.Sort(byKeys(keys))
	return keys
}

// FindByFuzzyValue performs a fuzzy search (Approximate string matching) against the keys in the trie.
func (t *Trie) FindByFuzzyValue(key string) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	values := fuzzycollectValues(t.root, []rune(key))
	return values
}

// FindByFuzzyAll performs a fuzzy search (Approximate string matching) against the keys in the trie.
func (t *Trie) FindByFuzzyAll(key string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return fuzzycollectAll(t.root, []rune(key))
}

// FindByPrefix performs a prefix search against the keys in the trie.
// It returns all the keys starting with `prefix` in the trie.
func (t *Trie) FindByPrefix(prefix string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(prefix))
	if node == nil {
		return nil
	}
	return collect(node)
}

// FindByPrefixValue returns all the values that have a key starting with `prefix`.
func (t *Trie) FindByPrefixValue(prefix string) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(prefix))
	if node == nil {
		return nil
	}
	return collectValues(node)
}

// FindByPrefixAll returns all the keys and values starting with `prefix`.
func (t *Trie) FindByPrefixAll(prefix string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(prefix))
	if node == nil {
		return nil
	}
	return collectAll(node)
}

// HasPrefix returns true if any of the keys in the trie starts with `prefix`.
func (t *Trie) HasPrefix(prefix string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(prefix))
	return node != nil
}

// Keys returns all the keys.
func (t *Trie) Keys() []string {
	return t.FindByPrefix("")
}

// Values returns all the values.
func (t *Trie) Values() []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(""))
	if node == nil {
		return nil
	}
	return collectValues(node)
}

// All returns a map for all matched keys and values.
// All the key of the map starts with `prefix`.
func (t *Trie) All() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := findNode(t.root, []rune(""))
	if node == nil {
		return nil
	}
	return collectAll(node)
}

// FindLongestMatchingPrefix finds a prefix key matching longestly with `key`
// from the trie and then returns the its key and inserted value.
// the key found is the longest matched prefix of the input `key`.
func (t *Trie) FindLongestMatchingPrefix(key string) (string, interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var found *trieNode
	node := t.root
	if node == nil {
		return "", nil, false
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
		return "", nil, false
	}
	return found.path, found.value, true
}

// FindMatchingPrefix finds all the matching prefixes against to the input `key`.
// The keys returned are the prefixes of the input `key`.
func (t *Trie) FindMatchingPrefix(key string) ([]string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	nodes, ok := t.findPrefixMatchNodes(key)
	if ok {
		keys := make([]string, 0, len(nodes))
		for _, n := range nodes {
			keys = append(keys, n.path)
		}
		return keys, true
	}
	return nil, false
}

// FindMatchingPrefixValue finds all the matched prefix keys against to the input `key`.
// The values of the matched keys are returned.
func (t *Trie) FindMatchingPrefixValue(key string) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	nodes, ok := t.findPrefixMatchNodes(key)
	if ok {
		vals := make([]interface{}, 0, len(nodes))
		for _, n := range nodes {
			vals = append(vals, n.value)
		}
		return vals
	}
	return nil
}

// FindMatchingPrefixAll finds all the matched prefix keys and the values against to
// the input `key`. The keys returned are the prefixes of the input `key`.
func (t *Trie) FindMatchingPrefixAll(key string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m := make(map[string]interface{})
	nodes, ok := t.findPrefixMatchNodes(key)
	if ok {
		for _, n := range nodes {
			m[n.path] = n.value
		}
	}
	return m
}

// FindAll finds all relative prefix keys against to the input `key` and
// all matched keys that starts with the input `key` in the trie.
// It returns the result of (FindByPrefixAll() + FindMatchingPrefixAll())
func (t *Trie) FindAll(key string) map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m := make(map[string]interface{})
	node := findNode(t.root, []rune(key))
	if node != nil {
		m = collectAll(node)
	}
	nodes, ok := t.findPrefixMatchNodes(key)
	if ok {
		for _, n := range nodes {
			m[n.path] = n.value
		}
	}
	return m
}

// findPrefixMatchNodes finds all matched nodes in the trie.
// The key of each node is a prefix of the input `key`.
func (t *Trie) findPrefixMatchNodes(key string) ([]*trieNode, bool) {
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

// Creates and returns a pointer to a new child for the node.
func (n *trieNode) newChild(rval rune, path string, bitmask uint64, value interface{}, term bool) *trieNode {
	node := &trieNode{
		rval:     rval,
		path:     path,
		mask:     bitmask,
		term:     term,
		value:    value,
		parent:   n,
		children: make(map[rune]*trieNode),
		depth:    n.depth + 1,
	}
	n.children[node.rval] = node
	n.mask |= bitmask
	return node
}

// removeChild removes the child
func (n *trieNode) removeChild(r rune) {
	delete(n.children, r)
	updateMask(n.parent)
	// for nd := n.parent; nd != nil; nd = nd.parent {
	// 	nd.mask ^= nd.mask
	// 	nd.mask |= uint64(1) << uint64(nd.rval-'a')
	// 	for _, c := range nd.children {
	// 		nd.mask |= c.mask
	// 	}
	// }
}

// updateMask updates n.mask
func updateMask(node *trieNode) {
	for ; node != nil; node = node.parent {
		node.mask ^= node.mask
		node.mask |= uint64(1) << uint64(node.rval-'a')
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
	node.value = nil
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
			values = append(values, n.value)
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
			m[word] = n.value
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

		if p.node.rval == partial[p.idx] {
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

func fuzzycollectValues(node *trieNode, partial []rune) []interface{} {
	if len(partial) == 0 {
		return collectValues(node)
	}

	var (
		m      uint64
		i      int
		p      potentialSubtree
		values []interface{}
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

		if p.node.rval == partial[p.idx] {
			p.idx++
			if p.idx == len(partial) {
				values = append(values, collectValues(p.node)...)
				continue
			}
		}

		for _, c := range p.node.children {
			potential = append(potential, potentialSubtree{node: c, idx: p.idx})
		}
	}
	return values
}

func fuzzycollectAll(node *trieNode, partial []rune) map[string]interface{} {
	if len(partial) == 0 {
		return collectAll(node)
	}

	var (
		m      uint64
		i      int
		p      potentialSubtree
		values map[string]interface{} = make(map[string]interface{})
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

		if p.node.rval == partial[p.idx] {
			p.idx++
			if p.idx == len(partial) {
				for k, v := range collectAll(p.node) {
					values[k] = v
				}
				continue
			}
		}

		for _, c := range p.node.children {
			potential = append(potential, potentialSubtree{node: c, idx: p.idx})
		}
	}
	return values
}
