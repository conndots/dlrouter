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

type domainConf struct {
	Domain    string
	Locations []string
	Target    interface{}
}

func getDomainConfs(conf *LocationConf) ([]*domainConf, error) {
	blocks := conf.MappingConf
	confs := make([]*domainConf, 0, len(blocks)*3/2)
	for _, block := range blocks {
		for _, domain := range block.Domains {
			confs = append(confs, &domainConf{
				Domain:    domain,
				Locations: block.Locations,
				Target:    conf.Target,
			})
		}
	}
	return confs, nil
}
