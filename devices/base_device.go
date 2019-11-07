package devices

import (
    "encoding/json"
    "github.com/gin-gonic/gin"
    M "lbeng/models"
    E "lbeng/pkg/e"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
    U "lbeng/pkg/utils"
    "net/http"
    "strings"
)

//IDPDevice interface has no-return, because when errored, there is web response
type IDPDevice interface {
    Init() (string, string)

    ZoneListDispatch()
    SftpDispatch()
    LoginDispatch()
    Dispatch()

    ZoneList()
    // Sftp()
    // Login()
    // Logout()
    // Restart()
    // ReCreate()
    // Supend()
    // Resume()
    // Delete()
}

type DPDevice struct {
    c   *gin.Context
    ctx []byte

    ur *M.UserReq
    br *M.BrokerRequest

    prot string
    req  string

    nodeIP string
}

func (dev *DPDevice) Init() (string, string) {
    if dev.ur != nil {
        dev.prot = dev.ur.Protocol
        dev.req = dev.ur.Request
    } else {
        dev.prot = dev.br.Protocol
        dev.req = dev.br.Request
    }

    return dev.req, dev.prot
}

//ZoneListDispatch ZoneListDispatch
func (dev *DPDevice) ZoneListDispatch() {
    dev.nodeIP = S.AppSetting.DefaultRedirectHost
    dev.ur.IPs = []string {dev.nodeIP}
    dev.Dispatch()
}

//SftpDispatch sftp to dest ip addr
func (dev *DPDevice) SftpDispatch() {
    dev.sftpDispatch()
}

func (dev *DPDevice) sftpDispatch() {
    ip := S.AppSetting.DefaultRedirectHost
    ur := dev.ur
    if len(ur.IPs) > 0 {
        ip = ur.IPs[0]
    }

    dev.nodeIP = ip
    lg.Info(dev.nodeIP)
    dev.doDispatch()
}

func (dev *DPDevice) LoginDispatch() {
    if gon := dev.preLogin(); !gon {
        return
    }

    dev.doDispatch()
}

func (dev *DPDevice) preLogin() bool {
    gon, dat := OneBuildMsgAutoLoginServer(dev.ur)
    if !gon {
        BuildAndReturnMsg2(dev.c, dat)
        return false
    }

    return true
}

//Dispatch default dispatch
func (dev *DPDevice) Dispatch() {
    dev.doDispatch()
}

func (dev *DPDevice) doDispatch() {
    err := doDispatch(dev.c, dev.ctx, dev.ur)
    if err != nil {
        ecode := E.ERR_UNKNOWN
        if strings.Contains(err.Error(), "Client.Timeout exceeded") {
            ecode = E.ERR_SYS_NW_TIMEOUT
        } else if strings.Contains(err.Error(), "connection refused") {
            ecode = E.ERR_SYS_NW_CONN_REFUSED
        }
        dev.errorProcess(err, ecode)
    }
    decrAllocCounter(dev.ur)
}

//broker request ---------------------------------------------------------

func (dev *DPDevice) ZoneList() {
    zonelist, err := dev.br.GetZoneList()
    if err != nil {
        dev.errorProcess(err, E.ERR_USR_ZONE_MISMATCH)
    }

    load := map[string]interface{}{
        "zonelist": zonelist,
    }

    dev.makeResponse(BuildMsg(dev.br, load))
}

//ErrorProcess error process
//code = 0, success
//code < 0, error, need process
func (dev *DPDevice) errorProcess(err error, code int) {
    lg.FmtError("error code:%d, error:%+v", code, err)
    E.Debug(err)
    msg := BuildMsgDisplay(dev.ur, err.Error(), code)
    dev.makeResponse(msg)
}

//ErrProc ErrProc
func (dev *DPDevice) ErrProc(err error, code int) {
    dev.errorProcess(err, code)
}

//MakeResponse make response
func (dev *DPDevice) makeResponse(v interface{}) {
    bytesData, err := json.Marshal(v)
    if err != nil {
        lg.Error(err.Error())
        bytesData = []byte(err.Error())
    }
    lg.FmtInfo("response:%s", bytesData)
    encryted := U.ECBEncrypt(bytesData)
    dev.c.Data(http.StatusOK, "application/json; charset=UTF-8", encryted)
}

//makeTaskResponse make task response, task is for broker server
func (dev *DPDevice) makeBrokerResponse(reqRst interface{}, errCode int) {
    br := dev.br
    var odat map[string]interface{}

    switch reqRst.(type) {
    case string:
        reqRst = reqRst.(string)
    case []byte:
        reqRst = reqRst.([]byte)
    case map[string]interface{}:
        odat = reqRst.(map[string]interface{})
    }

    odat["user"] = br.LoginName
    odat["zonename"] = br.ZoneName
    odat["protocol"] = br.Protocol
    odat["response"] = br.Request
    odat["code"] = errCode
    odat["return"] = E.GetMsg(errCode)

    dev.makeResponse(odat)
}

//CreateInstance CreateInstance
func CreateInstance(c *gin.Context, bytesCtx []byte, req interface{}) (ivm IDPDevice) {
    base := &DPDevice{
        c:   c,
        ctx: bytesCtx,
        ur:  nil,
        br:  nil,
    }
    switch req.(type) {
    case *M.UserReq:
        base.ur = req.(*M.UserReq)
    case *M.BrokerRequest:
        base.br = req.(*M.BrokerRequest)
    }

    switch _, p := base.Init(); p {
    case DPD_DK:
        ivm = &DKDevice{base: base}
    case DPD_LIN:
        ivm = &LinuxDevice{base: &InnerVMDevice{base: base}}
    case DPD_WIN, DPD_WIN_SVR:
        ivm = &WinDevice{base: &InnerVMDevice{base: base}}
    case DPD_TM_WIN:
        ivm = &TMWinDevice{base: base}
    case DPD_TM_LIN:
        ivm = &TMLinDevice{base: base}
    default:
        ivm = base
        lg.Warn("default instance, prot:", p)
    }
    return
}
