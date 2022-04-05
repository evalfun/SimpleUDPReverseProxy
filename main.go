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

	r.GET("/api/udprelay/list", listUDPRelayHandler)
	r.POST("/api/udprelay/create", createUDPRelayHandler)
	r.POST("/api/udprelay/delete", deleteUDPRelayServerHandler)

	r.GET("/api/client/list", listAdvancedRelayClientHandler)
	r.POST("/api/client/create", createAdvancedRelayClientHandler)
	r.POST("/api/client/delete", deleteAdcancedRelayClientHandler)
	r.POST("/api/client/connection", getClientConnectionListHandler)
	r.POST("/api/client/connection/restart", restartClientConnectionHandler)
	r.POST("/api/client/serveraddr/update", updateClientServerAddrHandler)
	r.POST("/api/client/session", getClientSessionHandler)

	r.GET("/api/server/list", listAdvancedRelayServerHandler)
	r.POST("/api/server/create", createAdvancedRelayServerHandler)
	r.POST("/api/server/delete", deleteAdcancedRelayServerHandler)
	r.POST("/api/server/connection", getServerConnectionListHandler)
	r.POST("/api/server/session", getServerSessionHandler)
	r.POST("/api/server/connect", connectToClientHandler)

	r.GET("/api/config/save", saveConfigHandler)

	r.StaticFS("/static", assetFS())
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/static/")
	})

	if *runTimeinfo != "None" {
		go func() {
			log.Println("pprof start at ", *runTimeinfo)
			log.Println(http.ListenAndServe(*runTimeinfo, nil))
		}()
	}
	log.Printf("http api started at %s", *listenAddr)
	r.Run(*listenAddr)
}
