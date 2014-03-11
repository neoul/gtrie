package trie

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
)

// Implementation of an R-Way Trie data structure.
//
// A Trie has a root Node which is the base of the tree.
// Each subsequent Node has a letter and children, which are
// nodes that have letter values associated with them.

type Node struct {
	val       int
	mask      uint64
	children  map[rune]*Node
	childvals []rune
}

type Trie struct {
	root *Node
	size int
}

const nul = 0x0

func newNode(val int, m uint64) *Node {
	return &Node{
		val:      val,
		mask:     m,
		children: make(map[rune]*Node),
	}
}

func (n *Node) NewChild(r rune, bitmask uint64, val int) *Node {
	node := newNode(val, bitmask)
	n.children[r] = node
	n.childvals = append(n.childvals, r)
	return node
}

// Returns the children of this node.
func (n Node) Children() map[rune]*Node {
	return n.children
}

func (n Node) ChildVals() []rune {
	return n.childvals
}

func (n Node) Val() int {
	return n.val
}

func (n Node) Mask() uint64 {
	return n.mask
}

// Creates a new Trie with an initialized root Node.
func CreateTrie() *Trie {
	node := newNode(0, 0)
	t := &Trie{
		root: node,
		size: 0,
	}
	return t
}

// Returns the root node for the Trie.
func (t *Trie) Root() *Node {
	return t.root
}

// Adds the key to the Trie.
func (t *Trie) AddKey(key string) int {
	t.size++
	runes := []rune(key)
	return t.addrune(t.Root(), runes, 0)
}

func (t *Trie) AddKeysFromFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		line = strings.TrimSuffix(line, "\n")
		t.AddKey(line)
	}
}

func (t *Trie) Keys() []string {
	return t.KeysWithPrefix("")
}

func (t Trie) FuzzySearch(pre string) []string {
	var (
		keys []string
		pm   = make([]rune, 35)
	)

	fuzzycollect(t.Root(), pm, []rune(pre), &keys)
	return keys
}

func (t Trie) KeysWithPrefix(pre string) []string {
	var keys []string

	node := t.nodeAtPath(pre)
	if node == nil {
		return keys
	}

	collect(node, []rune(pre), &keys)
	return keys
}

func (t Trie) nodeAtPath(pre string) *Node {
	runes := []rune(pre)
	return findNode(t.Root(), runes, 0)
}

func findNode(node *Node, runes []rune, d int) *Node {
	if node == nil {
		return nil
	}

	if len(runes) == 0 {
		return node
	}

	upper := len(runes)
	if d == upper {
		return node
	}

	n, ok := node.Children()[runes[d]]
	if !ok {
		return nil
	}

	d++
	return findNode(n, runes, d)
}

func (t Trie) addrune(node *Node, runes []rune, i int) int {
	if len(runes) == 0 {
		node.NewChild(nul, 0, t.size)
		return i
	}

	r := runes[0]
	c := node.Children()

	n, ok := c[r]
	bitmask := maskruneslice(runes)
	if !ok {
		n = node.NewChild(r, bitmask, 0)
	}
	n.mask |= bitmask

	i++
	return t.addrune(n, runes[1:], i)
}

func maskruneslice(rs []rune) uint64 {
	var m uint64
	i := uint64(1)
	for _, r := range rs {
		h := i << (uint64(r) - 97)
		m |= h
	}

	return m
}

func maskrune(r rune) uint64 {
	var m uint64
	i := uint64(1)
	h := i << (uint64(r) - 97)
	m |= h

	return m
}

func collect(node *Node, pre []rune, keys *[]string) {
	children := node.Children()
	for _, r := range node.ChildVals() {
		n := children[r]
		if n.Val() > 0 {
			*keys = append(*keys, string(pre))
			continue
		}

		npre := append(pre, r)
		collect(n, npre, keys)
	}
}

func fuzzycollect(node *Node, partialmatch, partial []rune, keys *[]string) {
	partiallen := len(partial)

	if partiallen == 0 {
		collect(node, partialmatch, keys)
		return
	}

	m := maskruneslice(partial)
	children := node.Children()
	for _, v := range node.ChildVals() {
		n := children[v]

		xor := n.Mask() ^ m
		if (xor & m) != 0 {
			continue
		}

		npartial := partial
		if v == partial[0] {
			if partiallen > 1 {
				npartial = partial[1:]
			} else {
				npartial = partial[0:0]
			}
		}

		fuzzycollect(n, append(partialmatch, v), npartial, keys)
	}
}
