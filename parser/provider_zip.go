package parser

import (
	"archive/zip"
	"io/ioutil"
	"path/filepath"
	"sort"

	util "github.com/wowauc/gowowuction/util"
)

/********************************************************************
 * ZipProvider
 ********************************************************************/

// provide entries from zip archive
type ZipProvider struct {
	reader  *zip.ReadCloser
	entries map[string]int
}

func NewZipProvider(zipname string) (provider *ZipProvider, err error) {
	reader, err := zip.OpenReader(zipname)
	if err != nil {
		return nil, err
	}
	entries := make(map[string]int)
	for ix, f := range reader.File {
		name := filepath.Base(f.Name)
		if _, _, ok := util.Parse_FName(name); ok {
			entries[name] = ix
		}
	}
	prov := &ZipProvider{reader, entries}
	return prov, nil
}

func (self *ZipProvider) List() (entries []string) {
	entries = make([]string, len(self.entries))
	i := 0
	for name, _ := range self.entries {
		entries[i] = name
		i++
	}
	sort.Sort(util.ByContent(entries))
	return
}

func (self *ZipProvider) Get(name string) (data []byte, err error) {
	ix, ok := self.entries[name]
	if !ok {
		return nil, ErrNotInList
	}
	f, err := self.reader.File[ix].Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err = ioutil.ReadAll(f)
	return data, err
}
