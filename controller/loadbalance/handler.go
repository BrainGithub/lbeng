package loadbalance

import (
    "github.com/gin-gonic/gin"
    "io/ioutil"
    "time"

    M "lbeng/models"
    lg "lbeng/pkg/logging"
    U "lbeng/pkg/utils"

    vm "lbeng/func_device"
)

var cnt = U.GetCounter()


//Handler portal
func Handler(c *gin.Context) {
    cnt.Incr("TotalConn")
    cnt.Incr("FromClient:" + c.ClientIP())

    do(c)

    cnt.Log("cnt")
}

//do, actual proc
func do(c *gin.Context) {
    var usreq M.UserReq
    var rawdata []byte
    var err error

    tstart := time.Now()

    if rawdata, err = c.GetRawData(); err != nil {
        lg.Error(err.Error())
        vm.BuildAndReturnMsg(c, &usreq, err.Error())
        return
    }
    plainCtx := U.ECBDecrypt(rawdata)
    err = ioutil.WriteFile(debugFile, plainCtx, 0644) //for last connection debug
    if err != nil {
        lg.Info(err.Error())
    }

    if err = M.UserReqMarshalAndVerify(plainCtx, &usreq); err != nil {
        vm.BuildAndReturnMsg(c, &usreq, err.Error())
        return
    }

    usreq.ClientIP = c.ClientIP()
    doLoadBalance(c, plainCtx, &usreq)

    elapsed := time.Since(tstart)
    lg.FmtInfo("request elapsed:%s,%d,%s,%s,%d ms",
        usreq.LoginName,
        usreq.ZoneID,
        usreq.Protocol,
        usreq.Request,
        elapsed/time.Millisecond)
    return
}

//doLoadBalance do load balance ********************************************
func doLoadBalance(c *gin.Context, bytesCtx []byte, ur *M.UserReq) {
    i := vm.CreateInstance(c, bytesCtx, ur)

    switch req, _ := i.Init(); req {
    case vm.VERSION, vm.ZONELIST:
        i.ZoneListDispatch()
    case vm.SFTP:
        i.SftpDispatch()
    case vm.SCREENUM:
        fallthrough
    default:
        i.LoginDispatch()
    }
}
