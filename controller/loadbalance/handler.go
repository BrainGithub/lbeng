package loadbalance

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"

	M "lbeng/models"
	lg "lbeng/pkg/logging"
	U "lbeng/pkg/utils"
)

//Handle handler
func Handler(c *gin.Context) {
	counter.incrTotalCounter()
	counter.incr("FromClient:" + c.ClientIP())
	lg.FmtInfo("EntranceStat:%+v", *counter)

	_do(c)
}

//_do
func _do(c *gin.Context) {
	var usreq M.UserReq

	rawdata, err := c.GetRawData()
	if err == nil {
		plainCtx := U.ECBDecrypt(rawdata)
		ioutil.WriteFile(".debug.json", plainCtx, 0644) //for last connection debug

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
