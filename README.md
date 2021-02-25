[![GoDoc](https://godoc.org/github.com/derekparker/trie?status.svg)](https://godoc.org/github.com/neoul/gtrie)

# Trie
Data structure and relevant algorithms for extremely fast prefix/fuzzy string searching.

Package gtrie is an implementation of an R-Way Trie data structure.
This package supports more useful functions for the trie based on
derekparker/trie (https://godoc.org/github.com/derekparker/trie).

## Usage

Create a Trie with:

```Go
t := trie.New()
```

Add Keys with:

```Go
// Add can take in data information which can be stored with the key.
// i.e. you could store any information you would like to associate with
// this particular key.
t.Add("foobar", 1)
```

Find a key with:

```Go
node, ok := t.Find("foobar")
data := node.data()
// use data with data.(type)
```

Remove Keys with:

```Go
t.Remove("foobar")
```

Prefix search with:

```Go
t.PrefixSearch("foo")
```

Fast test for valid prefix:

```Go
t.HasPrefix("foo")
```

Fuzzy search with:

```Go
t.FuzzySearch("fb")
```

## Contributing

Fork this repo and run tests with:

	go test

Create a feature branch, write your tests and code and submit a pull request.

## License
MIT
