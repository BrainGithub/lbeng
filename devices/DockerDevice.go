package devices

import (
    M "lbeng/models"
    E "lbeng/pkg/e"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
    U "lbeng/pkg/utils"
)

type DKDevice struct {
    base     *DPDevice
    destNode string
}

func (dk *DKDevice) Init() (string, string) {
    return dk.base.Init()
}

//ZoneListDispatch ZoneListDispatch
func (dk *DKDevice) ZoneListDispatch() {
    dk.base.ZoneListDispatch()
}

//SftpDispatch request sftp
func (dk *DKDevice) SftpDispatch() {
    base := dk.base
    _, err := dk.alloc()
    if err != nil {
        base.errorProcess(err, E.ERR_USR_SFTP)
        return
    }

    base.sftpDispatch()
}

//LoginDispatch request screennum
func (dk *DKDevice) LoginDispatch() {
    ur := dk.base.ur
    if len(ur.Prots) == 0 {
        eCode := E.ERR_CFG_NO_AVAIL_POOL
        dk.base.errorProcess(nil, eCode)
    }
    dk.reqScreenNum()
}

func (dk *DKDevice) reqScreenNum() {
    dk.dispatch()
    return
}

func (dk *DKDevice) alloc() (ip string, err error) {
    base := dk.base
    ur := dk.base.ur

    if M.PublicNetworkDetect() {
        lg.Info("do PublicNetworkDetect")
        if U.IsPublicNetwork(ur.ClientIP) {
            lg.Info("is PublicNetwork do default redirect")
            dk.destNode = S.AppSetting.DefaultRedirectHost
            ip = dk.destNode
            return
        }
    }

    found, err := InnerDockerLeastConn(ur)
    if err != nil {
        return
    }

    if !found {
        dk.destNode = S.AppSetting.DefaultRedirectHost
    } else {
        dk.destNode = base.ur.IPs[0]
    }

    ip = dk.destNode
    return
}

func (dk *DKDevice) sftpAlloc() (ip string, err error) {
    ip = S.AppSetting.DefaultRedirectHost
    ur := dk.base.ur
    if len(ur.IPs) > 0 {
        ip = ur.IPs[0]
    }
    return ip, nil
}

//Dispatch Dispatch
func (dk *DKDevice) Dispatch() {
    dk.dispatch()
}

func (dk *DKDevice) dispatch() {
    base := dk.base
    ip, err := dk.alloc()
    if err != nil {
        lg.Error(err)
        return
    }
    lg.Info(ip)
    base.nodeIP = ip
    base.doDispatch()
}

func (dk *DKDevice) makeResponse() {
    base := dk.base
    base.makeResponse(nil)
}

func (dk *DKDevice) ZoneList() {
    dk.base.ZoneList()
}
