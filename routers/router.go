package routers

import (
    "github.com/gin-gonic/gin"
    "lbeng/controller/port_pre_alloc"

    bs "lbeng/controller/broker"
    lb "lbeng/controller/loadbalance"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())

    //load balance handler portal
    {
        r.POST("/", lb.Handler)
        r.GET("/", bs.Help)

        lbgroup := r.Group("/lb")
        {
            lbgroup.GET("/debug", lb.Debug)
            lbgroup.POST("/getport", port_pre_alloc.GetPort)
            //lbgroup.POST("/checklic", lb.CheckLicense)
        }
    }

    //broker handler portal
    {
        broker := r.Group("/broker")
        //broker.Use(jwt.JWT())
        {
            broker.POST("/", bs.Handler)
            broker.GET("/help", bs.Help)
            broker.GET("/debug", bs.Debug)

            broker.GET("/test/mloginlinux", bs.TestMultiLoginVMLinux)
            broker.POST("/test/mloginlinux", bs.TestMultiLoginVMLinux)

            broker.GET("/test/mlogindocker", bs.TestMultiLoginVMDocker)
            broker.POST("/test/mlogindocker", bs.TestMultiLoginVMDocker)

            broker.GET("/test/mloginwin", bs.TestMultiLoginVMWin)
            broker.POST("/test/mloginwin", bs.TestMultiLoginVMWin)

            broker.GET("/test/racingtestvm", bs.RacingTestVM)
            broker.GET("/test/racingtestdocker", bs.RacingTestDocker)
            broker.GET("/test/racingtestsharevm", bs.RacingTestShareVM)
            broker.GET("/test/db", bs.TestDB)
        }
    }

    return r
}
