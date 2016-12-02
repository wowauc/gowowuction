package parser

import (
	"log"
	"path/filepath"
	"sort"

	util "github.com/wowauc/gowowuction/util"
)

/********************************************************************
 * ZipDirectoryProvider
 ********************************************************************/

func NewZipDirectoryProvider(dirname string, prefix string) (prov *CompositeProvider, err error) {
	mask := dirname + prefix + "-*.zip"
	log.Printf("collect zipfiles by mask: %s", mask)
	zipfnames, err := filepath.Glob(mask)

	if err != nil {
		log.Println("[!] glob failed:", err)
		return nil, err
	}
	sort.Sort(util.ByBasename(zipfnames))
	prov = NewCompositeProvider()
	log.Printf("... %d zipfiles found", len(zipfnames))
	for _, zipfname := range zipfnames {
		zp, err := NewZipProvider(zipfname)
		if err != nil {
			log.Printf("[!] ZipProvider(%v) not created: error %s", zipfname, err)
			continue
		}
		prov.Add(zp)
	}
	return prov, nil
}
