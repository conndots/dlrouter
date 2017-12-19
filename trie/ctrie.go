package trie

import (
	"fmt"
	"strings"
)

type NodeType uint8

const (
	NodeTypeRoot    NodeType = 0
	NodeTypeLeaf    NodeType = 2
	NodeTypeDefault NodeType = 1
	NodeTypeVar     NodeType = 3

	varSymbol = ':'
)

var (
	valID = uint(0)
)
type target struct {
	valID uint
	value interface{}
	pathVars []string
}

type CTrie struct {
	childrenIdx map[byte]*CTrie
	Size        int

	LeafValues []*target
	path       string
	nodeType   NodeType
	pathVars   []string
}

func NewCompressedTrie() *CTrie {
	return &CTrie{
		childrenIdx: make(map[byte]*CTrie),
		Size:        0,
		LeafValues:  make([]*target, 0, 1),
		nodeType:    NodeTypeRoot,
		pathVars:    make([]string, 0, 0),
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func (ct *CTrie) Add(str string, value interface{}) error {
	valID++
	tar := target{
		valID: valID,
		value: value,
		pathVars: make([]string, 0, 0),
	}

	ct.Size++
	if len(ct.path) > 0 || len(ct.childrenIdx) > 0 {
	loopStart:
		for {
			diffSt := 0
			minLen := min(len(ct.path), len(str))
			for diffSt < minLen && str[diffSt] == ct.path[diffSt] {
				diffSt++
			}

			if diffSt < len(ct.path) { //split the node
				child := &CTrie{
					childrenIdx: ct.childrenIdx,
					path:        ct.path[diffSt:],
					Size:        ct.Size,
					LeafValues:  append([]*target, ct.LeafValues...),
					nodeType:    ct.nodeType,
				}
				if ct.nodeType == NodeTypeRoot {
					if len(ct.LeafValues) > 0 {
						child.nodeType = NodeTypeLeaf
					} else {
						child.nodeType = NodeTypeDefault
					}
				}

				ct.childrenIdx = make(map[byte]*CTrie)
				ct.childrenIdx[ct.path[diffSt]] = child
				ct.path = ct.path[:diffSt]
				if ct.nodeType == NodeTypeLeaf {
					ct.nodeType = NodeTypeDefault
				}
				ct.LeafValues = make([]*target, 0, 0)
			}

			if diffSt < len(str) { //str has diff
				if str[diffSt] != varSymbol {
					str = str[diffSt:]
					c := str[0]

					sub, existed := ct.childrenIdx[c]
					if existed {
						ct = sub
						continue loopStart
					}

					//add to a new child
					child := &CTrie{
						LeafValues: []*target{&tar},
						nodeType:   NodeTypeLeaf,
						path:       str,
						Size:       1,
					}
					if ct.childrenIdx == nil {
						ct.childrenIdx = make(map[byte]*CTrie)
					}
					ct.childrenIdx[c] = child
					return nil
				} else { //next part is a path Variable
					str = str[diffSt:]
					pos := strings.IndexByte(str, '/')
					pathVar := ""
					leafValues := make([]*target, 0, 1)
					if pos == -1 {
						str = ""
						pathVar = fmt.Sprintf("%d_%s", tar.valID, str[1:])
						tar.pathVars = append(tar.pathVars, pathVar)
						leafValues = append(leafValues, tar)
					} else {
						pathVar = fmt.Sprintf("%d_%s", tar.valID, str[1:pos])
						tar.pathVars = append(tar.pathVars, pathVar)
						str = str[pos:]
					}

					sub, existed := ct.childrenIdx[varSymbol]
					if existed {
						ct = sub
						ct.pathVars = append(ct.pathVars, pathVar)
						continue loopStart
					}

					child := &CTrie{
						LeafValues: leafValues,
						path: "",
						nodeType: NodeTypeVar,
						Size: len(leafValues),
						pathVars: append(make([]string, 0, 1), pathVar),
					}
				}

			} else if diffSt == len(str) {
				ct.LeafValues = append(ct.LeafValues, &tar)
				if ct.nodeType != NodeTypeRoot {
					ct.nodeType = NodeTypeLeaf
				}
				return nil
			}
		}
	} else {
		ct.path = str
		ct.LeafValues = append(ct.LeafValues, value)
		ct.nodeType = NodeTypeRoot
	}
}

func (ct *CTrie) GetCandidateLeafs(target string) (candidates []interface{}, fullMatch bool) {
	candidates = make([]interface{}, 0, 2)
	if len(target) == 0 {
		fullMatch = false
		return
	}
	defer func() {
		//reverse it, because the longest match matters.
		for st, end := 0, len(candidates)-1; st < end; st, end = st+1, end-1 {
			candidates[st], candidates[end] = candidates[end], candidates[st]
		}
	}()
	tlen := len(target)

	for {
		if ct.path == target {
			candidates = append(candidates, ct.LeafValues...)
			fullMatch = true
			return
		} else if tlen < len(ct.path) {
			fullMatch = false
			return
		} else {
			st := 0
			ctlen := len(ct.path)
			tlen := len(target)
			for st < ctlen && st < tlen && ct.path[st] == target[st] {
				st++
			}
			if st == tlen || st != ctlen {
				fullMatch = false
				return
			}
			target = target[st:]
			if ct.nodeType == NodeTypeLeaf || (len(ct.LeafValues) > 0 && ct.nodeType == NodeTypeRoot) {
				candidates = append(candidates, ct.LeafValues...)
			}
			c := target[0]
			sub, existed := ct.childrenIdx[c]
			if existed {
				ct = sub
			} else {
				fullMatch = false
				return
			}
		}
	}
}

func (ct *CTrie) Print() {
	type Node struct {
		node  *CTrie
		depth int
	}
	queue := make([]*Node, 0, 10)
	queue = append(queue, &Node{node: ct, depth: 0})
	currDepth := 0
	fmt.Println("-------------------------")
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		depth := curr.depth
		if depth > currDepth {
			fmt.Println()
			currDepth = depth
		}
		fmt.Print("[", curr.node.path, " depth(", depth, ") nodeType:", curr.node.nodeType, " value:", curr.node.LeafValues, "]\t")

		for _, sub := range curr.node.childrenIdx {
			queue = append(queue, &Node{node: sub, depth: depth + 1})
		}
	}
	fmt.Println("\n-------------------------")
}
