package backup

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	//gzip "github.com/klauspost/compress/gzip"
	//gzip "github.com/klauspost/pgzip"
	//zip "github.com/klauspost/compress/zip"

	//	xz "github.com/danielrh/go-xz"
	xz "github.com/ulikunitz/xz"

	util "github.com/wowauc/gowowuction/util"
)

func tar_it(tarwriter *tar.Writer, data []byte, name string, ts time.Time) error {
	hdr := new(tar.Header)
	hdr.Name = name
	hdr.Size = int64(len(data))
	hdr.ModTime = ts
	hdr.Mode = 0644

	if err := tarwriter.WriteHeader(hdr); err != nil {
		return err
	}
	log.Printf("tar %d bytes for file %s", hdr.Size, hdr.Name)

	if _, err := tarwriter.Write(data); err != nil {
		return err
	}
	return nil
}

func MakeTarball(tarname string, fnames []string) error {
	log.Printf("tarring %d entrires to %s ...", len(fnames), tarname)
	tarfile, err := os.Create(tarname)
	if err != nil {
		return err
	}
	defer tarfile.Close()
	var tarwriter *tar.Writer
	if strings.HasSuffix(tarname, ".gz") {
		zipper := gzip.NewWriter(tarfile)
		defer zipper.Close()
		tarwriter = tar.NewWriter(zipper)
		/* remove or add trailing '/' after asterisk ---> */
		// for pure go'ish xz
		// xz "github.com/ulikunitz/xz.git"
	} else if strings.HasSuffix(tarname, ".xz") {
		zipper, err := xz.NewWriter(tarfile)
		if err != nil {
			return err
		}
		defer zipper.Close()
		tarwriter = tar.NewWriter(zipper)
		/**/
		/* remove or add trailing '/' after asterisk ---> *
			// xz "github.com/danielrh/go-xz"
			} else if strings.HasSuffix(tarname, ".xz") {
					zipper := xz.NewCompressionWriter(tarfile)
					defer zipper.Close()
					tarwriter = tar.NewWriter(zipper)
		/**/
	} else {
		tarwriter = tar.NewWriter(tarfile)
	}
	defer tarwriter.Close()
	var md5sum bytes.Buffer
	var sha1sum bytes.Buffer

	for _, fname := range fnames {
		realm, ts, good := util.Parse_FName(fname)
		if !good {
			log.Printf("warning: skip ill-named file '%s'", fname)
			continue // skip
		}
		data, err := util.Load(fname)
		if err != nil {
			return err
		}
		name := util.Make_FName(realm, ts, false)
		fmt.Fprintln(&md5sum, util.MakeMD5(data), name)
		fmt.Fprintln(&sha1sum, util.MakeSHA1(data), name)
		if err := tar_it(tarwriter, data, name, ts); err != nil {
			return err
		}
	}
	ts := time.Now()
	if err := tar_it(tarwriter, md5sum.Bytes(), "md5sum.txt", ts); err != nil {
		return err
	}
	if err := tar_it(tarwriter, sha1sum.Bytes(), "sha1sum.txt", ts); err != nil {
		return err
	}
	if err := tarwriter.Flush(); err != nil {
		return err
	}
	log.Printf("%s tarred without errors", tarname)
	return nil
}

func zip_it(zipwriter *zip.Writer, data []byte, name string, ts time.Time) error {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	header.SetModTime(ts)
	header.SetMode(0644)
	f, err := zipwriter.CreateHeader(header)

	if err != nil {
		return err
	}
	log.Printf("zip %d bytes for file %s", len(data), name)
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func MakeZip(zipname string, fnames []string) error {
	log.Printf("zipping %d entrires to %s ...", len(fnames), zipname)
	zipfile, err := os.Create(zipname)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	zipwriter := zip.NewWriter(zipfile)
	defer zipwriter.Close()

	var md5sum bytes.Buffer
	var sha1sum bytes.Buffer

	for _, fname := range fnames {
		realm, ts, good := util.Parse_FName(fname)
		if !good {
			log.Printf("warning: skip ill-named file '%s'", fname)
			continue // skip
		}
		data, err := util.Load(fname)
		if err != nil {
			return err
		}
		name := util.Make_FName(realm, ts, false)
		fmt.Fprintln(&md5sum, util.MakeMD5(data), name)
		fmt.Fprintln(&sha1sum, util.MakeSHA1(data), name)
		if err := zip_it(zipwriter, data, name, ts); err != nil {
			return err
		}
	}
	ts := time.Now()
	if err := zip_it(zipwriter, md5sum.Bytes(), "md5sum.txt", ts); err != nil {
		return err
	}
	if err := zip_it(zipwriter, sha1sum.Bytes(), "sha1sum.txt", ts); err != nil {
		return err
	}
	if err := zipwriter.Flush(); err != nil {
		return err
	}
	log.Printf("%s zipped without errors", zipname)
	return nil
}

func Backup(srcdir, dstdir, timeformat, ext string, completeOnly bool, doMove bool) {
	// Backup("/opt/wowauc/download", "/opt/wowauc/backup", "20060102", ".tar.gz")
	fnames, err := filepath.Glob(srcdir + "/*.json.gz")
	if err != nil {
		log.Fatalln("glob failed:", err)
	}
	log.Printf("... %d entries collected", len(fnames))

	rmap := make(map[string]map[string][]string)

	for _, fname := range fnames {
		realm, ts, good := util.Parse_FName(fname)
		if good {
			log.Printf("fname %s -> %s, %v", fname, realm, ts)
			rlm := util.Safe_Realm(realm)
			key := rlm + "-" + ts.Format(timeformat)
			if _, ok := rmap[rlm]; !ok {
				rmap[rlm] = make(map[string][]string)
			}
			if _, ok := rmap[rlm][key]; !ok {
				rmap[rlm][key] = make([]string, 0, 0)
			}
			rmap[rlm][key] = append(rmap[rlm][key], fname)
		} else {
			log.Printf("skip fname %s", fname)
		}
	}
	if completeOnly {
		log.Println("throw out last keys from every collected realm")
		for rlm, _ := range rmap {
			log.Printf("... for realm %s (%d entries)", rlm, len(rmap[rlm]))
			var keys []string
			for key, _ := range rmap[rlm] {
				keys = append(keys, key)
			}
			sort.Sort(util.ByContent(keys))
			sz := len(keys)
			lastkey := keys[sz-1]
			log.Printf("... ... remove %v", lastkey)
			delete(rmap[rlm], lastkey)
		}
	} else {
		log.Println("keep all keys")
	}

	var rlms []string
	for rlm, _ := range rmap {
		rlms = append(rlms, rlm)
	}
	sort.Sort(util.ByContent(rlms))

	for _, rlm := range rlms {
		var keys []string
		for key, _ := range rmap[rlm] {
			keys = append(keys, key)
		}
		sort.Sort(util.ByContent(keys))
		for _, key := range keys {
			fnames := rmap[rlm][key]
			log.Printf("backup %d entries for %s ...", len(fnames), key)
			sort.Sort(util.ByBasename(fnames))
			if ext == ".tar.gz" {
				tarname := dstdir + "/" + key + ext
				err := MakeTarball(tarname, fnames)
				if err != nil {
					log.Printf("[!] MakeTarball(%s) failed: %s", tarname, err)
					continue
				}
			} else if ext == ".tar.xz" {
				tarname := dstdir + "/" + key + ext
				err := MakeTarball(tarname, fnames)
				if err != nil {
					log.Printf("[!] MakeTarball(%s) failed: %s", tarname, err)
					continue
				}
			} else if ext == ".zip" {
				zipname := dstdir + "/" + key + ext
				err := MakeZip(zipname, fnames)
				if err != nil {
					log.Printf("[!] MakeZip(%s) failed: %s", zipname, err)
					continue
				}
			}
			if doMove {
				log.Printf("remove %d backed entries...", len(fnames))
				for _, fname := range fnames {
					if err := os.Remove(fname); err != nil {
						log.Printf("[!] remove(%s) failed: %s", fname, err)
					}
				}
			}
		}
	}
	return
}
