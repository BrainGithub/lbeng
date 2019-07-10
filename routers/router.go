package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	bs "lbeng/controller/broker"
	lb "lbeng/controller/loadbalance"
	"lbeng/routers/api"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.StaticFS("/upload/images", http.Dir(api.GetImageFullPath()))

	//main handler
	r.POST("/", lb.Handler)

	//for test
	broker := r.Group("/broker")
	//broker.Use(jwt.JWT())
	{
		broker.GET("/help", bs.Help)
		broker.GET("/debug", bs.Debug)
		broker.POST("/auth", api.GetAuth)
		broker.POST("/upload", api.UploadImage)

		broker.GET("/test", bs.Test)
		broker.GET("/racingtestvm", bs.RacingTestVM)
		broker.GET("/racingtestdocker", bs.RacingTestDocker)
		broker.GET("/racingtestsharevm", bs.RacingTestShareVM)
		broker.GET("/testdb", bs.TestDB)
	}

	return r
}
