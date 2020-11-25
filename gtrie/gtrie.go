// Package gtrie is defined for the convenient use of the trie
// [defined function]
//  - func (gt *Trie) Add(key string, value interface{})
//  - func (gt *Trie) Find(key string) (interface{}, bool)
//  - func (gt *Trie) FindLongestMatch(key string) (string, interface{}, bool)
//  - func (gt *Trie) HasKeysWithPrefix(key string) bool
//  - func (gt *Trie) Keys() []string
//  - func (gt *Trie) FuzzySearch(pre string) []string
//  - func (gt *Trie) PrefixSearch(pre string) []string
package gtrie

import (
	"github.com/neoul/trie"
)

// Trie = trie.Trie
type Trie trie.Trie

// New creates a new Trie with an initialized root Node.
func New() *Trie {
	trie := trie.New()
	return (*Trie)(trie)
}

// Add adds a new key with a value
func (gt *Trie) Add(key string, value interface{}) {
	t := (*trie.Trie)(gt)
	t.Add(key, value)
}

// Find finds the value associated with the key.
func (gt *Trie) Find(key string) (interface{}, bool) {
	t := (*trie.Trie)(gt)
	node, ok := t.Find(key)
	if !ok {
		return nil, false
	}
	return node.Meta(), ok
}

// FindLongestMatch finds a longest matched key in the trie and
// returns a matched key, inserted value.
func (gt *Trie) FindLongestMatch(key string) (string, interface{}, bool) {
	t := (*trie.Trie)(gt)
	node, ok := t.FindLongestMatchedNode(key)
	if !ok {
		return "", nil, false
	}
	return node.Key(), node.Meta(), true
}
