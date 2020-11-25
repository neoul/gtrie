// Package gtrie is defined for the convenient use of the trie
// [defined function]
//  - func (gt *GTrie) Add(key string, value interface{})
//  - func (gt *GTrie) Find(key string) (interface{}, bool)
//  - func (gt *GTrie) FindLongestMatch(key string) (string, interface{}, bool)
//  - func (gt *GTrie) HasKeysWithPrefix(key string) bool
//  - func (gt *GTrie) Keys() []string
//  - func (gt *GTrie) FuzzySearch(pre string) []string
//  - func (gt *GTrie) PrefixSearch(pre string) []string
package gtrie

import (
	"github.com/neoul/trie"
)

// GTrie = trie.Trie
type GTrie trie.Trie

// New creates a new Trie with an initialized root Node.
func New() *GTrie {
	trie := trie.New()
	return (*GTrie)(trie)
}

// Add adds a new key with a value
func (gt *GTrie) Add(key string, value interface{}) {
	t := (*trie.Trie)(gt)
	t.Add(key, value)
}

// Find finds the value associated with the key.
func (gt *GTrie) Find(key string) (interface{}, bool) {
	t := (*trie.Trie)(gt)
	node, ok := t.Find(key)
	if !ok {
		return nil, false
	}
	return node.Meta(), ok
}

// FindLongestMatch finds a longest matched key in the trie and
// returns a matched key, inserted value.
func (gt *GTrie) FindLongestMatch(key string) (string, interface{}, bool) {
	t := (*trie.Trie)(gt)
	node, ok := t.FindLongestMatchedNode(key)
	if !ok {
		return "", nil, false
	}
	return node.Key(), node.Meta(), true
}
