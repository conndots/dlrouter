package dlrouter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/conndots/dlrouter/pathtree"
)

const (
	domainSearchStageNum = 3
)
var (
	NotSameDomainErr = errors.New("[dlrouter compile] domains are not identical")
)
type RegexTarget struct {
	RegexExp *regexp.Regexp
	Targets  []interface{}
}
type Target struct {
	Value interface{}
	Variables map[string]string
}

type DomainRouter struct {
	Domain               string
	LocationExactSearch  map[string][]interface{}
	LocationPrefixSearch *pathtree.PathTree
	LocationRegexSearch  map[string]*RegexTarget
}

type DomainLocationRouter struct {
	DomainExactSearch   map[string]*DomainRouter
	DomainPostfixSearch *pathtree.PathTree
	DomainPrefixSearch  *pathtree.PathTree
}

func NewDomainRouter(domain string) *DomainRouter {
	return &DomainRouter{
		Domain:               domain,
		LocationExactSearch:  make(map[string][]interface{}, 3),
		LocationPrefixSearch: pathtree.NewPathTree(),
		LocationRegexSearch:  make(map[string]*RegexTarget, 3),
	}
}

func (dm *DomainRouter) newCompileError(location string, err error) error {
	return fmt.Errorf("[dlrouter compile] Compile Error: %v. Domain: %s. Location: %s", err, dm.Domain, location)
}
func (dm *DomainRouter) AppendConf(dconf *DomainConf) []error {
	if dconf.Domain != dm.Domain {
		return []error{NotSameDomainErr}
	}

	errs := make([]error, 0, 2)

	for _, location := range dconf.Locations {
		location = strings.TrimSpace(location)

		if len(location) == 0 {
			continue
		}

		if strings.Index(location, "= ") == 0 {
			remain := strings.TrimSpace(location[2:])
			tlist, exist := dm.LocationExactSearch[remain]
			if exist {
				tlist = append(tlist, dconf.Target)
			} else {
				tlist = []interface{}{dconf.Target}
			}
			dm.LocationExactSearch[remain] = tlist
		} else if strings.Index(location, "~ ") == 0 {
			remain := strings.TrimSpace(location[2:])
			regexExp, err := regexp.Compile(remain)
			if err != nil {
				errs = append(errs, dm.newCompileError(location, err))
				continue
			} else {
				target, exist := dm.LocationRegexSearch[remain]
				if exist {
					target.Targets = append(target.Targets, dconf.Target)
				} else {
					target = &RegexTarget{
						RegexExp: regexExp,
						Targets: []interface{}{dconf.Target},
					}
					dm.LocationRegexSearch[remain] = target
				}
			}
		} else {
			err := dm.LocationPrefixSearch.Add(location, dconf.Target)
			if err != nil {
				errs = append(errs, dm.newCompileError(location, err))
			}
		}

	}
	return errs
}

func (dm *DomainRouter) GetTargetsForPath(path string, getAll bool) ([]*Target, bool) {
	targets := make([]*Target, 0, 1)
	//首先寻求精确匹配
	tlist, present := dm.LocationExactSearch[path]
	if present && len(tlist) > 0 {
		if !getAll {
			targets = append(targets, &Target{
				Value: tlist[0],
			})
			return targets, true
		}
		for _, t := range tlist {
			targets = append(targets, &Target{
				Value: t,
			})
		}
	}

	//前缀匹配
	pathBytes := []byte(path)
	if dm.LocationPrefixSearch.Size > 0 {
		candidates := dm.LocationPrefixSearch.GetCandidateLeafs(path)
		if len(candidates) > 0 {
			for _, candidate := range candidates {
				targets = append(targets, &Target{
					Value: candidate.Value,
					Variables: candidate.Variables,
				})
			}
			if !getAll {
				return targets, true
			}
		}
	}

	for _, regexTar := range dm.LocationRegexSearch {
		match := regexTar.RegexExp.Find(pathBytes)
		if match != nil {
			for _, t := range regexTar.Targets {
				targets = append(targets, &Target{
					Value: t,
				})
			}
			if !getAll {
				return targets, true
			}
		}
	}
	return targets, len(targets) > 0
}


func NewRouter(locationConfs []*LocationConf) (*DomainLocationRouter, []error) {
	domainExactSearch := make(map[string]*DomainRouter)

	allErrs := make([]error, 0, 3)
	for _, lconf := range locationConfs {
		if len(lconf.MappingConf) == 0 || lconf.Target == nil {
			continue
		}
		confs, err := GetDomainConfs(lconf)
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}

		for _, conf := range confs {
			if man, existed := domainExactSearch[conf.Domain]; existed {
				appErrs := man.AppendConf(conf)
				allErrs = append(allErrs, appErrs...)
			} else {
				newMan := NewDomainRouter(conf.Domain)
				appErrs := newMan.AppendConf(conf)
				allErrs = append(allErrs, appErrs...)
				domainExactSearch[conf.Domain] = newMan
			}
		}
	}

	ins := &DomainLocationRouter{
		DomainExactSearch:   domainExactSearch,
		DomainPostfixSearch: pathtree.NewPathTree(),
		DomainPrefixSearch:  pathtree.NewPathTree(),
	}

	for domain, man := range domainExactSearch {
		domainBytes := []byte(domain)
		pErr := ins.DomainPrefixSearch.Add(domain, man)
		if pErr != nil {
			allErrs = append(allErrs, pErr)
		}

		domainBytesRev := GetReversedBytes(domainBytes)
		rpErr := ins.DomainPostfixSearch.Add(string(domainBytesRev), man)
		if rpErr != nil {
			allErrs = append(allErrs, rpErr)
		}
	}

	return ins, allErrs
}

//Iterator Pattern using Closure
func (m *DomainLocationRouter) getDomainManagerIterator(domain string) func() (*DomainRouter, bool) {
	currentStage := 0
	stageIdx := 0
	stageCandidates := make([]*DomainRouter, 0, 2)
	iteredManagers := make(map[*DomainRouter]byte, domainSearchStageNum)

	return func() (*DomainRouter, bool) {
		GetNextInStage := func(stage int) (*DomainRouter, bool) {
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
					reversedDomain := string(GetReversedBytes(domainBytes))
					candidateTargets := m.DomainPostfixSearch.GetCandidateLeafs(reversedDomain)
					for _, t := range candidateTargets {
						dmm := t.Value.(*DomainRouter)
						stageCandidates = append(stageCandidates, dmm)
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
					candidateTargets := m.DomainPrefixSearch.GetCandidateLeafs(domain)
					for _, t := range candidateTargets {
						dmm := t.Value.(*DomainRouter)
						stageCandidates = append(stageCandidates, dmm)
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

		if currentStage >= domainSearchStageNum {
			return nil, false
		}

		var man *DomainRouter
		ok := false
		for man, ok = GetNextInStage(currentStage); currentStage < domainSearchStageNum; man, ok = GetNextInStage(currentStage) {
			if !ok {
				//upgrade stage
				currentStage++
				stageCandidates = make([]*DomainRouter, 0, 0)
				stageIdx = 0
			} else if _, itered := iteredManagers[man]; itered {
				continue
			} else {
				iteredManagers[man] = 1
				return man, ok
			}
		}
		return nil, false
	}
}


func (m *DomainLocationRouter) GetTarget(domain string, path string) (*Target, bool) {
	dmanIterator := m.getDomainManagerIterator(domain)

	for dm, present := dmanIterator(); present; dm, present = dmanIterator() {
		targets, matched := dm.GetTargetsForPath(path, false)
		if matched {
			return targets[0], true
		}
	}
	return nil, false
}

func (m *DomainLocationRouter) GetRouterInfosOfDomain(domain string) ([]*DomainRouter, bool) {
	routers := make([]*DomainRouter, 0, 1)

	dmanIter := m.getDomainManagerIterator(domain)
	for dm, present := dmanIter(); present; dm, present = dmanIter() {
		routers = append(routers, dm)
	}
	return routers, len(routers) > 0
}

func (m *DomainLocationRouter) GetAllRouterInfos() []*DomainRouter {
	routers := make([]*DomainRouter, 0, len(m.DomainExactSearch))

	for _, dm := range m.DomainExactSearch {
		routers = append(routers, dm)
	}
	return routers
}

func (m *DomainLocationRouter) GetAllTargets(domain string, path string) ([]*Target, bool) {
	dmanIterator := m.getDomainManagerIterator(domain)
	targets := make([]*Target, 0, 2)

	for dm, present := dmanIterator(); present; dm, present = dmanIterator() {
		tars, matched := dm.GetTargetsForPath(path, true)
		if matched {
			targets = append(targets, tars...)
		}
	}

	//remove duplicates
	targets = RemoveDuplicates(targets)

	return targets, len(targets) > 0
}
