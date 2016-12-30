package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
	"github.com/shiyanhui/dht"
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

	w := dht.NewWire(65536, 1024, 256)
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
			for _, char := range name {
//				if (char > 128 && char < start) {
//					fmt.Printf("foreign torrent: %s\n", name)
//					data, err := json.Marshal(bt)
//					if err == nil {
//						fmt.Printf("%s\n\n", data)
//					}
//					break
				if (char >= start && char < end) {
					fmt.Printf("%s: found korean torrent: %s\n", time.Now().Format(time.RFC3339), name)
					data, err := json.Marshal(bt)
					if err == nil {
						fmt.Printf("%s\n\n", data)
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
