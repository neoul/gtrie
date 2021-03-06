[![GoDoc](https://godoc.org/github.com/neoul/gtrie?status.svg)](https://godoc.org/github.com/neoul/gtrie)

# gtrie.Trie

Data structure and relevant algorithms for extremely fast prefix/fuzzy string searching and longest prefix match searching or combinational search of them.

Package gtrie is an implementation of an R-Way Trie data structure. This package supports more useful functions for the trie based on [derekparker/trie](https://godoc.org/github.com/derekparker/trie).

## Usage

```go

    // Create a trie
    trie := gtrie.New()
    input := []string{
        "/interfaces",
        "/interfaces/interface",
        "/interfaces/interface[name=1/2]",
        "/interfaces/interface[name=1/2]/state",
        "/interfaces/interface[name=1/2]/state/oper-status",
        "/interfaces/interface[name=1/2]/state/enabled",
        "/interfaces/interface[name=1/1]/state/enabled",
        "/interfaces/interface[name=1/2]/state/admin-status",
        "/interfaces/interface[name=1/2]/state/counters",
        "/interfaces/interface[name=1/3]",
        "/interfaces/interface[name=1/3]/state",
        "/interfaces/interface[name=1/3]/state/oper-status",
        "/interfaces/interface[name=1/3]/state/enabled",
        "/interfaces/interface[name=1/3]/state/enabled",
        "/interfaces/interface[name=1/3]/state/admin-status",
        "/interfaces/interface[name=1/3]/state/counters",
        "/interfaces/interface/state/counters",
    }

    // Add keys with your data (value)
    for _, key := range input {
        trie.Add(key, true)
    }

    // Find - Find your data with a key.
    pretty.Println(trie.Find("/interfaces/interface[name=1/3]"))

    // FindByPrefix - Find all keys starting with `prefix` from the trie.
    pretty.Println(trie.FindByPrefix("/interfaces/interface[name=1/2]/state"))

    // FindByFuzzy - Find all keys by fuzzy search (Approximate string matching).
    pretty.Println(trie.FindByFuzzy("/interfaces/interface/state"))

    // FindLongestMatchingPrefix - Find a prefix key matching longestly with input `key`.
    pretty.Println(trie.FindLongestMatchingPrefix("/interfaces/interface[name=1/3]/state/absss"))

    // FindMatchingPrefix - Find all the matching prefixes against to the input `key`.
    pretty.Println(trie.FindMatchingPrefix("/interfaces/interface[name=1/3]/state/absss"))

    // Remove the key from the trie.
    trie.Remove("/interfaces")
    trie.Remove("/interfaces/interface")

    // FindRelativeAll = FindByPrefix + FindByFuzzy + FindMatchingPrefix
    m := trie.FindRelativeAll("/interfaces/interface/state")
    pretty.Print(m)
    if len(m) != 12 {
        fmt.Printf("got result(%d), expect(12)", len(m))
    }

    // Clear all keys and inserted values.
    trie.Clear()
```

## License
MIT
