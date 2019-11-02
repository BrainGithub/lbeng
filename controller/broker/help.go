package broker

import (
    "net/http"

    "github.com/gin-gonic/gin"

    lg "lbeng/pkg/logging"
)

//Help can be used as alive test
func Help(c *gin.Context) {
    c.JSON(
        http.StatusOK,
        gin.H{
            "urls":    "/  /lb/debug  /lb/getport  /broker/help  /broker/debug  /broker/test/mloginlinux  /broker/test/mlogindocker  /broker/test/mloginwin",
            "status":  http.StatusOK,
        })

    lg.Info()
}
