package main

import (
	"testws/chat"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	hub := chat.NewHub()
	go hub.Run()
	r.GET("/ws", func(ctx *gin.Context) {
		chat.ServeWS(ctx, hub)
	})

	r.Run(":8080")

}
