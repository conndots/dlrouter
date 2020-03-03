package dlrouter

type pathConfType string

const (
	pathConfTypePrefix = ""
	pathConfTypeRegex  = "~"
	pathConfTypeEqual  = "="
)

type LocationConf struct {
	Target      interface{}
	MappingConf []*MappingBlock
}

type MappingBlock struct {
	Domains   []string `yaml:"domains" json:"domains"`
	Locations []string `yaml:"locations,omitempty" json:"locations,omitempty"`
}

type DomainConf struct {
	Domain    string
	Locations []string
	Target    interface{}
}

func GetDomainConfs(conf *LocationConf) ([]*DomainConf, error) {
	blocks := conf.MappingConf
	confs := make([]*DomainConf, 0, len(blocks)*3/2)
	for _, block := range blocks {
		for _, domain := range block.Domains {
			confs = append(confs, &DomainConf{
				Domain:    domain,
				Locations: block.Locations,
				Target:    conf.Target,
			})
		}
	}
	return confs, nil
}
