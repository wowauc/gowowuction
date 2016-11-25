package fetcher

import (
	"errors"
	//	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	config "github.com/wowauc/gowowuction/config"
)

type FDesc struct {
	Url string `json:"url"`
	Lmt int64  `json:"lastModified"`
}

type Rec0 struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type Rec1 struct {
	Files []FDesc `json:"files"`
}

type Session struct {
	Config *config.Config
	Client *http.Client
}

func (s *Session) Get(url string) (body []byte, err error) {
	err = nil
	if s.Client == nil {
		s.Client = new(http.Client)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[!] request not created: %s", url, err)
		return
	}
	request.Header.Add("Accept-Encoding", "gzip")
	log.Printf("GET %s", url)
	response, err := s.Client.Do(request)
	if err != nil {
		log.Printf("[!] request failed: %s", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		msg := fmt.Sprintf("status code %d != 200 : %s",
			response.StatusCode, response.Status)
		err = errors.New(msg)
		log.Printf("[!] " + msg)
		return
	}

	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			log.Printf("[!] gzip reader failed: %s", err)
			return
		}
		defer reader.Close()
	default:
		reader = response.Body
	}
	body, err = ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("[!] request read failed: %s", err)
		return
	}
	return
}

func (s *Session) Fetch_FileURL(realm string, locale string) (url string, ts time.Time, err error) {
	url = ""
	ts = time.Time{}
	err = nil
	v := strings.Split(realm, ":")
	if len(v) != 2 {
		msg := "realm is in bad format: '" + realm + "'"
		log.Println("[!] " + msg)
		err = errors.New(msg)
		return
	}
	var data []byte
	url = fmt.Sprintf("https://%s.api.battle.net/wow/auction/data/%s?locale=%s&apikey=%s",
		v[0], v[1], locale, s.Config.APIKey)
	data, err = s.Get(url)
	if err != nil {
		log.Printf("[!] GET request failed for %s ...", url)
		return
	}
	log.Println("parse auction file metainfo ...")

	var p0 Rec0
	if err = json.Unmarshal(data, &p0); err != nil {
		log.Printf("[!] json to R0 failed: %s", err)
		return
	}

	if p0.Status == "nok" {
		msg := fmt.Sprintf("realm=%v locale=%v returned status=%v reason=%v",
			realm, locale, p0.Status, p0.Reason)
		log.Printf("[!] " + msg)
		err = errors.New(msg)
		return
	}

	var p1 Rec1
	if err = json.Unmarshal(data, &p1); err != nil {
		log.Printf("[!] json failed: %s", err)
		return
	}

	if len(p1.Files) < 1 {
		log.Printf("thesre is no files (this is not an error)")
		return
	}

	url = p1.Files[0].Url
	lmt := p1.Files[0].Lmt
	ts = time.Unix(lmt/1000, lmt%1000).UTC()
	log.Printf("... url=%s, mtime=%s", url, ts)
	return
}
