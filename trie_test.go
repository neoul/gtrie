package gtrie

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
)

func addFromFile(t *Trie, path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewScanner(file)

	for reader.Scan() {
		t.Add(reader.Text(), nil)
	}

	if reader.Err() != nil {
		log.Fatal(err)
	}
}

func TestTrieAdd(t *testing.T) {
	trie := New()

	trie.Add("foo", 1)

	if v, ok := trie.Find("foo"); !ok || v != 1 {
		t.Errorf("Expected 1, got: %v", v)
	}
}

func TestTrieFind(t *testing.T) {
	trie := New()
	trie.Add("foo", 1)

	n, ok := trie.Find("foo")
	if ok != true {
		t.Fatal("Could not find node")
	}

	if n.(int) != 1 {
		t.Errorf("Expected 1, got: %d", n.(int))
	}
}

func TestTrieFindMissingWithSubtree(t *testing.T) {
	trie := New()
	trie.Add("fooish", 1)
	trie.Add("foobar", 1)

	n, ok := trie.Find("foo")
	if ok != false {
		t.Errorf("Expected ok to be false")
	}
	if n != nil {
		t.Errorf("Expected nil, got: %v", n)
	}
}

func TestTrieHasPrefix(t *testing.T) {
	trie := New()
	trie.Add("fooish", 1)
	trie.Add("foobar", 1)

	testcases := []struct {
		key      string
		expected bool
	}{
		{"foobar", true},
		{"foo", true},
		{"fool", false},
	}
	for _, testcase := range testcases {
		if trie.HasPrefix(testcase.key) != testcase.expected {
			t.Errorf("HasPrefix(\"%s\"): expected result to be %t", testcase.key, testcase.expected)
		}
	}
}

func TestTrieFindMissing(t *testing.T) {
	trie := New()

	n, ok := trie.Find("foo")
	if ok != false {
		t.Errorf("Expected ok to be false")
	}
	if n != nil {
		t.Errorf("Expected nil, got: %v", n)
	}
}

func TestRemove(t *testing.T) {
	trie := New()
	initial := []string{"football", "foostar", "foosball"}

	for _, key := range initial {
		trie.Add(key, nil)
	}

	trie.Remove("foosball")
	keys := trie.Keys("")

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys got %d", len(keys))
	}

	for _, k := range keys {
		if k != "football" && k != "foostar" {
			t.Errorf("key was: %s", k)
		}
	}

	keys = trie.FuzzySearch("foo")
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys got %d", len(keys))
	}

	for _, k := range keys {
		if k != "football" && k != "foostar" {
			t.Errorf("Expected football got: %#v", k)
		}
	}
}

func TestTrieKeys(t *testing.T) {
	tableTests := []struct {
		name         string
		expectedKeys []string
	}{
		{"Two", []string{"bar", "foo"}},
		{"One", []string{"foo"}},
		{"Empty", []string{}},
	}

	for _, test := range tableTests {
		t.Run(test.name, func(t *testing.T) {
			trie := New()
			for _, key := range test.expectedKeys {
				trie.Add(key, nil)
			}

			keys := trie.Keys("")
			if len(keys) != len(test.expectedKeys) {
				t.Errorf("Expected %v keys, got %d, keys were: %v", len(test.expectedKeys), len(keys), trie.Keys(""))
			}

			sort.Strings(keys)
			for i, key := range keys {
				if key != test.expectedKeys[i] {
					t.Errorf("Expected %#v, got %#v", test.expectedKeys[i], key)
				}
			}
		})
	}
}

func TestPrefixSearch(t *testing.T) {
	trie := New()
	expected := []string{
		"foo",
		"foosball",
		"football",
		"foreboding",
		"forementioned",
		"foretold",
		"foreverandeverandeverandever",
		"forbidden",
	}

	defer func() {
		r := recover()
		if r != nil {
			t.Error(r)
		}
	}()

	trie.Add("bar", nil)
	for _, key := range expected {
		trie.Add(key, nil)
	}

	tests := []struct {
		pre      string
		expected []string
		length   int
	}{
		{"fo", expected, len(expected)},
		{"foosbal", []string{"foosball"}, 1},
		{"abc", []string{}, 0},
	}

	for _, test := range tests {
		actual := trie.PrefixSearch(test.pre)
		sort.Strings(actual)
		sort.Strings(test.expected)
		if len(actual) != test.length {
			t.Errorf("Expected len(actual) to == %d for pre %s", test.length, test.pre)
		}

		for i, key := range actual {
			if key != test.expected[i] {
				t.Errorf("Expected %v got: %v", test.expected[i], key)
			}
		}
	}

	trie.PrefixSearch("fsfsdfasdf")
}

func TestTrie_FindLongestMatch(t *testing.T) {
	trie := New()
	expected := []string{
		"foo",
		"foosball",
		"football",
		"foreboding",
		"forementioned",
		"foretold",
		"foreverandeverandeverandever",
		"forbidden",
		"ABC",
		"/interfaces",
		"/interfaces/interface",
		"/interfaces/interface[name=1/2]",
		"/interfaces/interface[name=1/2]/state",
		"/interfaces/interface[name=1/2]/state/oper-status",
		"/interfaces/interface[name=1/2]/state/enabled",
		"/interfaces/interface[name=1/1]/state/enabled",
		"/interfaces/interface[name=1/2]/state/admin-status",
	}

	// defer func() {
	// 	r := recover()
	// 	if r != nil {
	// 		t.Error(r)
	// 	}
	// }()

	trie.Add("bar", nil)
	for _, key := range expected {
		trie.Add(key, nil)
	}

	tests := []struct {
		input    string
		expected string
		ok       bool
	}{
		{"fooo", "foo", true},
		{"foretoldme", "foretold", true},
		{"abc", "", false},
		{"/interfaces/interface[name=1/2]/config/hello", "/interfaces/interface[name=1/2]", true},
	}

	for _, test := range tests {
		t.Log("TEST input:", test.input)
		output, ok := trie.FindLongestMatchedkey(test.input)
		if (test.ok != ok) || test.expected != output {
			t.Errorf("ok %v output %s, expected %s for input %s", ok, output, test.expected, test.input)
		}
	}
}

func TestFuzzySearch(t *testing.T) {
	setup := []string{
		"foosball",
		"football",
		"bmerica",
		"ked",
		"kedlock",
		"frosty",
		"bfrza",
		"foo/bart/baz.go",
	}
	tests := []struct {
		partial string
		length  int
	}{
		{"fsb", 1},
		{"footbal", 1},
		{"football", 1},
		{"fs", 2},
		{"oos", 1},
		{"kl", 1},
		{"ft", 3},
		{"fy", 1},
		{"fz", 2},
		{"a", 5},
		{"", 8},
		{"zzz", 0},
	}

	trie := New()
	for _, key := range setup {
		trie.Add(key, nil)
	}

	for _, test := range tests {
		t.Run(test.partial, func(t *testing.T) {
			actual := trie.FuzzySearch(test.partial)
			if len(actual) != test.length {
				t.Errorf("Expected len(actual) to == %d, was %d for %s actual was %#v",
					test.length, len(actual), test.partial, actual)
			}
		})
	}
}

func TestFuzzySearchSorting(t *testing.T) {
	trie := New()
	setup := []string{
		"foosball",
		"football",
		"bmerica",
		"ked",
		"kedlock",
		"frosty",
		"bfrza",
		"foo/bart/baz.go",
	}

	for _, key := range setup {
		trie.Add(key, nil)
	}

	actual := trie.FuzzySearch("fz")
	expected := []string{"bfrza", "foo/bart/baz.go"}

	if len(actual) != len(expected) {
		t.Fatalf("expected len %d got %d", len(expected), len(actual))
	}
	for i, v := range expected {
		if actual[i] != v {
			t.Errorf("Expected %s got %s", v, actual[i])
		}
	}

}

func BenchmarkTieKeys(b *testing.B) {
	trie := New()
	keys := []string{"bar", "foo", "baz", "bur", "zum", "burzum", "bark", "barcelona", "football", "foosball", "footlocker"}

	for _, key := range keys {
		trie.Add(key, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Keys("")
	}
}

func BenchmarkPrefixSearch(b *testing.B) {
	trie := New()
	addFromFile(trie, "/usr/share/dict/words")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trie.PrefixSearch("fo")
	}
}

func BenchmarkFuzzySearch(b *testing.B) {
	trie := New()
	addFromFile(trie, "/usr/share/dict/words")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trie.FuzzySearch("fs")
	}
}

func BenchmarkBuildTree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trie := New()
		addFromFile(trie, "/usr/share/dict/words")
	}
}

func TestSupportChinese(t *testing.T) {
	trie := New()
	expected := []string{"苹果 沂水县", "苹果", "大蒜", "大豆"}

	for _, key := range expected {
		trie.Add(key, nil)
	}

	tests := []struct {
		pre      string
		expected []string
		length   int
	}{
		{"苹", expected[:2], len(expected[:2])},
		{"大", expected[2:], len(expected[2:])},
		{"大蒜", []string{"大蒜"}, 1},
	}

	for _, test := range tests {
		actual := trie.PrefixSearch(test.pre)
		sort.Strings(actual)
		sort.Strings(test.expected)
		if len(actual) != test.length {
			t.Errorf("Expected len(actual) to == %d for pre %s", test.length, test.pre)
		}

		for i, key := range actual {
			if key != test.expected[i] {
				t.Errorf("Expected %v got: %v", test.expected[i], key)
			}
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	f, err := os.Open("/usr/share/dict/words")
	if err != nil {
		b.Fatal("couldn't open bag of words")
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var words []string
	for scanner.Scan() {
		word := scanner.Text()
		words = append(words, word)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie := New()
		for k := range words {
			trie.Add(words[k], nil)
		}
	}
}

func BenchmarkAddRemove(b *testing.B) {
	words := []string{"AAAA1", "AAAA2", "ABAA1", "AABA1", "ABAA2"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie := New()
		for k := range words {
			trie.Add(words[k], nil)
		}
		for k := range words {
			trie.Remove(words[k])
		}
	}
}

func TestTrie_All(t *testing.T) {
	trie := New()
	expected := []string{
		"foo",
		"foosball",
		"football",
		"foreboding",
		"forementioned",
		"foretold",
		"foreverandeverandeverandever",
		"forbidden",
		"ABC",
		"/interfaces",
		"/interfaces/interface",
		"/interfaces/interface[name=1/2]",
		"/interfaces/interface[name=1/2]/state",
		"/interfaces/interface[name=1/2]/state/oper-status",
		"/interfaces/interface[name=1/2]/state/enabled",
		"/interfaces/interface[name=1/1]/state/enabled",
		"/interfaces/interface[name=1/2]/state/admin-status",
	}

	for _, key := range expected {
		trie.Add(key, true)
	}

	tests := []struct {
		name string
		pre  string
		want map[string]interface{}
	}{
		{
			name: "All",
			pre:  "/interfaces",
			want: map[string]interface{}{
				"/interfaces":                                        true,
				"/interfaces/interface":                              true,
				"/interfaces/interface[name=1/2]":                    true,
				"/interfaces/interface[name=1/2]/state":              true,
				"/interfaces/interface[name=1/2]/state/oper-status":  true,
				"/interfaces/interface[name=1/2]/state/enabled":      true,
				"/interfaces/interface[name=1/1]/state/enabled":      true,
				"/interfaces/interface[name=1/2]/state/admin-status": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trie.All(tt.pre); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Trie.All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrie_FindMatchedKey(t *testing.T) {
	trie := New()
	expected := []string{
		"foo",
		"foosball",
		"football",
		"foreboding",
		"forementioned",
		"foretold",
		"foreverandeverandeverandever",
		"forbidden",
		"ABC",
		"/interfaces",
		"/interfaces/interface",
		"/interfaces/interface[name=1/2]",
		"/interfaces/interface[name=1/2]/state",
		"/interfaces/interface[name=1/2]/state/oper-status",
		"/interfaces/interface[name=1/2]/state/enabled",
		"/interfaces/interface[name=1/1]/state/enabled",
		"/interfaces/interface[name=1/2]/state/admin-status",
		"/interfaces/interface/state/counters",
	}

	for _, key := range expected {
		trie.Add(key, true)
	}

	tests := []struct {
		name string
		key  string
		want []string
	}{
		{
			name: "FindMatchedKey",
			key:  "/interfaces/interface[name=1/2]/",
			want: []string{
				"/interfaces",
				"/interfaces/interface",
				"/interfaces/interface[name=1/2]",
			},
		},
		{
			name: "FindMatchedKey",
			key:  "/interfaces/interface[name=1/2]/state",
			want: []string{
				"/interfaces",
				"/interfaces/interface",
				"/interfaces/interface[name=1/2]",
				"/interfaces/interface[name=1/2]/state",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, ok := trie.FindMatchedKey(tt.key); !ok || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Trie.All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrie_FindAll(t *testing.T) {
	trie := New()
	expected := []string{
		"foo",
		"foosball",
		"football",
		"foreboding",
		"forementioned",
		"foretold",
		"foreverandeverandeverandever",
		"forbidden",
		"ABC",
		"/interfaces",
		"/interfaces/interface",
		"/interfaces/interface[name=1/2]",
		"/interfaces/interface[name=1/2]/state",
		"/interfaces/interface[name=1/2]/state/oper-status",
		"/interfaces/interface[name=1/2]/state/enabled",
		"/interfaces/interface[name=1/1]/state/enabled",
		"/interfaces/interface[name=1/2]/state/admin-status",
		"/interfaces/interface/state/counters",
	}

	for _, key := range expected {
		trie.Add(key, true)
	}

	tests := []struct {
		name string
		key  string
		want map[string]interface{}
	}{
		{
			name: "FindAll",
			key:  "/interfaces/interface[name=1/2]",
			want: map[string]interface{}{
				"/interfaces":                                        true,
				"/interfaces/interface":                              true,
				"/interfaces/interface[name=1/2]":                    true,
				"/interfaces/interface[name=1/2]/state":              true,
				"/interfaces/interface[name=1/2]/state/oper-status":  true,
				"/interfaces/interface[name=1/2]/state/enabled":      true,
				"/interfaces/interface[name=1/2]/state/admin-status": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trie.FindAll(tt.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Trie.FindAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
