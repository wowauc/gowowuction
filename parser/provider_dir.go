package parser

import (
	"log"
	"path/filepath"
	"sort"

	util "github.com/wowauc/gowowuction/util"
)

/********************************************************************
 * DirectoryProvider
 ********************************************************************/

// provide entries (pfx-*.json.gz) from directory
type DirectoryProvider struct {
	entries map[string]string
}

func NewDirectoryProvider(dirname, prefix string) (provider *DirectoryProvider, err error) {
	mask := dirname + prefix + "-*.json.gz"
	log.Printf("scan by mask %s ...", mask)
	fnames, err := filepath.Glob(mask)
	if err != nil {
		log.Println("[!] glob failed:", err)
		return nil, err
	}
	log.Printf("... %d entries collected", len(fnames))

	provider = &DirectoryProvider{make(map[string]string)}

	for _, fname := range fnames {
		// realm, ts, good := util.Parse_FName(fname)
		_, _, good := util.Parse_FName(fname)
		if good {
			// log.Printf("fname %s -> %s, %v", fname, realm, ts)
			provider.entries[filepath.Base(fname)] = fname
		} else {
			// log.Printf("skip fname %s", fname)
		}
	}
	return
}

func (self *DirectoryProvider) List() (entries []string) {
	entries = make([]string, len(self.entries))
	i := 0
	for name, _ := range self.entries {
		entries[i] = name
		i++
	}
	sort.Sort(util.ByContent(entries))
	return
}

func (self *DirectoryProvider) Get(name string) (data []byte, err error) {
	if fname, ok := self.entries[name]; ok {
		data, err = util.Load(fname)
		return
	}
	return nil, ErrNotInList
}
