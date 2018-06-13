package dlrouter

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"
)

type targetData struct {
	ID int
}

var test0Conf = `- domains:
    - hotsoon.bytedance.com
    - hotsoon.toutiao.com
    - products.byted.org
  locations:
    - = /api/account/info
    - = /api/video/info
    - /api/user
    - ~ /api/video/detail/[0-9]+
    - /page/video/
    - /page/user/settings
    - ~ /page/user/profile/[0-9]+
    - /info/:version/group/:group_id/
    - /info/:version/item/:item_id/
    - /page/common
- domains:
    - api.hotsoon.com
    - api.hotsoon.org
    - hotsoon.byted.org
    - api.byted.org
    - products.byted.org
  locations:
    - = /api/hotsoon/account/auth
    - = /api/hotsoon/video/list
    - /api/hotsoon/video/comment
    - ~ /api/hotsoon/video/detail/[0-9]+
    - = /common/api/
    - /admin
`
var test1Conf = `- domains:
    - neihan.bytedance.com
    - neihan.toutiao.com
    - products.byted.org
  locations:
    - = /common/api/
    - = /api/account/info
    - = /api/video/info
    - /api/user
    - ~ /api/post/detail/[0-9]+
    - /page/post/
    - /page/user/settings
    - ~ /page/user/profile/[0-9]+
    - ~ /page/common/[0-9]+
- domains:
    - api.neihan.com
    - api-neihan.byted.org
    - api.byted.org
    - products.byted.org
  locations:
    - = /api/neihan/account/auth
    - = /api/neihan/video/list
    - /api/neihan/video/comment
    - ~ /api/neihan/video/detail/[0-9]+
    - ~ /api/neihan/post/comment/[0-9]+/[a-zA-Z]+
    - /admin
- domains:
    - 10.3.23.40
  locations:
    - = /wenda/web/feed/brow/
`
var testData = []*LocationConf{
	&LocationConf{
		Target:      1,
		MappingConf: getConfFromYaml(test0Conf),
	},
	&LocationConf{
		Target:      2,
		MappingConf: getConfFromYaml(test1Conf),
	},
}

func getConfFromYaml(yamlConf string) []*MappingBlock {
	var blocks []*MappingBlock
	err := yaml.Unmarshal([]byte(yamlConf), &blocks)
	if err != nil {
		fmt.Println(err)
	}
	return blocks
}

func getMappingManager() *DomainLocationRouter {
	m, _ := NewRouter(testData)
	return m
}

func sliceHas(slice []*Target, target interface{}) bool {
	for _, s := range slice {
		if target == s.Value {
			return true
		}
	}
	return false
}
func TestGetAllTarget(t *testing.T) {
	sm := getMappingManager()

	targets, exist := sm.GetAllTargets("products.byted.org", "/admin/accounts/delete")
	if !exist || len(targets) != 2 || !sliceHas(targets, 1) || !sliceHas(targets, 2) {
		t.Errorf("get all targets error. targets=%v", targets)
	}

	targets, exist = sm.GetAllTargets("products.byted.org", "/page/common/1234937432/3123")
	if !exist || len(targets) != 2 || !sliceHas(targets, 1) || !sliceHas(targets, 2) {
		t.Errorf("get all targets error. targets=%v", targets)
	}

	targets, exist = sm.GetAllTargets("products.byted.org", "/page/common/tt1234937432/3123")
	if !exist || len(targets) != 1 || !sliceHas(targets, 1) {
		t.Errorf("get all targets error. targets=%v", targets)
	}
	targets, exist = sm.GetAllTargets("products.byted.org", "/common/api/")
	if !exist || len(targets) != 2 || !sliceHas(targets, 1) || !sliceHas(targets, 2) {
		t.Errorf("get all targets error. targets=%v", targets)
	}
}
func TestGetTarget(t *testing.T) {
	sm := getMappingManager()

	target, exist := sm.GetTarget("api.neihan.com", "/api/neihan/video/detail/123435345")
	if !exist || target.Value != 2 {
		t.Errorf("get target error. expected: %v %v; got: %v %v", true, 2, exist, target)
	}

	target, exist = sm.GetTarget("api-hotsoon.byted.org", "/api/hotsoon/video/comment/avbasdfaskdfsdf/12345")
	if !exist || target.Value != 1 {
		t.Errorf("get target error. expected: %v %v; got: %v %v", true, 1, exist, target)
	}

	target, exist = sm.GetTarget("api.neihan.com", "/api/neihan/post/comment/123445/sdfjklHUIIHJFEewfsdfSDSDF")
	if !exist || target.Value != 2 {
		t.Errorf("get target error. expected: %v %v; got: %v %v", true, 2, exist, target)
	}

	target, exist = sm.GetTarget("products.byted.org", "/page/video/sdfsdfweruFHUIER/1")
	if !exist || target.Value != 1 {
		t.Errorf("get target error. expected: %v %v; got: %v %v", true, 1, exist, target)
	}

	target, exist = sm.GetTarget("products.byted.org", "/page/post/sdfsdfweruFHUIER/1")
	if !exist || target.Value != 2 {
		t.Errorf("get target error. expected: %v %v; got: %v %v", true, 2, exist, target)
	}

	target, exist = sm.GetTarget("products.byted.org", "/page/postit/sdfsdfweruFHUIER/1")
	if exist {
		t.Errorf("get target error. expected: %v %v; got: %v %v", false, nil, exist, target)
	}
	target, exist = sm.GetTarget("10.3.23.40:9009", "/wenda/web/feed/brow/")

	if !exist {
		t.Errorf("get target error.")
	}

	target, exist = sm.GetTarget("products.byted.org", "/info/4/group/12345/comments/")

	if !exist  || target.Value != 1 || target.Variables["version"] != "4" || target.Variables["group_id"] != "12345" {
		t.Errorf("get target error.")
	}
	target, exist = sm.GetTarget("products.byted.org", "/info/3/item/123454321/comments/")

	if !exist || target.Value != 1 || target.Variables["version"] != "3" || target.Variables["item_id"] != "123454321" {
		t.Errorf("get target error: %v", *target)
	}
}

func BenchmarkGetSceneRegex(b *testing.B) {
	sm := getMappingManager()
	for i := 0; i < b.N; i++ {
		sm.GetTarget("api.neihan.com", "/api/neihan/post/comment/123445/sdfjklHUIIHJFEewfsdfSDSDF")
	}
}

func BenchmarkGetScenePrefix(b *testing.B) {
	sm := getMappingManager()
	for i := 0; i < b.N; i++ {
		sm.GetTarget("products.byted.org", "/page/post/sdfsdfweruFHUIER/1")
	}
}

func BenchmarkGetSceneMissed(b *testing.B) {
	sm := getMappingManager()
	for i := 0; i < b.N; i++ {
		sm.GetTarget("products.byted.org", "/page/postit/sdfsdfweruFHUIER/1")
	}
}
