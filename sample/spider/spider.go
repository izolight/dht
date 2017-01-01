package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"log"
	"github.com/izolight/dht"
	"net/http"
	_ "net/http/pprof"
)

type file struct {
	Path   []interface{} `json:"path"`
	Length int           `json:"length"`
}

type bitTorrent struct {
	InfoHash string `json:"infohash"`
	Name     string `json:"name"`
	Files    []file `json:"files,omitempty"`
	Length   int    `json:"length,omitempty"`
}

func main() {
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	const start = '가'
	const end = '힣'
	logfile, _ := os.OpenFile("./log.txt", os.O_RDWR|os.O_APPEND, 0660)
	logger := log.New(logfile, "DHT Spider: ", log.LstdFlags|log.Lshortfile)

	w := dht.NewWire(65536, 4096, 2048)
	go func() {
		for resp := range w.Response() {
			metadata, err := dht.Decode(resp.MetadataInfo)
			if err != nil {
				continue
			}
			info := metadata.(map[string]interface{})

			if _, ok := info["name"]; !ok {
				continue
			}

			bt := bitTorrent{
				InfoHash: hex.EncodeToString(resp.InfoHash),
				Name:     info["name"].(string),
			}

			if v, ok := info["files"]; ok {
				files := v.([]interface{})
				bt.Files = make([]file, len(files))

				for i, item := range files {
					f := item.(map[string]interface{})
					bt.Files[i] = file{
						Path:   f["path"].([]interface{}),
						Length: f["length"].(int),
					}
				}
			} else if _, ok := info["length"]; ok {
				bt.Length = info["length"].(int)
			}

			name := bt.Name
					data, err := json.Marshal(bt)
					if err == nil {
						log.Printf("%s", data)
						logger.Printf("%s", data)
					}
			for _, char := range name {
//				if (char > 128 && char < start) {
//					fmt.Printf("foreign torrent: %s\n", name)
//					data, err := json.Marshal(bt)
//					if err == nil {
//						fmt.Printf("%s\n\n", data)
//					}
//					break
				if (char >= start && char < end) {
					data, err := json.Marshal(bt)
					if err == nil {
						log.Printf("%s", data)
						logger.Printf("%s", data)
					}
					break
				}
			}
		}
	}()
	go w.Run()

	config := dht.NewCrawlConfig()
	config.OnAnnouncePeer = func(infoHash, ip string, port int) {
		w.Request([]byte(infoHash), ip, port)
	}
	d := dht.New(config)

	d.Run()
}
