//go:generate go-bindata-assetfs  static/...
package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
)

var configFilePath string

func main() {

	listenAddr := flag.String("l", "127.0.0.1:3480", "http api listen addr.")
	configfile := flag.String("c", "config.json", "config file path")
	runTimeinfo := flag.String("r", "None", "Golang pprof tools.Default disabled.")
	flag.Parse()

	configFilePath = *configfile

	loadConfig()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	setURL(r)
	if *runTimeinfo != "None" {
		go func() {
			log.Println("pprof start at ", *runTimeinfo)
			log.Println(http.ListenAndServe(*runTimeinfo, nil))
		}()
	}
	log.Printf("http api started at %s", *listenAddr)
	r.Run(*listenAddr)
}
