package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}
