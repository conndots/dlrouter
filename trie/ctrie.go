package trie

import (
	"fmt"
	"strings"
	"github.com/golang-collections/collections/queue"
	"github.com/golang-collections/collections/stack"
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
type pathVar struct {
	valID uint
	variable string
}
type target struct {
	valID uint
	value interface{}
	pathVars []*pathVar // the path vars the new path added
}

type TargetCandidate struct {
	Value interface{}
	Variables map[string]string
}

type CTrie struct {
	childrenIdx map[byte]*CTrie
	Size        int

	LeafValues []*target
	path       string
	nodeType   NodeType
}

func NewCompressedTrie() *CTrie {
	return &CTrie{
		childrenIdx: make(map[byte]*CTrie),
		Size:        0,
		LeafValues:  make([]*target, 0, 1),
		nodeType:    NodeTypeRoot,
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func getPathVarWithID(pvar string, id uint) *pathVar {
	return &pathVar {
		variable: pvar,
		valID: id,
	}
}
func (ct *CTrie) Add(str string, value interface{}) error {
	valID++
	tar := target{
		valID: valID,
		value: value,
		pathVars: make([]*pathVar, 0, 0),
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
					LeafValues:  ct.LeafValues,
					nodeType:    ct.nodeType,
				}
				if ct.nodeType == NodeTypeRoot {
					if len(ct.LeafValues) > 0 {
						child.nodeType = NodeTypeLeaf
					} else {
						child.nodeType = NodeTypeDefault
					}
				}

				ct.childrenIdx = make(map[byte]*CTrie, 2)
				ct.childrenIdx[ct.path[diffSt]] = child
				ct.path = ct.path[:diffSt]
				if ct.nodeType == NodeTypeLeaf {
					ct.nodeType = NodeTypeDefault
				}
				ct.LeafValues = make([]*target, 0, 0)
			}

			if diffSt < len(str) { //str has diff
				if str[diffSt] != varSymbol { //normal string
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
						ct.childrenIdx = make(map[byte]*CTrie, 2)
					}
					ct.childrenIdx[c] = child
					return nil
				} else { //next part is a path Variable
					str = str[diffSt:]
					pos := strings.IndexByte(str, '/')
					var pvar *pathVar
					if pos == -1 { //the last element
						str = ""
						pvar = getPathVarWithID(str[1:], tar.valID)
						tar.pathVars = append(tar.pathVars, pvar)
					} else {
						pvar = getPathVarWithID(str[1:pos], tar.valID)
						tar.pathVars = append(tar.pathVars, pvar)
						str = str[pos:]
					}

					sub, existed := ct.childrenIdx[varSymbol]
					if existed {
						if sub.nodeType != NodeTypeVar {
							return fmt.Errorf("wrong node type %d, expected %d", sub.nodeType, NodeTypeVar)
						}
						ct = sub
						ct.LeafValues = append(ct.LeafValues, &tar)
						if len(str) > 0 {
							continue loopStart
						} else {
							return nil
						}
					}

					child := &CTrie{
						LeafValues: []*target{&tar},
						path: "",
						nodeType: NodeTypeVar,
						Size: 1,
					}
					if ct.childrenIdx == nil {
						ct.childrenIdx = make(map[byte]*CTrie, 2)
					}
					ct.childrenIdx[varSymbol] = child

					if len(str) == 0 {
						return nil
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
		ct.LeafValues = append(ct.LeafValues, &tar)
		ct.nodeType = NodeTypeRoot
	}
	return nil
}



// processVarNodes returns the target string after the process
func (t *target) processVarNodes(target string, pathVarsMap map[uint]map[string]string) string {
	var varValue string
	end := strings.IndexByte(target, '/')
	remain := ""
	if end == -1 || end == len(target) - 1 {
		varValue = target
	} else {
		varValue = target[:end]
		remain = target[end + 1:]
	}
	for _, pvar := range t.pathVars {
		pmap, exist := pathVarsMap[pvar.valID]
		if !exist {
			pmap = make(map[string]string, 2)
			pathVarsMap[pvar.valID] = pmap
		}
		pmap[pvar.variable] = varValue
	}
	return remain
}
func (ct *CTrie) getTargetCandidates(target string, pathVarsMap map[uint]map[string]string, candidates []*TargetCandidate) []*TargetCandidate {
	for _, lval := range ct.LeafValues {
		if ct.nodeType == NodeTypeVar {
			target = lval.processVarNodes(target, pathVarsMap)

			pathVars := pathVarsMap[lval.valID]
			candidates = append(candidates, &TargetCandidate{
				Value: lval.value,
				Variables: pathVars,
			})
		} else {
			pathVars := pathVarsMap[lval.valID]
			candidates = append(candidates, &TargetCandidate{
				Value: lval.value,
				Variables: pathVars,
			})
		}
	}
}

type searchContext struct {
	node *CTrie
	partialTarget string
}

func (ct *CTrie) GetCandidateLeafs(target string) (candidates []*TargetCandidate) {
	if len(target) == 0 {
		return make([]*TargetCandidate, 0, 0)
	}
	candidates = make([]*TargetCandidate, 0, 2)
	defer func() {
		//reverse it, because the longest match matters.
		for st, end := 0, len(candidates)-1; st < end; st, end = st+1, end-1 {
			candidates[st], candidates[end] = candidates[end], candidates[st]
		}
	}()

	/**
	路径上无NodeTypeVar类型节点时，无需回溯；否则，使用广度优先遍历，需要回溯.stack里的节点全部是NodeTypeVar类型节点
	为何选择广度优先遍历？返回值默认按照最长匹配的顺序返回候选。广度优先遍历保证数组添加顺序是按照匹配长度递增的顺序
	 */
	varMode := 0
	//记录遇到的所有路径上所有的pathVars
	queue := queue.New()
	pathVarsMap := make(map[uint]map[string]string, 2)  //map[valID]map[varName]varValue
	tlen := len(target)
	queue.Enqueue(&searchContext{
		node: ct,
		partialTarget: target,
	})

	for queue.Len() > 0 {
		ctx := queue.Dequeue().(searchContext)
		curr := ctx.node
		tar := ctx.partialTarget
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
