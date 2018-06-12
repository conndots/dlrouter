package pathtree

import (
	"fmt"
	"testing"
)

func getPreparedCTrie() *PathTree {
	trie := NewPathTree()
	trie.Add("www.google", 1)
	fmt.Println("add www.google", 1)
	trie.Print()
	fmt.Println("add www.", 2)
	trie.Add("www.", 2)
	trie.Print()
	trie.Add("www.google.hk.", 3)
	fmt.Println("add www.google.hk.", 3)
	trie.Print()
	trie.Add("www.google.us.", 4)
	fmt.Println("add www.google.us.", 4)
	trie.Print()
	trie.Add("www.google.uk", 5)
	fmt.Println("add www.google.uk", 5)
	trie.Print()
	trie.Add("www.google.uk.wtf", 6)
	fmt.Println("add www.google.uk.wtf", 6)
	trie.Print()
	return trie
}

func getPathTreeWithVar(t *PathTree) *PathTree {
	t.Add("/aweme/v:version/user/:user_id", "aweme_user")
	t.Add("/toutiao/pc/a:item_id", "tt_item")
	t.Add("/toutiao/pc/a:group_id", "tt_group")
	t.Add("/item/:item_id", "tt_item")
	t.Add("/group/:group_id", "tt_group")
	return t
}

func TestTreeWithVar(t *testing.T) {
	tree := getPathTreeWithVar(getPreparedCTrie())
	tree.Print()

	cands := tree.GetCandidateLeafs("/aweme/v1/user/12345")
	if len(cands) != 1 || cands[0].Value != "aweme_user" || cands[0].Variables["version"] != "1" || cands[0].Variables["user_id"] != "12345" {
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}
	cands = tree.GetCandidateLeafs("/toutiao/pc/a3121212")
	if len(cands) != 1 || cands[0].Value != "tt_item" || cands[0].Variables["item_id"] != "3121212" {
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}

}

func TestCTrieAddSame(t *testing.T) {
	trie := NewPathTree()
	trie.Add("www.google.com", 1)
	trie.Print()
	trie.Add("www.google.com", 2)
	trie.Print()
	cs := trie.GetCandidateLeafs("www.google.com.hk")
	if len(cs) != 2 {
		t.Errorf("error,add same elements. candidates=%v", cs)
	}
}
func TestCTrieAdd(t *testing.T) {
	trie := NewPathTree()
	trie.Add("www.fuck", true)
	if trie.Size != 1 {
		t.Errorf("pathtree add: size expected: 1. got: %v", trie.Size)
	}
	trie.Add("www.hell", true)
	if trie.Size != 2 {
		t.Errorf("pathtree add: size expected: 2, got: %v", trie.Size)
	}

	trie.Print()

	cands := trie.GetCandidateLeafs("www.fuck")
	cands = trie.GetCandidateLeafs("www.hell")
	cands = trie.GetCandidateLeafs("www.fuck.com")
	if len(cands) != 1 {
		t.Errorf("pathtree add: get expected len(candidates) == 1, got:%v", len(cands))
	}
}

func TestCTrieGetCandidates(t *testing.T) {
	trie := getPreparedCTrie()
	trie.Print()
	if trie.Size != 6 {
		t.Errorf("size not 6")
	}
	candidates := trie.GetCandidateLeafs("www.google.uk.wtf.fuck")
	if len(candidates) != 4 {
		t.Errorf("www.google.uk.wtf.fuck expected: candidates length: 4; got: %v", candidates)
	}
	if candidates[0].Value != 6 || candidates[1].Value != 5 || candidates[2].Value != 1 || candidates[3].Value != 2 {
		t.Errorf("www.google.uk.wtf.fuck expected: %v got: %v", []int{6, 5, 1, 2}, candidates)
	}
}

func BenchmarkCTrieGetCandidates(b *testing.B) {
	trie := getPreparedCTrie()
	for n := 0; n < b.N; n++ {
		trie.GetCandidateLeafs("www.google.uk.wtf.fuck.hello.what.the.fuck")
	}
}
