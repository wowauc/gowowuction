package parser

import (
	"sort"

	util "github.com/wowauc/gowowuction/util"
)

/********************************************************************
 * CompositeProvider
 ********************************************************************/

type CompositeProvider struct {
	providers []*Provider          // all registered providers
	entries   map[string]*Provider // which fname provided
}

func NewCompositeProvider() *CompositeProvider {
	master := &CompositeProvider{}
	master.providers = []*Provider{}
	master.entries = make(map[string]*Provider)
	return master
}

func (self *CompositeProvider) Add(provider *Provider) {
	self.providers = append(self.providers, provider)
	entries := (*provider).List()
	for _, name := range entries {
		if _, ok := self.entries[name]; !ok {
			self.entries[name] = provider
		}
	}
}

func (self *CompositeProvider) List() (entries []string) {
	entries = make([]string, len(self.entries))
	i := 0
	for entry, _ := range self.entries {
		entries[i] = entry
		i++
	}
	sort.Sort(util.ByContent(entries))
	return
}

func (self *CompositeProvider) Get(name string) (data []byte, err error) {
	if provider, ok := self.entries[name]; ok {
		return (*provider).Get(name)
	}
	return
}
