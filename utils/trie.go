/**
 * Created: 2020/3/27
 * @author: Jason
 */

package utils

import "sync"

type TrieNode struct {
	Children sync.Map //map[rune]*TrieNode
	End      bool
}

func NewTrieNode() *TrieNode {
	t := &TrieNode{
		End: false,
	}
	return t
}

type Trie struct {
	Node *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		Node: NewTrieNode(),
	}
}

func (t *Trie) Add(text string) {
	if len(text) == 0 {
		return
	}
	chars := []rune(text)
	charLen := len(chars)
	node := t.Node

	for i := 0; i < charLen; i++ {
		var v *TrieNode = nil
		key := chars[i]
		childrenNode, ok := node.Children.Load(key)
		if !ok {
			v = NewTrieNode()
			node.Children.Store(key, v)
		} else {
			v = childrenNode.(*TrieNode)
		}
		node = v
	}
	node.End = true
}

func (t *Trie) Find(text string) bool {
	chars := []rune(text)
	charLen := len(chars)
	node := t.Node

	for i := 0; i < charLen; i++ {
		key := chars[i]
		childrenNode, ok := node.Children.Load(key)
		if !ok {
			continue
		}
		v, ok := childrenNode.(*TrieNode)
		if !ok {
			continue
		}
		node = v
		for j := i + 1; j < charLen; j++ {
			key2 := chars[j]
			childrenNode2, ok := node.Children.Load(key2)
			if !ok {
				break
			}
			v, ok := childrenNode2.(*TrieNode)
			if !ok {
				continue
			}
			node = v
			if node.End {
				return true
			}
		}
		node = t.Node
	}
	return false
}

func (t *Trie) Replace(text string) (string, []string) {
	chars := []rune(text)
	charLen := len(chars)
	node := t.Node

	result := []rune(text)
	find := make([]string, 0, 8)

	for i := 0; i < charLen; i++ {
		key := chars[i]
		childrenNode, ok := node.Children.Load(key)
		if !ok {
			continue
		}
		v, ok := childrenNode.(*TrieNode)
		if !ok {
			continue
		}
		node = v
		for j := i + 1; j < charLen; j++ {
			key2 := chars[j]
			childrenNode2, ok := node.Children.Load(key2)
			if !ok {
				break
			}
			v, ok := childrenNode2.(*TrieNode)
			if !ok {
				continue
			}
			node = v
			if node.End {
				for n := i; n <= j; n++ {
					result[n] = '*'
				}
				find = append(find, string(chars[i:j+1]))
				i = j
				break
			}
		}
		node = t.Node
	}

	return string(result), find
}
