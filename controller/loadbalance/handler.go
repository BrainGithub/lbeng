package loadbalance

import (
	"net/http"

	"github.com/gin-gonic/gin"

	M "lbeng/models"
	lg "lbeng/pkg/logging"
	U "lbeng/pkg/utils"
)

//Handle handler
func Handle(c *gin.Context) {
	_do(c)
}

//_do
func _do(c *gin.Context) {
	var usreq M.UserReq

	rawdata, err := c.GetRawData()
	if err == nil {
		plainCtx := U.ECBDecrypt(rawdata)
		if err = M.UserReqMarshalAndVerify(plainCtx, &usreq); err == nil {
			err = dispatch(c, plainCtx, &usreq)
		}
	}

	if err != nil {
		lg.Error(err, err.Error())
		c.JSON(
			http.StatusOK,
			gin.H{
				"request":  usreq.Request,
				"status":   "dispatch failed",
				"comments": err.Error(),
			})
	}
	return
}
