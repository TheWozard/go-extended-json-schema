package internal

import "regexp"

// SparseTree allows lookup of the last value along a given path would be for a tree
type SparseTree struct {
	ok      bool
	value   string
	matcher matcher
}

// Returns the last value in the tree that exists along the path
func (s *SparseTree) Search(path []string) (string, bool) {
	if s == nil {
		return "", false
	}

	if len(path) == 0 {
		return s.value, s.ok
	}

	if s.matcher != nil {
		if tree := s.matcher.Match(path[0]); tree != nil {
			if value, ok := tree.Search(path[1:]); ok {
				return value, true
			}
		}
	}

	return s.value, s.ok
}

// Generic matching for keys to a SparseTree
type matcher interface {
	Match(key string) *SparseTree
}

type objectMatcher struct {
	matches map[string]*SparseTree
}

func (om objectMatcher) Match(key string) *SparseTree {
	return om.matches[key]
}

type regexMatcher struct {
	exp  *regexp.Regexp
	tree *SparseTree
}

func (rm regexMatcher) Match(key string) *SparseTree {
	if rm.exp.Match([]byte(key)) {
		return rm.tree
	}
	return nil
}
