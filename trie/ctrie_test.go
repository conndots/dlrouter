package trie

import (
	"fmt"
	"testing"
)

func getPreparedCTrie() *CTrie {
	trie := NewCompressedTrie()
	trie.Add("www.google", 1)
	fmt.Println("add www.google")
	trie.Print()
	fmt.Println("add www.")
	trie.Add("www.", 2)
	trie.Print()
	trie.Add("www.google.hk.", 3)
	fmt.Println("add www.google.hk.")
	trie.Print()
	trie.Add("www.google.us.", 4)
	fmt.Println("add www.google.us.")
	trie.Print()
	trie.Add("www.google.uk", 5)
	fmt.Println("add www.google.uk")
	trie.Print()
	trie.Add("www.google.uk.wtf", 6)
	fmt.Println("add www.google.uk.wtf")
	trie.Print()
	return trie
}

func TestCTrieAddSame(t *testing.T) {
	trie := NewCompressedTrie()
	trie.Add("www.google.com", 1)
	trie.Print()
	trie.Add("www.google.com", 2)
	trie.Print()
	cs, _ := trie.GetCandidateLeafs("www.google.com.hk")
	if len(cs) != 2 {
		t.Errorf("error,add same elements. candidates=%v", cs)
	}
}
func TestCTrieAdd(t *testing.T) {
	trie := NewCompressedTrie()
	trie.Add("www.fuck", true)
	if trie.Size != 1 {
		t.Errorf("trie add: size expected: 1. got: %v", trie.Size)
	}
	trie.Add("www.hell", true)
	if trie.Size != 2 {
		t.Errorf("trie add: size expected: 2, got: %v", trie.Size)
	}

	trie.Print()

	_, fullMatch := trie.GetCandidateLeafs("www.fuck")
	if !fullMatch {
		t.Errorf("trie add: get expected fullMatch=%v, got: %v", true, false)
	}
	_, fullMatch = trie.GetCandidateLeafs("www.hell")
	if !fullMatch {
		t.Errorf("trie add: get expected fullMatch=%v, got: %v", true, false)
	}
	cand, fullMatch := trie.GetCandidateLeafs("www.fuck.com")
	if fullMatch {
		t.Errorf("trie add: get expected fullMatch=%v, got: %v", false, true)
	}
	if len(cand) != 1 {
		t.Errorf("trie add: get expected len(candidates) == 1, got:%v", len(cand))
	}
}

func TestCTrieGetCandidates(t *testing.T) {
	trie := getPreparedCTrie()
	trie.Print()
	if trie.Size != 6 {
		t.Errorf("size not 6")
	}
	candidates, fullMatch := trie.GetCandidateLeafs("www.google.uk.wtf.fuck")
	if fullMatch || len(candidates) != 4 {
		t.Errorf("www.google.uk.wtf.fuck expected: fullMatch=%v, candidates length: 4; got: %v, %v", false, fullMatch, candidates)
	}
	if candidates[0] != 6 || candidates[1] != 5 || candidates[2] != 1 || candidates[3] != 2 {
		t.Errorf("www.google.uk.wtf.fuck expected: %v got: %v", []int{6, 5, 1, 2}, candidates)
	}
}

func BenchmarkCTrieGetCandidates(b *testing.B) {
	trie := getPreparedCTrie()
	for n := 0; n < b.N; n++ {
		trie.GetCandidateLeafs("www.google.uk.wtf.fuck.hello.what.the.fuck")
	}
}
