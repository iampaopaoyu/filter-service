package actire

import (
	"container/list"
	"strings"
)

type trieNode struct {
	// relation data
	isEnd bool               // 是否是结束节点
	index int                // 第几个节点
	fail  *trieNode          // 失败跳转节点
	child map[rune]*trieNode // 节点下面的子节点

	// extra data
	fkID  int64 // 这个节点对应的filter key id
	Level int64 // 过滤等级
}

func newTrieNode() *trieNode {
	return &trieNode{
		fail:  nil,
		index: -1,
		child: make(map[rune]*trieNode),
	}
}

type Matcher struct {
	size        int
	root        *trieNode
	indexLenMap map[int]int
}

func NewMatcher() *Matcher {
	return &Matcher{
		size:        0,
		root:        newTrieNode(),
		indexLenMap: make(map[int]int),
	}
}

func (m *Matcher) Insert(s string, fkID int64, level int64) {
	m.insert(s, fkID, level)
}

func (m *Matcher) Build() {
	m.build()
}

func (m *Matcher) Filter(in string) (out string, level int64, hitWords []string) {
	curNode := m.root
	var p *trieNode = nil
	re := []rune(in)
	vs := []rune(strings.ToLower(in))
	for i, v := range vs {
		for curNode.child[v] == nil && curNode != m.root {
			curNode = curNode.fail
		}
		curNode = curNode.child[v]
		if curNode == nil {
			curNode = m.root
		}
		p = curNode
		if p != m.root && p.isEnd {
			start := i - m.indexLenMap[p.index] + 1
			for m := start; m <= i; m++ {
				re[m] = rune('*')
			}
			if p.Level > level {
				level = p.Level
			}
			hitWords = append(hitWords, string(vs[start:i+1]))
			p = p.fail
		}
	}
	out = string(re)
	return
}

func (m *Matcher) Test(in string) (fkIds []int64) {
	curNode := m.root
	var p *trieNode = nil
	vs := []rune(strings.ToLower(in))
	for _, v := range vs {
		for curNode.child[v] == nil && curNode != m.root {
			curNode = curNode.fail
		}
		curNode = curNode.child[v]
		if curNode == nil {
			curNode = m.root
		}
		p = curNode
		if p != m.root && p.isEnd {
			fkIds = append(fkIds, p.fkID)
			p = p.fail
		}
	}
	return
}

func (m *Matcher) build() {
	ll := list.New()
	ll.PushBack(m.root)
	for ll.Len() > 0 {
		temp := ll.Remove(ll.Front()).(*trieNode)
		var p *trieNode = nil
		for i, v := range temp.child {
			if temp == m.root {
				v.fail = m.root
			} else {
				p = temp.fail
				for p != nil {
					if p.child[i] != nil {
						v.fail = p.child[i]
						break
					}
					p = p.fail
				}
				if p == nil {
					v.fail = m.root
				}
			}
			ll.PushBack(v)
		}
	}
}

func (m *Matcher) insert(s string, fkID int64, level int64) {
	curNode := m.root
	vs := []rune(s)
	for _, v := range vs {
		if curNode.child[v] == nil {
			curNode.child[v] = newTrieNode()
		}
		curNode = curNode.child[v]
	}
	curNode.isEnd = true
	m.indexLenMap[m.size] = len([]rune(s))

	// extra data
	curNode.fkID = fkID
	curNode.Level = level

	curNode.index = m.size
	m.size++
}
