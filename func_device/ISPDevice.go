package func_device

import (
    M "lbeng/models"
    E "lbeng/pkg/e"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
    U "lbeng/pkg/utils"
)

type ISPDevice struct {
    base     *DPDevice
    destNode string
}

func (isp *ISPDevice) Init() (string, string) {
    return isp.base.Init()
}

//ZoneListDispatch ZoneListDispatch
func (isp *ISPDevice) ZoneListDispatch() {
    isp.base.ZoneListDispatch()
}

//SftpDispatch request sftp
func (isp *ISPDevice) SftpDispatch() {
    base := isp.base
    _, err := isp.alloc()
    if err != nil {
        base.errorProcess(err, E.ERR_USR_SFTP)
        return
    }

    base.sftpDispatch()
}

//LoginDispatch request screennum
func (isp *ISPDevice) LoginDispatch() {
    ur := isp.base.ur
    if len(ur.Prots) == 0 {
        eCode := E.ERR_CFG_NO_AVAIL_POOL
        isp.base.errorProcess(nil, eCode)
    }
    isp.reqScreenNum()
}

func (isp *ISPDevice) reqScreenNum() {
    isp.dispatch()
    return
}

func (isp *ISPDevice) alloc() (ip string, err error) {
    base := isp.base
    ur := isp.base.ur

    if M.PublicNetworkDetect() {
        lg.Info("do PublicNetworkDetect")
        if U.IsPublicNetwork(ur.ClientIP) {
            lg.Info("is PublicNetwork do default redirect")
            isp.destNode = S.AppSetting.DefaultRedirectHost
            ip = isp.destNode
            return
        }
    }

    found, err := InnerDockerLeastConn(ur)
    if err != nil {
        return
    }

    if !found {
        isp.destNode = S.AppSetting.DefaultRedirectHost
    } else {
        isp.destNode = base.ur.IPs[0]
    }

    ip = isp.destNode
    return
}

func (isp *ISPDevice) sftpAlloc() (ip string, err error) {
    ip = S.AppSetting.DefaultRedirectHost
    ur := isp.base.ur
    if len(ur.IPs) > 0 {
        ip = ur.IPs[0]
    }
    return ip, nil
}

//Dispatch Dispatch
func (isp *ISPDevice) Dispatch() {
    isp.dispatch()
}

func (isp *ISPDevice) dispatch() {
    base := isp.base
    ip, err := isp.alloc()
    if err != nil {
        lg.Error(err)
        return
    }
    lg.Info(ip)
    base.nodeIP = ip
    base.doDispatch()
}

func (isp *ISPDevice) makeResponse() {
    base := isp.base
    base.makeResponse(nil)
}

func (isp *ISPDevice) ZoneList() {
    isp.base.ZoneList()
}
