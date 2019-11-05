package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"

	"lbeng/models"
	"lbeng/pkg/logging"
	"lbeng/pkg/setting"
	"lbeng/routers"
	"lbeng/func_device"
)

func init() {
    setting.Setup()
    logging.Setup()
    models.Setup()

    func_device.InitSysLicense()
}

func main() {
    log.Printf("[info] main")

    gin.SetMode(setting.ServerSetting.RunMode)

    models.SetClusterCache(nil)

    routersInit := routers.InitRouter()
    readTimeout := setting.ServerSetting.ReadTimeout
    writeTimeout := setting.ServerSetting.WriteTimeout
    endPoint := fmt.Sprintf(":%s", setting.ServerSetting.HttpPort)

    server := &http.Server{
        Addr:         endPoint,
        Handler:      routersInit,
        ReadTimeout:  readTimeout,
        WriteTimeout: writeTimeout,
    }

    log.Printf("[info] start http server listening %s..", endPoint)

    // server.ListenAndServe()
    l, err := net.Listen("tcp4", endPoint)
    if err != nil {
        log.Fatal(err)
    }
    err = server.Serve(l)
    if err != nil {
        panic(err)
    }
}
