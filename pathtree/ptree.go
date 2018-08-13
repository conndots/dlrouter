package pathtree

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"strings"
	"bytes"
)

type NodeType uint8

const (
	NodeTypeRoot    NodeType = 0
	NodeTypeLeaf    NodeType = 2
	NodeTypeDefault NodeType = 1
	NodeTypeVar     NodeType = 3

	varSymbol    = ':'
	pathSplitter = '/'
)

var (
	valID = uint(0)
)

type pathVar struct {
	valID    uint
	variable string
}
type target struct {
	valID    uint
	value    interface{}
}

type TargetCandidate struct {
	Value     interface{}
	Variables map[string]string
}

type PathTree struct {
	childrenIdx map[byte]*PathTree
	Size        int

	LeafValues []*target
	pathVars   []*pathVar //the pathVariables the node contains
	path       string
	nodeType   NodeType
}

func NewPathTree() *PathTree {
	return &PathTree{
		childrenIdx: make(map[byte]*PathTree),
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
	return &pathVar{
		variable: pvar,
		valID:    id,
	}
}

func getPathEndingAtVar(path string) string {
	pos := strings.IndexByte(path, varSymbol)
	if pos >= 0 {
		return path[:pos]
	}
	return path
}
func (ct *PathTree) Add(str string, value interface{}) error {
	valID++
	ct.Size++

	if ct.Size == 1 {
		ct.path = str
		ct.nodeType = NodeTypeRoot
	}
loopStart:
	for {
		diffSt := 0
		minLen := min(len(ct.path), len(str))
		for diffSt < minLen && str[diffSt] != varSymbol && str[diffSt] == ct.path[diffSt] {
			diffSt++
		}

		if diffSt < len(ct.path) { //split the node
			child := &PathTree{
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

			ct.childrenIdx = make(map[byte]*PathTree, 2)
			ct.childrenIdx[ct.path[diffSt]] = child
			ct.path = ct.path[:diffSt]
			if ct.nodeType == NodeTypeLeaf {
				ct.nodeType = NodeTypeDefault
			}
			ct.LeafValues = make([]*target, 0, 0)
		}

		if diffSt < len(str) { //str has diff
			if str[diffSt] == varSymbol { //var
				str = str[diffSt:]
				pos := strings.IndexByte(str, pathSplitter)
				var pvar *pathVar
				if pos == -1 { //the last element
					pvar = getPathVarWithID(str[1:], valID)
					str = ""
				} else {
					pvar = getPathVarWithID(str[1:pos], valID)
					str = str[pos:]
				}

				sub, existed := ct.childrenIdx[varSymbol]
				if existed {
					if sub.nodeType != NodeTypeVar {
						return fmt.Errorf("wrong node type %d, expected %d", sub.nodeType, NodeTypeVar)
					}
					ct = sub
					ct.pathVars = append(ct.pathVars, pvar)
					if len(str) == 0 { //str已经添加完成
						ct.LeafValues = append(ct.LeafValues, &target{
							valID: valID,
							value: value,
						})
					}

					if len(str) > 0 {
						continue loopStart
					} else {
						return nil
					}
				}

				child := &PathTree{
					pathVars: []*pathVar{pvar},
					path:     "",
					nodeType: NodeTypeVar,
					Size:     1,
				}
				if ct.childrenIdx == nil {
					ct.childrenIdx = make(map[byte]*PathTree, 2)
				}
				ct.childrenIdx[varSymbol] = child
				ct = child

				if len(str) == 0 {
					child.LeafValues = []*target{{
						value:    value,
						valID:    valID,
					}}
					return nil
				}
			} else { //normal
				str = str[diffSt:]
				c := str[0]
				sub, existed := ct.childrenIdx[c]
				if existed {
					ct = sub
					continue loopStart
				}

				//add to a new child
				child := &PathTree{
					nodeType: NodeTypeDefault,
					path:     getPathEndingAtVar(str), //无需添加变量符之后的内容，变量符需要split
					Size:     1,
				}
				if ct.childrenIdx == nil {
					ct.childrenIdx = make(map[byte]*PathTree, 2)
				}
				ct.childrenIdx[c] = child

				ct = child
				continue loopStart

			}

		} else if diffSt == len(str) {
			ct.LeafValues = append(ct.LeafValues, &target{
				value:    value,
				valID:    valID,
			})
			if ct.nodeType != NodeTypeRoot {
				ct.nodeType = NodeTypeLeaf
			}
			return nil
		}
	}

	return nil
}

func (ct *PathTree) getTargetCandidates(target string, pathVarsMap map[uint]map[string]string, candidates []*TargetCandidate) []*TargetCandidate {
	var varValue string
	end := strings.IndexByte(target, pathSplitter)
	if end == -1 {
		varValue = target
	} else {
		varValue = target[:end]
	}

	for _, pvar := range ct.pathVars {
		pmap, exist := pathVarsMap[pvar.valID]
		if !exist {
			pmap = make(map[string]string, 2)
			pathVarsMap[pvar.valID] = pmap
		}
		pmap[pvar.variable] = varValue
	}

	for _, lval := range ct.LeafValues {
		if ct.nodeType == NodeTypeVar {
			pathVars := pathVarsMap[lval.valID]
			candidates = append(candidates, &TargetCandidate{
				Value:     lval.value,
				Variables: pathVars,
			})
		} else {
			pathVars := pathVarsMap[lval.valID]
			candidates = append(candidates, &TargetCandidate{
				Value:     lval.value,
				Variables: pathVars,
			})
		}
	}
	return candidates
}

type searchContext struct {
	node          *PathTree
	partialTarget string
}

func (ct *PathTree) GetCandidateLeafs(target string) (candidates []*TargetCandidate) {
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
	广度优先遍历
	为何选择广度优先遍历？返回值默认按照最长匹配的顺序返回候选。广度优先遍历保证数组添加顺序是按照匹配长度递增的顺序
	*/
	//记录遇到的所有路径上所有的pathVars
	pathVarsMap := make(map[uint]map[string]string, 2) //map[valID]map[varName]varValue
	queue := queue.New()

	queue.Enqueue(&searchContext{
		node:          ct,
		partialTarget: target,
	})

	for queue.Len() > 0 {
		ctx := queue.Dequeue().(*searchContext)
		curr := ctx.node
		tar := ctx.partialTarget

		if curr.nodeType == NodeTypeVar {
			candidates = curr.getTargetCandidates(tar, pathVarsMap, candidates)
			pos := strings.IndexByte(tar, pathSplitter)
			if pos >= 0 {
				nextTar := tar[pos:]
				nextCh, hasChild := curr.childrenIdx[pathSplitter]
				if hasChild {
					queue.Enqueue(&searchContext{
						node:          nextCh,
						partialTarget: nextTar,
					})
				}
			}
		} else {
			i := 0
			tlen, plen := len(tar), len(curr.path)
			for ; i < tlen && i < plen && tar[i] == curr.path[i]; i++ {
			}
			if i < plen { //path与target不匹配
				continue
			} else {
				candidates = curr.getTargetCandidates(tar, pathVarsMap, candidates)

				if i < tlen { // target还有未处理的
					nextTar := tar[i:]
					next, hasChild := curr.childrenIdx[nextTar[0]]
					nextVar, hasVarChild := curr.childrenIdx[varSymbol]
					if !hasChild && !hasVarChild {
						continue
					}

					if hasVarChild {
						queue.Enqueue(&searchContext{
							node:          nextVar,
							partialTarget: nextTar,
						})
					}
					if hasChild {
						queue.Enqueue(&searchContext{
							node:          next,
							partialTarget: nextTar,
						})
					}
				}
			}
		}
	}
	return candidates
}

func (ct *PathTree) String() string {
	var buf bytes.Buffer

	type Node struct {
		node  *PathTree
		depth int
	}
	queue := make([]*Node, 0, 10)
	queue = append(queue, &Node{node: ct, depth: 0})
	currDepth := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		depth := curr.depth
		if depth > currDepth {
			buf.WriteByte('\n')
			currDepth = depth
		}
		var valStr bytes.Buffer
		valStr.WriteByte('[')
		for _, leaf := range curr.node.LeafValues {
			valStr.WriteString(fmt.Sprint(leaf.value))
			valStr.WriteString(",")
		}
		valStr.WriteByte(']')
		var varStr bytes.Buffer
		varStr.WriteByte('[')
		for _, pvar := range curr.node.pathVars {
			varStr.WriteString(pvar.variable)
			varStr.WriteString(",")
		}
		varStr.WriteByte(']')
		buf.WriteByte('[')
		buf.WriteString(curr.node.path)
		buf.WriteString(" depth(")
		buf.WriteString(fmt.Sprint(depth))
		buf.WriteString(") nodeType:")
		buf.WriteString(fmt.Sprint(curr.node.nodeType))
		buf.WriteString(" value:")
		buf.WriteString(valStr.String())
		buf.WriteString(" vars:")
		buf.WriteString(varStr.String())
		buf.WriteString("]\t")

		for _, sub := range curr.node.childrenIdx {
			queue = append(queue, &Node{node: sub, depth: depth + 1})
		}
	}
	return buf.String()
}
func (ct *PathTree) Print() {
	fmt.Println("-------------------------")
	fmt.Println(ct.String())
	fmt.Println("\n-------------------------")
}
