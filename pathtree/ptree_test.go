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
	trie.Add("/aweme/v1/aweme/post", "aweme_post")
	trie.Print()
	return trie
}

func getPathTreeWithVar(t *PathTree) *PathTree {
	t.Add("/aw/v:version/user/:user_id", "aw_user")
	t.Add("/aw/v:version/poi/feed/", "aw_poi_feed")
	t.Add("/tt/pc/a:item_id", "tt_item")
	t.Add("/tt/pc/a:group_id", "tt_group")
	t.Add("/item/:item_id", "tt_item")
	t.Add("/i:item_id/info/", "m_info")
	t.Add("/group/:group_id", "tt_group")
	t.Add("/service/:version/information/:group_id/", "app_info")
	t.Add("/aw/v1/:search_type/search/", "aw_search")
	t.Add("/aw/v:version/feed/", "aw_feed")
	t.Add("/hot/item/video/play/", "hot_play")
	t.Add("/hot/item/:item_id/comments/", "hot_comments")
	return t
}

func TestTreeWithVar(t *testing.T) {
	tree := getPathTreeWithVar(getPreparedCTrie())
	// tree := getPathTreeWithVar(NewPathTree())
	fmt.Println("===============================")
	tree.Print()

	cands := tree.GetCandidateLeafs("/aw/v1/user/12345")
	if len(cands) != 1 || cands[0].Value != "aw_user" || cands[0].Variables["version"] != "1" || cands[0].Variables["user_id"] != "12345" {
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}
	// cands = tree.GetCandidateLeafs("/toutiao/pc/a3121212")
	// if len(cands) != 2 {
	// 	for i, c := range cands {
	// 		t.Errorf("candidate %d: %v", i, *c)
	// 	}
	// }
	cands = tree.GetCandidateLeafs("/aw/v2/hehe/feed/")
	if len(cands) != 0 {
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}

	cands = tree.GetCandidateLeafs("www.google.uk.wtf.hehe")
	if len(cands) != 4 {
		t.Errorf("www.google.uk.wtf.hehe get err")
	}

	candidates := tree.GetCandidateLeafs("www.google.uk.wtf.fuck")
	if len(candidates) != 4 {
		t.Errorf("www.google.uk.wtf.fuck expected: candidates length: 4; got: %v", candidates)
	}
	if candidates[0].Value != 6 || candidates[1].Value != 5 || candidates[2].Value != 1 || candidates[3].Value != 2 {
		t.Errorf("www.google.uk.wtf.fuck expected: %v got: %v", []int{6, 5, 1, 2}, candidates)
	}

	candidates = tree.GetCandidateLeafs("www.facebook.cn")
	if len(candidates) != 1 || candidates[0].Value != 2 {
		t.Errorf("www.google.cn error")
		for i, c := range candidates {
			t.Errorf("candidate %d: %v", i, *c)
		}

	}
	candidates = tree.GetCandidateLeafs("wtf.google.cn")
	if len(candidates) != 0 {
		t.Errorf("wtf.google.cn error")
	}

	cands = tree.GetCandidateLeafs("/service/2/information/12345/detail")
	if len(cands) != 1 || cands[0].Value != "app_info" || cands[0].Variables["version"] != "2" || cands[0].Variables["group_id"] != "12345" {

		t.Errorf("/service/2/information/12345/detail get err")
	}
	cands = tree.GetCandidateLeafs("/aw/v1/discover/search/")
	if len(cands) != 1 || cands[0].Value != "aw_search" || cands[0].Variables["search_type"] != "discover" {
		t.Errorf("/aw/v1/discover/search/ get err")
	}
	cands = tree.GetCandidateLeafs("/aw/v1/aw/post/")
	if len(cands) > 0 {
		t.Errorf("/aw/v1/post/ candidates: %v", cands)
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}

	cands = tree.GetCandidateLeafs("/hot/item/12345/comments/")
	if len(cands) == 0 || cands[0].Value != "hot_comments" || cands[0].Variables["item_id"] != "12345" {
		t.Errorf("/hotsoon/item/12345/comments/ candidates: %v", cands)
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}

	cands = tree.GetCandidateLeafs("/hot/item/video/play/")
	if len(cands) == 0 || cands[0].Value != "hot_play" {
		t.Errorf("/hot/item/video/play/ candidates: %v", cands)
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}

	cands = tree.GetCandidateLeafs("/hot/item/12345/act/")
	if len(cands) > 0 {
		t.Errorf("/hot/item/12345/act/ candidates: %v", cands)
		for i, c := range cands {
			t.Errorf("candidate %d: %v", i, *c)
		}
	}

	cands = tree.GetCandidateLeafs("/i6588031967848038915/info/")
	if len(cands) == 0 || cands[0].Value != "m_info" || cands[0].Variables["item_id"] != "6588031967848038915" {
		t.Errorf("/i:item/info/ err: %v", cands)
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
	if trie.Size != 7 {
		t.Errorf("size not 6")
	}
	candidates := trie.GetCandidateLeafs("www.google.uk.wtf.fuck")
	if len(candidates) != 4 {
		t.Errorf("www.google.uk.wtf.fuck expected: candidates length: 4; got: %v", candidates)
	}
	if candidates[0].Value != 6 || candidates[1].Value != 5 || candidates[2].Value != 1 || candidates[3].Value != 2 {
		t.Errorf("www.google.uk.wtf.fuck expected: %v got: %v", []int{6, 5, 1, 2}, candidates)
	}

	candidates = trie.GetCandidateLeafs("www.facebook.cn")
	if len(candidates) != 1 || candidates[0].Value != 2 {
		t.Errorf("www.google.cn error")
		for i, c := range candidates {
			t.Errorf("candidate %d: %v", i, *c)
		}

	}
	candidates = trie.GetCandidateLeafs("wtf.google.cn")
	if len(candidates) != 0 {
		t.Errorf("wtf.google.cn error")
	}
	candidates = trie.GetCandidateLeafs("/aweme/v1/aweme/post/")
	if len(candidates) != 1 || candidates[0].Value != "aweme_post" {
		t.Errorf("/aweme/v1/aweme/post/ error")
	}
}

func BenchmarkCTrieGetCandidates(b *testing.B) {
	trie := getPreparedCTrie()
	for n := 0; n < b.N; n++ {
		trie.GetCandidateLeafs("www.google.uk.wtf.fuck.hello.what.the.fuck")
	}
}
