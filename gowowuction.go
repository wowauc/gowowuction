package main

import (
	"io"
	"log"
	"os"
	"time"

	backup "github.com/wowauc/gowowuction/backup"
	config "github.com/wowauc/gowowuction/config"
	fetcher "github.com/wowauc/gowowuction/fetcher"
	parser "github.com/wowauc/gowowuction/parser"
	util "github.com/wowauc/gowowuction/util"
)

func DoFetch(cf *config.Config) {
	log.Println("=== FETCH BEGIN ===")
	s := new(fetcher.Session)
	s.Config = cf
	for _, realm := range cf.RealmsList {
		for _, locale := range cf.LocalesList {
			file_url, file_ts, err := s.Fetch_FileURL(realm, locale)
			if err != nil {
				log.Printf("[!] NO FILE URL FOR realm=%#v locale=%#v ", realm, locale)
				continue
			}
			if file_url == "" {
				log.Printf("[i] NO FILES FOR realm=%#v locale=%#v", realm, locale)
				continue
			}
			log.Printf("FILE URL: %s", file_url)
			log.Printf("FILE PIT: %s / %s", file_ts, util.TSStr(file_ts.UTC()))
			fname := util.Make_FName(realm, file_ts, true)
			json_fname := cf.DownloadDirectory + fname
			if !util.CheckFile(json_fname) {
				log.Printf("downloading from %s ...", file_url)
				data, err := s.Get(file_url)
				if err != nil {
					log.Printf("[!] DATA NOT RETRIEVED FOR realm=%#v locale=%#v", realm, locale)
					continue
				}
				log.Printf("... got %d octets", len(data))
				log.Printf("validate snapshot data ...")
				j, err := parser.ParseSnapshot(data)
				if err != nil {
					log.Printf("[!] json parse error: %s", err)
					continue
				}
				if j.Realms == nil {
					log.Println("[!] Realms is nil")
					continue
				}
				if len(j.Realms) == 0 {
					log.Println("[!] Realms is empty")
					continue
				}
				if j.Auctions == nil {
					log.Println("[!] Auctions is nil")
					continue
				}
				if len(j.Auctions) == 0 {
					log.Println("[!] Auctions is empty")
					continue
				}

				log.Printf("... data seems valid and contains %d auctions from %d realm(s).",
					len(j.Auctions), len(j.Realms))
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

func DoParse(cf *config.Config) {
	log.Println("=== PARSE BEGIN ===")
	for _, realm := range cf.RealmsList {
		parser.ParseDir(cf, realm, false)
	}
	log.Println("=== PARSE END ===")
}

func DoBackup(cf *config.Config) {
	log.Println("=== BACKUP BEGIN ===")
	srcdir := cf.DownloadDirectory
	dstdir := cf.BackupDirectory
	util.CheckDir(dstdir)
	//backup.Backup(srcdir, dstdir, "20060102", "", true, false)
	//backup.Backup(srcdir, dstdir, "20060102", ".tar.gz", true, false)
	//backup.Backup(srcdir, dstdir, "20060102", ".tar.xz", true, false)
	backup.Backup(srcdir, dstdir, "20060102", ".zip", true, false)
	log.Println("=== BACKUP END ===")
}

func main() {
	log.Println("preinitialize ...")
	cf, err := config.AppConfig()
	if err != nil {
		log.Fatalln("config load error: ", err)
	}
	util.CheckDir(cf.LogDirectory)
	logname := cf.GetLogFName(true)
	logf, err := os.OpenFile(logname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] log file %s not opened: %s", logname, err)
	} else {
		defer logf.Close()
		log.SetOutput(io.MultiWriter(logf, os.Stdout))
	}

	log.Println("=== application started at " + util.TSStr(time.Now()))
	cf.Dump()

	util.CheckDir(cf.DownloadDirectory)
	util.CheckDir(cf.ResultDirectory)
	util.CheckDir(cf.BackupDirectory)

	if len(os.Args) == 0 {
		DoFetch(cf)
	} else {
		for _, arg := range os.Args[1:] {
			switch arg {
			case "fetch":
				DoFetch(cf)
			case "parse":
				DoParse(cf)
			case "backup":
				DoBackup(cf)
			default:
				log.Printf("unknown arg: \"%s\"", arg)
			}
		}
	}
	log.Println("=== application finished at " + util.TSStr(time.Now()))
}
