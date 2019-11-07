package broker

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"

	vm "lbeng/devices"
	M "lbeng/models"
	lg "lbeng/pkg/logging"
	U "lbeng/pkg/utils"

	L "lbeng/pkg/logging"
)

var lger = L.GetLogrus()

//Handler : main portal
func Handler(c *gin.Context) {
	var err error
	var rawdata []byte
	var br M.BrokerRequest

	if rawdata, err = c.GetRawData(); err == nil {

		plainCtx := U.ECBDecrypt(rawdata)
		ioutil.WriteFile(debugFile, plainCtx, 0644) //for last connection debug

		if err = json.Unmarshal(plainCtx, &br); err == nil {
			doTask(c, plainCtx, &br)
		}
	}

	if err != nil {
		lg.Error(err.Error())
	}
}

//doTask do doTask ********************************************
func doTask(c *gin.Context, bytesCtx []byte, br *M.BrokerRequest) {
	i := vm.CreateInstance(c, bytesCtx, br)

	switch req, _ := i.Init(); req {
	case vm.ZONELIST:
		i.ZoneList()
	case vm.SFTP:
		// i.Sftp()
	case vm.SCREENUM:
		// i.Login()
	default:
		lger.Error("request error:%s", req)
	}
}
