package main

import "github.com/gin-gonic/gin"

func setURL(r *gin.Engine) {
	r.GET("/api/udprelay/list", listUDPRelayHandler)
	r.POST("/api/udprelay/create", createUDPRelayHandler)
	r.POST("/api/udprelay/delete", deleteUDPRelayServerHandler)
	// Client
	r.GET("/api/client/list", listAdvancedRelayClientHandler)
	r.POST("/api/client/create", createAdvancedRelayClientHandler)
	r.POST("/api/client/delete", deleteAdcancedRelayClientHandler)
	r.POST("/api/client/connection", getClientConnectionListHandler)
	r.POST("/api/client/connection/restart", restartClientConnectionHandler)
	r.POST("/api/client/serveraddr/update", updateClientServerAddrHandler)
	r.POST("/api/client/session", getClientSessionHandler)

	r.POST("/api/client/tracker", getClientTrackerHandler)
	r.POST("/api/client/tracker/set", setClientTrackerHandler)
	//Server
	r.GET("/api/server/list", listAdvancedRelayServerHandler)
	r.POST("/api/server/create", createAdvancedRelayServerHandler)
	r.POST("/api/server/delete", deleteAdcancedRelayServerHandler)
	r.POST("/api/server/connection", getServerConnectionListHandler)
	r.POST("/api/server/session", getServerSessionHandler)
	r.POST("/api/server/connect", connectToClientHandler)

	r.POST("/api/server/tracker", getServerTrackerHandler)
	r.POST("/api/server/tracker/set", setServerTrackerHandler)
	//Config
	r.GET("/api/config/save", saveConfigHandler)

	//r.Static("/static", "static")
	r.StaticFS("/static", assetFS())
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/static/")
	})
}
