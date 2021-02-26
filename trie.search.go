package gtrie

// SearchType of Search func
//  [SearchExactly, SearchByPrefix, SearchLongestMatchingPrefix, SearchMatcingPrefix, SearchApproximate]
type SearchType int

const (
	// SearchExactly - finds the key exactly matching to input `key`.
	SearchExactly = 0

	// SearchByPrefix - finds all matching keys that starts with the input `key`
	// The input `key` is the prefix of the keys found.
	SearchByPrefix SearchType = 1

	// SearchLongestMatchingPrefix - finds the longest matching prefix substring
	// against to intput `key` from the trie using Longest Prefix Match algorithm.
	SearchLongestMatchingPrefix SearchType = 2

	// SearchMatcingPrefix - finds all matching prefix keys with the intput `key`.
	SearchMatcingPrefix SearchType = 3

	// SearchApproximate (Fuzzy search: Approximate string matching) - finds all matched strings by fuzzy search.
	SearchApproximate SearchType = 4

	// SearchAllRelativeKey = SearchByPrefix + SearchMatcingPrefix + SearchApproximate
	SearchAllRelativeKey SearchType = 5
)

// Search finds all matching keys according to stype (SearchType).
func (t *Trie) Search(key string, stype SearchType) []string {
	switch stype {
	case SearchExactly:
		if _, ok := t.Find(key); ok {
			return []string{key}
		}
	case SearchByPrefix:
		return t.FindByPrefix(key)
	case SearchLongestMatchingPrefix:
		if k, _, ok := t.FindLongestMatchingPrefix(key); ok {
			return []string{k}
		}
	case SearchMatcingPrefix:
		if keys, ok := t.FindMatchingPrefix(key); ok {
			return keys
		}
	case SearchApproximate:
		return t.FindByFuzzy(key)
	case SearchAllRelativeKey:
		return t.FindRelative(key)
	}
	return nil
}

// SearchValues finds all matching keys according to stype (SearchType)
// and returns all the values of the matching keys.
func (t *Trie) SearchValues(key string, stype SearchType) []interface{} {
	switch stype {
	case SearchExactly:
		if v, ok := t.Find(key); ok {
			return []interface{}{v}
		}
	case SearchByPrefix:
		return t.FindByPrefixValue(key)
	case SearchLongestMatchingPrefix:
		if _, v, ok := t.FindLongestMatchingPrefix(key); ok {
			return []interface{}{v}
		}
	case SearchMatcingPrefix:
		return t.FindMatchingPrefixValue(key)
	case SearchApproximate:
		return t.FindByFuzzyValue(key)
	case SearchAllRelativeKey:
		return t.FindRelativeValues(key)
	}
	return nil
}

// SearchAll finds all matching keys and values according to stype (SearchType).
func (t *Trie) SearchAll(key string, stype SearchType) map[string]interface{} {
	switch stype {
	case SearchExactly:
		if v, ok := t.Find(key); ok {
			return map[string]interface{}{key: v}
		}
	case SearchByPrefix:
		return t.FindByPrefixAll(key)
	case SearchLongestMatchingPrefix:
		if k, v, ok := t.FindLongestMatchingPrefix(key); ok {
			return map[string]interface{}{k: v}
		}
	case SearchMatcingPrefix:
		return t.FindMatchingPrefixAll(key)
	case SearchApproximate:
		return t.FindByFuzzyAll(key)
	case SearchAllRelativeKey:
		return t.FindRelativeAll(key)
	}
	return nil
}

// FindRelative finds all relative keys against to the input `key`.
// It returns the result of (FindByPrefix + FindMatchingPrefix + FindByFuzzy)
func (t *Trie) FindRelative(key string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m := map[string]interface{}{}
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
	keys := t.FindByFuzzy(key)
	for i := range keys {
		m[keys[i]], _ = t.Find(key)
	}
	keys = make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// FindRelativeValues finds all relative values against to the input `key`.
// It returns the result of (FindByPrefix + FindMatchingPrefix + FindByFuzzy)
func (t *Trie) FindRelativeValues(key string) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m := map[string]interface{}{}
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
	keys := t.FindByFuzzy(key)
	for i := range keys {
		m[keys[i]], _ = t.Find(key)
	}
	values := make([]interface{}, 0, len(m))
	for k := range m {
		values = append(values, k)
	}
	return values
}

// FindRelativeAll finds all relative keys against to the input `key`.
// It returns the result of (FindByPrefix + FindMatchingPrefix + FindByFuzzy)
func (t *Trie) FindRelativeAll(key string) map[string]interface{} {
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
	fm := t.FindByFuzzyAll(key)
	for k, v := range fm {
		m[k] = v
	}
	return m
}
