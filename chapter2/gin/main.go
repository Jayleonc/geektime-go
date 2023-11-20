package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 如何定义路由，包括参数路由、通配符路由
// 如何处理输入输出
// 如何使用 中间件 解决 AOP 问题
func main() {
	server := gin.Default()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello My Love")
	})

	server.GET("/bye", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "GoodBye My Love")
	})

	server.Run(":8080")
}
