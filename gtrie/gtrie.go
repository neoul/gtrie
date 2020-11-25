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
type Trie struct {
	*trie.Trie
}

// New creates a new Trie with an initialized root Node.
func New() *Trie {
	trie := trie.New()
	if trie == nil {
		return nil
	}
	return &Trie{Trie: trie}
}

// Add adds a new key with a value
func (t *Trie) Add(key string, value interface{}) {
	t.Trie.Add(key, value)
}

// Find finds the value associated with the key.
func (t *Trie) Find(key string) (interface{}, bool) {
	node, ok := t.Trie.Find(key)
	if !ok {
		return nil, false
	}
	return node.Meta(), ok
}

// FindLongestMatch finds a longest matched key in the trie and
// returns a matched key, inserted value.
func (t *Trie) FindLongestMatch(key string) (string, interface{}, bool) {
	node, ok := t.Trie.FindLongestMatchedNode(key)
	if !ok {
		return "", nil, false
	}
	return node.Key(), node.Meta(), true
}
