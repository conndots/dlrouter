package dlrouter

import (
	"regexp"
	"strings"
	"log"

	"code.byted.org/whale/tools/trie"
	"code.byted.org/whale/tools/util"
)

type RegexTarget struct {
	RegexExp *regexp.Regexp
	Target   interface{}
}

type DomainMappingManager struct {
	Domain               string
	LocationExactSearch  map[string]interface{}
	MinPrefixLength      int
	LocationPrefixSearch *trie.CTrie
	LocationRegexSearch  []*RegexTarget
}

type LocationsMappingManager struct {
	DomainExactSearch   map[string]*DomainMappingManager
	DomainPostfixSearch *trie.CTrie
	DomainPrefixSearch  *trie.CTrie
}

func newDomainMappingManager(domain string) *DomainMappingManager {
	return &DomainMappingManager{
		Domain:               domain,
		LocationExactSearch:  make(map[string]interface{}),
		LocationPrefixSearch: trie.NewCompressedTrie(),
		MinPrefixLength:      int(^uint(0) >> 1),
		LocationRegexSearch:  make([]*RegexTarget, 0, 3),
	}
}

func (dm *DomainMappingManager) appendConf(dconf *domainConf) {
	if dconf.Domain != dm.Domain {
		return
	}

	for _, location := range dconf.Locations {
		location = strings.TrimSpace(location)

		if len(location) == 0 {
			continue
		}

		if strings.Index(location, "= ") == 0 {
			remain := strings.TrimSpace(location[2:])
			dm.LocationExactSearch[remain] = dconf.Target
		} else if strings.Index(location, "~ ") == 0 {
			remain := strings.TrimSpace(location[2:])
			regexExp, err := regexp.Compile(remain)
			if err != nil {
				util.Logger.Error("[scene_map] Regex compile error. error=%v. origin=%s. target=%v.", err, remain, dconf.Target)
			} else {
				dm.LocationRegexSearch = append(dm.LocationRegexSearch, &RegexTarget{
					RegexExp: regexExp,
					Target:   dconf.Target,
				})
			}
		} else {
			dm.LocationPrefixSearch.Add(location, dconf.Target)
			locLen := len(location)
			if locLen < dm.MinPrefixLength {
				dm.MinPrefixLength = locLen
			}
		}
	}
}

func NewLocationsMappingManager(locationConfs []*LocationConf) *LocationsMappingManager {
	domainExactSearch := make(map[string]*DomainMappingManager)
	for _, lconf := range locationConfs {
		if len(lconf.MappingConf) == 0 || lconf.Target == nil {
			continue
		}
		confs, err := getDomainConfs(lconf)
		if err != nil {
			log.Fatalf("get scene domain confs error: %v.\n", err)
			continue
		}

		for _, conf := range confs {
			if man, existed := domainExactSearch[conf.Domain]; existed {
				man.appendConf(conf)
			} else {
				newMan := newDomainMappingManager(conf.Domain)
				newMan.appendConf(conf)
				domainExactSearch[conf.Domain] = newMan
			}
		}
	}

	ins := &LocationsMappingManager{
		DomainExactSearch:   domainExactSearch,
		DomainPostfixSearch: trie.NewCompressedTrie(),
		DomainPrefixSearch:  trie.NewCompressedTrie(),
	}

	for domain, man := range domainExactSearch {
		domainBytes := []byte(domain)
		ins.DomainPrefixSearch.Add(domain, man)

		domainBytesRev := util.GetReversedBytes(domainBytes)
		ins.DomainPostfixSearch.Add(string(domainBytesRev), man)
	}

	return ins
}

//Iterator Pattern using Closure
func (m *LocationsMappingManager) getDomainManagerIterator(domain string) func() (*DomainMappingManager, bool) {
	currentStage := 0
	stageIdx := 0
	stageCandidates := make([]*DomainMappingManager, 2)

	return func() (*DomainMappingManager, bool) {
		GetNextInStage := func(stage int) (*DomainMappingManager, bool) {
			switch stage {
			case 0: //exact match
				if stageIdx > 0 {
					return nil, false
				}
				man, present := m.DomainExactSearch[domain]
				stageIdx++
				return man, present
			case 1: //后缀反向匹配
				if stageIdx == 0 {
					domainBytes := []byte(domain)
					reversedDomain := string(util.GetReversedBytes(domainBytes))
					candidatesRaw, _ := m.DomainPostfixSearch.GetCandidateLeafs(reversedDomain)
					for _, raw := range candidatesRaw {
						stageCandidates = append(stageCandidates, raw.(*DomainMappingManager))
					}
				}
				if len(stageCandidates) <= stageIdx {
					return nil, false
				}
				next := stageCandidates[stageIdx]
				stageIdx++
				return next, true
			case 2: //前缀匹配
				if stageIdx == 0 {
					candidatesRaw, _ := m.DomainPrefixSearch.GetCandidateLeafs(domain)
					for _, raw := range candidatesRaw {
						stageCandidates = append(stageCandidates, raw.(*DomainMappingManager))
					}
				}
				if len(stageCandidates) <= stageIdx {
					return nil, false
				}
				next := stageCandidates[stageIdx]
				stageIdx++
				return next, true
			default:
				return nil, false
			}
		}

		if currentStage >= 2 {
			return nil, false
		}

		var man *DomainMappingManager
		ok := false
		for man, ok = GetNextInStage(currentStage); currentStage < 2 && !ok; man, ok = GetNextInStage(currentStage) {
			//upgrade stage
			currentStage++
			stageCandidates = make([]*DomainMappingManager, 0, 0)
			stageIdx = 0
		}

		return man, ok
	}
}

func (dm *DomainMappingManager) getTargetsForPath(path string, getAll bool) ([]interface{}, bool) {
	targets := make([]interface{}, 0, 3)
	//首先寻求精确匹配
	t, present := dm.LocationExactSearch[path]
	if present {
		targets = append(targets, t)
		if !getAll {
			return targets, true
		}
	}

	//前缀匹配
	pathBytes := []byte(path)
	if dm.LocationPrefixSearch.Size > 0 && dm.MinPrefixLength <= len(path) {
		candidates, _ := dm.LocationPrefixSearch.GetCandidateLeafs(path)
		if len(candidates) > 0 {
			targets = append(targets, candidates...)
			if !getAll {
				return targets, true
			}
		}
	}

	for _, regexTar := range dm.LocationRegexSearch {
		match := regexTar.RegexExp.Find(pathBytes)
		if match != nil {
			targets = append(targets, regexTar.Target)
			if !getAll {
				return targets, true
			}
		}
	}
	return targets, len(targets) > 0
}

func (m *LocationsMappingManager) GetTarget(domain string, path string) (interface{}, bool) {
	dmanIterator := m.getDomainManagerIterator(domain)

	for dm, present := dmanIterator(); present; dm, present = dmanIterator() {
		targets, matched := dm.getTargetsForPath(path, false)
		if matched {
			return targets[0], true
		}
	}
	return nil, false
}

func (m *LocationsMappingManager) GetAllTargets(domain string, path string) ([]interface{}, bool) {
	dmanIterator := m.getDomainManagerIterator(domain)
	targets := make([]interface{}, 0, 3)

	for dm, present := dmanIterator(); present; dm, present = dmanIterator() {
		tars, matched := dm.getTargetsForPath(path, true)
		if matched {
			targets = append(targets, tars...)
		}
	}

	//remove duplicates
	targets = util.RemoveDuplicates(targets)

	return targets, len(targets) > 0
}
