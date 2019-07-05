package loadbalance

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	M "lbeng/models"
	lg "lbeng/pkg/logging"
	U "lbeng/pkg/utils"
)

//Handle handler
func Handle(c *gin.Context) {
	counter.incrTotalCounter()
	counter.incr("total client:" + c.ClientIP())

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
		edat := map[string]interface{}{
			"request":  usreq.Request,
			"status":   "dispatch failed",
			"comments": err.Error(),
		}
		bytesData, err := json.Marshal(edat)
		if err != nil {
			lg.Error(err.Error())
		}
		encryted := U.ECBEncrypt(bytesData)
		c.Data(http.StatusOK, "application/json; charset=UTF-8", encryted)
	}
	return
}
