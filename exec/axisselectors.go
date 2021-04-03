package exec

import (
	"sort"

	"github.com/ChrisTrenkamp/xsel/store"
)

func unique(s []store.Cursor) []store.Cursor {
	if len(s) == 0 {
		return s
	}

	seen := make([]store.Cursor, 0, len(s))

slice:
	for i, n := range s {
		if i == 0 {
			s = s[:0]
		}

		for _, t := range seen {
			if n.Pos() == t.Pos() {
				continue slice
			}
		}

		seen = append(seen, n)
		s = append(s, n)
	}

	return s
}

type ForwardSort []store.Cursor

func (a ForwardSort) Len() int           { return len(a) }
func (a ForwardSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ForwardSort) Less(i, j int) bool { return a[i].Pos() < a[j].Pos() }

type BackwardSort []store.Cursor

func (a BackwardSort) Len() int           { return len(a) }
func (a BackwardSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BackwardSort) Less(i, j int) bool { return a[i].Pos() > a[j].Pos() }

func cleanupForwardAxis(nextResult NodeSet) NodeSet {
	sort.Sort(ForwardSort(nextResult))
	return unique(nextResult)
}

func cleanupBackwardAxis(nextResult NodeSet) NodeSet {
	sort.Sort(BackwardSort(nextResult))
	return unique(nextResult)
}

func selectChild(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		for _, j := range i.Nodes() {
			result = append(result, j)
		}
	}

	return cleanupForwardAxis(result)
}

func selectAttributes(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		for _, j := range i.Attributes() {
			result = append(result, j)
		}
	}

	return NodeSet(result)
}

func selectAncestor(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = appendAncestors(i.Parent(), result)
	}

	return cleanupBackwardAxis(result)
}

func selectAncestorOrSelf(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = appendAncestors(i, result)
	}

	return cleanupBackwardAxis(result)
}

func appendAncestors(cursor store.Cursor, result []store.Cursor) []store.Cursor {
	if cursor.Parent() == cursor {
		return result
	}

	result = append(result, cursor)
	return appendAncestors(cursor.Parent(), result)
}

func selectDescendant(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = appendDescendant(i, result)
	}

	return cleanupForwardAxis(result)
}

func selectDescendantOrSelf(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = append(result, i)
		result = appendDescendant(i, result)
	}

	return cleanupForwardAxis(result)
}

func appendDescendant(cursor store.Cursor, result []store.Cursor) []store.Cursor {
	for _, i := range cursor.Nodes() {
		result = append(result, i)
		result = appendDescendant(i, result)
	}

	return result
}

func selectFollowing(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = appendFollowing(i, result)
	}

	return cleanupForwardAxis(result)
}

func appendFollowing(cursor store.Cursor, result []store.Cursor) []store.Cursor {
	parent := cursor.Parent()

	if parent == cursor {
		return result
	}

	found := false

	for _, i := range parent.Nodes() {
		if i == cursor {
			found = true
			continue
		}

		if found {
			result = append(result, i)
			result = appendDescendant(i, result)
		}
	}

	return appendFollowing(parent, result)
}

func selectFollowingSibling(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = appendFollowingSibling(i, result)
	}

	return cleanupForwardAxis(result)
}

func appendFollowingSibling(cursor store.Cursor, result []store.Cursor) []store.Cursor {
	parent := cursor.Parent()

	if parent == cursor {
		return result
	}

	children := parent.Nodes()
	start := 0

	for i := range children {
		if children[i] == cursor {
			start = i
			break
		}
	}

	return append(result, children[start+1:]...)
}

func selectNamespace(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		for _, j := range i.Namespaces() {
			result = append(result, j)
		}
	}

	return cleanupForwardAxis(result)
}

func selectParent(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = append(result, i.Parent())
	}

	return cleanupForwardAxis(result)
}

func selectPreceding(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = appendPreceding(i, result)
	}

	return cleanupBackwardAxis(result)
}

func appendPreceding(cursor store.Cursor, result []store.Cursor) []store.Cursor {
	parent := cursor.Parent()

	if parent == cursor {
		return result
	}

	found := false
	children := parent.Nodes()

	for i := len(children) - 1; i >= 0; i-- {
		if children[i] == cursor {
			found = true
			continue
		}

		if found {
			result = append(result, children[i])
			result = appendDescendant(children[i], result)
		}
	}

	return appendPreceding(parent, result)
}

func selectPrecedingSibling(nodeSet NodeSet) Result {
	result := make([]store.Cursor, 0)

	for _, i := range nodeSet {
		result = appendPrecedingSibling(i, result)
	}

	return cleanupBackwardAxis(result)
}

func appendPrecedingSibling(cursor store.Cursor, result []store.Cursor) []store.Cursor {
	parent := cursor.Parent()

	if parent == cursor {
		return result
	}

	children := parent.Nodes()
	end := 0

	for i := len(children) - 1; i >= 0; i-- {
		if children[i] == cursor {
			end = i
			break
		}
	}

	return append(result, children[:end]...)
}
