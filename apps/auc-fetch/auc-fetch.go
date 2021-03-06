package main

import (
	"log"

	config "github.com/wowauc/gowowuction/config"
	fetcher "github.com/wowauc/gowowuction/fetcher"
	util "github.com/wowauc/gowowuction/util"
)

func DoFetch(cf *config.Config) {
	log.Println("=== FETCH BEGIN ===")
	s := new(fetcher.Session)
	s.Config = cf
	for _, realm := range cf.RealmsList {
		for _, locale := range cf.LocalesList {
			file_url, file_ts := s.Fetch_FileURL(realm, locale)
			log.Printf("FILE URL: %s", file_url)
			log.Printf("FILE PIT: %s / %s", file_ts, util.TSStr(file_ts.UTC()))
			fname := util.Make_FName(realm, file_ts, true)
			json_fname := cf.DownloadDirectory + fname
			if !util.CheckFile(json_fname) {
				log.Printf("downloading from %s ...", file_url)
				data := s.Get(file_url)
				log.Printf("... got %d octets", len(data))
				zdata := util.Zip(data)
				log.Printf("... zipped to %d octets (%d%%)",
					len(zdata), len(zdata)*100/len(data))
				util.Store(json_fname, zdata)
				log.Printf("stored to %s .", json_fname)
			} else {
				log.Println("... already downloaded")
			}
		}
	}
	log.Println("=== FETCH END ===")
}

func main() {
	log.Println("start")
	cf, err := config.AppConfig()
	if err != nil {
		log.Fatalln("config load error: ", err)
	}

	cf.Dump()

	util.CheckDir(cf.DownloadDirectory)
	util.CheckDir(cf.ResultDirectory)

	DoFetch(cf)
	log.Println("done")
}
