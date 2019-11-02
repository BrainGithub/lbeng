package func_device

import (
    "encoding/json"
    M "lbeng/models"
    lg "lbeng/pkg/logging"
    U "lbeng/pkg/utils"
    E "lbeng/pkg/e"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

func getProtocolSet(user *M.UserReq) []string {
    var info []string
    prots := make(map[string]bool)
    if len(user.Prots) > 1 {
        for _, p := range user.Prots {
            if !prots[p] {
                prots[p] = true
                info = append(info, p)
            }
        }
    }

    return info
}

func BuildMsgMultiProtocol(user *M.UserReq) (odat map[string]interface{}) {
    var info []interface{}
    arr := strings.Split(user.LoginName, ".")
    domain := arr[0]
    ip := ""
    prots := getProtocolSet(user)
    if len(prots) > 1 {
        for _, p := range prots {
            d := user.GetDescription(user.LoginName, user.ZoneID, p, ip)
            m := map[string]string{"domain": domain, "server": ip, "protocol": p, "description": d}
            info = append(info, m)
        }

        odat = map[string]interface{}{
            "request":                user.Request,
            "return":                 "ok",
            "code":                   E.SUCC,
            "auto_login_server_list": info,
            "user":                   user.LoginName,
            "zonename":               user.ZoneName,
            "comments":               "multi protocols, please make choice",
        }
        return
    }

    return nil
}

//BuildMsgAutoLoginServer back to client
func BuildMsgAutoLoginServer(user *M.UserReq) (odat map[string]interface{}) {
    odat = BuildMsgMultiProtocol(user)
    if odat != nil {
        return
    }

    var info []interface{}
    arr := strings.Split(user.LoginName, ".")
    domain := arr[0]

    if len(user.IPs) > 1 {
        for i, ip := range user.IPs {
            var p, d string
            if len(user.IPs) == len(user.Prots) {
                p = user.Prots[i]
            } else {
                p = user.Protocol
            }

            d = user.GetDescription(user.LoginName, user.ZoneID, p, ip)
            m := map[string]string{"domain": domain, "server": ip, "protocol": p, "description": d}
            info = append(info, m)
        }
    } else if len(user.Prots) > 1 {
        for i, p := range user.Prots {
            var ip, d string
            if len(user.IPs) == len(user.Prots) {
                ip = user.IPs[i]
            }

            d = user.GetDescription(user.LoginName, user.ZoneID, p, ip)
            m := map[string]string{"domain": domain, "server": ip, "protocol": p, "description": d}
            info = append(info, m)
        }
    }


    odat = map[string]interface{}{
        "request":                user.Request,
        "return":                 "ok",
        "auto_login_server_list": info,
        "code":                   E.SUCC,
        "user":                   user.LoginName,
        "zonename":               user.ZoneName,
        "comments":               "multi protocols, please make choice",
    }
    return
}

//OneBuildMsgAutoLoginServer back to client
func OneBuildMsgAutoLoginServer(user *M.UserReq) (gon bool, odat map[string]interface{}) {
    var info []interface{}
    arr := strings.Split(user.LoginName, ".")
    domain := arr[0]
    var prots []string
    var alsIP string

    if user.AutoLS.IP != "" && user.AutoLS.Prot != "" && user.AutoLS.Domain != "" {
        prots = append(prots, user.AutoLS.Prot)
        alsIP = user.AutoLS.IP
    } else {
        prots = user.Prots
    }

    pmap := make(map[string]bool)
    for _, p := range prots {
        if !pmap[p] {
            pmap[p] = true

            if alsIP != "" {
                ip := ""
                if p == "DPD-TM-Win" || p == "DPD-GRA-TM" {
                    ip = alsIP
                }

                desc := user.GetDescription(user.LoginName, user.ZoneID, p, ip)
                item := map[string]string{"domain": domain, "server": ip, "protocol": p, "description": desc}
                info = append(info, item)
            } else {
                ips := user.GetAutoLoginSvrIP(user.LoginName, user.ZoneID, p)
                if len(ips) == 0 {
                    ip := ""
                    desc := user.GetDescription(user.LoginName, user.ZoneID, p, ip)
                    item := map[string]string{"domain": domain, "server": ip, "protocol": p, "description": desc}
                    info = append(info, item)
                } else if len(ips) == 1 {
                    ip := ""
                    if p == "DPD-TM-Win" || p == "DPD-GRA-TM" {
                        ip = ips[0]
                    }
                    desc := user.GetDescription(user.LoginName, user.ZoneID, p, ip)
                    item := map[string]string{"domain": domain, "server": ip, "protocol": p, "description": desc}
                    info = append(info, item)
                } else {
                    for _, ip := range ips {
                        desc := user.GetDescription(user.LoginName, user.ZoneID, p, ip)
                        item := map[string]string{"domain": domain, "server": ip, "protocol": p, "description": desc}
                        info = append(info, item)
                    }
                }
            }
        }
    }

    odat = map[string]interface{}{
        "request":                user.Request,
        "return":                 "ok",
        "code":                   E.SUCC,
        "auto_login_server_list": info,
        "user":                   user.LoginName,
        "zonename":               user.ZoneName,
        "comments":               "multi protocols, please make choice",
    }

    l := len(info)
    if l == 1 {
        gon = true
    } else if l > 1 {
        gon = false
    } else {
        gon = false
        odat = BuildErrCodeMsgDisplay(user, E.ERR_CFG_NO_AVAIL_POOL)
    }
    return
}

//BuildMsgDisplay back to client
func BuildMsgDisplay(user *M.UserReq, msg string, ecode int) (odat map[string]interface{}) {
    code := ecode
    if code != E.SUCC {
        if strings.Contains(msg, "Client.Timeout exceeded") {
            msg = "Request timeout exceeded, please wait for a while"
            code = E.ERR_SYS_NW_TIMEOUT
        } else if strings.Contains(msg, "connection refused") {
            msg = "Connection refused, please double check the NetWork and Services"
            code = E.ERR_SYS_NW_CONN_REFUSED
        } else if strings.Contains(msg, "System License not inited") {
            code = E.ERR_LIC_NOT_INIT
        } else if strings.Contains(msg, "License is not valid or expired") {
            code = E.ERR_LIC_NOT_VALID
        } else if strings.Contains(msg, "License reach limit") {
            code = E.ERR_LIC_REACH_LIMIT
        } else if strings.Contains(msg, "available resource pool assigned") {
            code = E.ERR_CFG_NO_AVAIL_POOL
        } else if strings.Contains(msg, "Shared-VM client login configuration") {
            code = E.ERR_CFG_SHARED_VM
        } else if strings.Contains(msg, "Cluster is not stable") {
            code = E.ERR_CLUSTER_NOT_STABLE
        }
    }

    odat = map[string]interface{}{
        "request" : user.Request,
        "return"  : msg,
        "code"    : code,
        "user"    : user.LoginName,
        "zonename": user.ZoneName,
        "prot"    : user.Protocol,
        "comments": "",
    }
    return
}

//BuildErrCodeMsgDisplay
func BuildErrCodeMsgDisplay(user *M.UserReq, ecode int) (odat map[string]interface{}) {
    code := ecode
    msg := E.GetMsg(ecode)

    odat = map[string]interface{}{
        "request" : user.Request,
        "return"  : msg,
        "code"    : code,
        "user"    : user.LoginName,
        "zonename": user.ZoneName,
        "prot"    : user.Protocol,
        "comments": "",
    }
    return
}


//BuildAndReturnMsgUnfinishedLogging back to client
func BuildAndReturnMsgUnfinishedLogging(c *gin.Context, user *M.UserReq) {
    odat := map[string]interface{}{
        "request":  user.Request,
        "return":   "Current loggin has not finished, please wait",
        "code":     E.ERR_USR_LOGIN_UNFIN,
        "user":     user.LoginName,
        "zonename": user.ZoneName,
        "comments": "notice display",
    }
    bytesData, err := json.Marshal(odat)
    if err != nil {
        lg.Error(err.Error())
        return
    }
    encryted := U.ECBEncrypt(bytesData)
    lg.FmtInfo("%s", bytesData)
    c.Data(http.StatusOK, "application/json; charset=UTF-8", encryted)
}

//BuildAndReturnMsgUnfinishedLogging back to client
func BuildAndReturnMsg(c *gin.Context, user *M.UserReq, msg string) {
    odat := BuildMsgDisplay(user, msg, E.ERR_UNKNOWN)
    bytesData, err := json.Marshal(odat)
    if err != nil {
        lg.Error(err.Error())
        return
    }
    encryted := U.ECBEncrypt(bytesData)
    lg.FmtInfo("%s", bytesData)
    c.Data(http.StatusOK, "application/json; charset=UTF-8", encryted)
}

//BuildAndReturnMsg2 BuildAndReturnMsg2
func BuildAndReturnMsg2(c *gin.Context, idat map[string]interface{}) {
    bytesData, err := json.Marshal(idat)
    if err != nil {
        lg.Error(err.Error())
        return
    }
    encryted := U.ECBEncrypt(bytesData)
    lg.FmtInfo("%s", bytesData)
    c.Data(http.StatusOK, "application/json; charset=UTF-8", encryted)
}

//broker server build msg
//BuildMsg base method
func BuildMsg(br *M.BrokerRequest, idat map[string]interface{}) (odat map[string]interface{}) {
    //default output data
    odat = map[string]interface{}{
        "request":  br.Request,
        "return":   "ok",
        "user":     br.LoginName,
        "zonename": br.ZoneName,
        "code":     0,
        "comments": "notice display",
    }

    for k, v := range idat {
        odat[k] = v
    }

    return
}
