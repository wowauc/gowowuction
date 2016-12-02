package parser

import (
	"fmt"
	"log"

	config "github.com/wowauc/gowowuction/config"
	util "github.com/wowauc/gowowuction/util"
)

const TRIM_COUNT = 0

func ProcessSnapshot(ss *SnapshotData) {
	log.Printf("snapshot for %d auctions in %d realms",
		len(ss.Auctions), len(ss.Realms))
	log.Printf("  realms:")
	for _, realm := range ss.Realms {
		log.Printf("  name=%s, slug=%s", realm.Name, realm.Slug)
	}
	count := len(ss.Auctions)
	if TRIM_COUNT > 0 && TRIM_COUNT < count {
		count = TRIM_COUNT
	}
	log.Printf("  auctions: %d", count)
	for _, auc := range ss.Auctions {
		//log.Printf("raw=%#v", auc)
		blob := PackAuctionData(&auc)
		fmt.Println(string(blob))
		count--
		if count <= 0 {
			break
		}
	}
}

/******************************************************************************
 * 1. populate datafile locations
 *    from all datafiles (realm-*.json.gz)
 *    and from backup files (realm-*.zip)
 *    files from zip-files ase more important than ordinal datafiles
 * 2. create AuctionProcessor for realm
 * 3. process each entry from ordered list:
 *    read & parse content
 ******************************************************************************/

func ParseFromProvider(cf *config.Config, realm string, safe bool, prov Provider) {
	prc := new(AuctionProcessor)
	prc.Init(cf, realm)
	prc.LoadState()
	badfiles := make(map[string]string)

	for _, name := range prov.List() {
		//log.Println(fname)
		f_realm, f_time, ok := util.Parse_FName(name)
		if !ok {
			log.Printf("[!] not parsed correctly: %s", name)
			continue
		}
		if f_realm != realm {
			log.Printf("not my realm (%s != %s)", f_realm, realm)
			continue
		}
		if !prc.SnapshotNeeded(f_time) {
			log.Printf("snapshot not needed: %s", util.TSStr(f_time))
			continue
		}
		data, err := prov.Get(name)
		if err != nil {
			log.Printf("[!] %s LOAD ERROR: %s", name, err)
			badfiles[name] = fmt.Sprint(err)
			continue
		}
		ss, err := ParseSnapshot(data)
		if err != nil {
			log.Printf("[!] %s PARSE ERROR: %s", name, err)
			badfiles[name] = fmt.Sprint(err)
			continue
		}

		prc.StartSnapshot(f_time)
		for _, auc := range ss.Auctions {
			prc.AddAuctionEntry(&auc)
		}
		prc.FinishSnapshot()
		if safe {
			prc.SaveState()
		}
	}
	if !safe {
		prc.SaveState()
	}
	if len(badfiles) == 0 {
		log.Printf("all files loaded without errors")
	} else {
		log.Printf("%d files with errors", len(badfiles))
		for fname, err := range badfiles {
			log.Printf("%s: %s", fname, err)
		}
	}
}

func ParseDir(cf *config.Config, realm string, safe bool) {
	p, err := NewDirectoryProvider(cf.DownloadDirectory, util.Safe_Realm(realm))
	if err != nil {
		log.Printf("[!] create directory provider failed: %s", err)
		return
	}
	ParseFromProvider(cf, realm, safe, p)
}

func ParseComplex(cf *config.Config, realm string, safe bool) {
	prefix := util.Safe_Realm(realm)
	pzips, err := NewZipDirectoryProvider(cf.BackupDirectory, prefix)
	if err != nil {
		log.Printf("[!] create directory provider failed: %s", err)
		return
	}
	pdir, err := NewDirectoryProvider(cf.DownloadDirectory, prefix)
	if err != nil {
		log.Printf("[!] create directory provider failed: %s", err)
		return
	}
	p := NewCompositeProvider()
	p.Add(pzips)
	p.Add(pdir)
	ParseFromProvider(cf, realm, safe, p)
}
