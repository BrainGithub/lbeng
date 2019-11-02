package func_device

import (
    E "lbeng/pkg/e"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
)

type InnerVMDevice struct {
    base     *DPDevice
    destNode string
}

func (dev *InnerVMDevice) Init() (string, string) {
    return dev.base.Init()
}

//ZoneListDispatch ZoneListDispatch
func (dev *InnerVMDevice) ZoneListDispatch() {
    dev.base.ZoneListDispatch()
}

//SftpDispatch request sftp
func (dev *InnerVMDevice) SftpDispatch() {
    base := dev.base
    _, err := dev.alloc()
    if err != nil {
        base.errorProcess(err, E.ERR_USR_SFTP)
        return
    }

    base.sftpDispatch()
}

//LoginDispatch request screennum
func (dev *InnerVMDevice) LoginDispatch() {
    dev.dispatch()
}

func (dev *InnerVMDevice) alloc() (ip string, err error) {
    base := dev.base
    ur := dev.base.ur

    found, err := InnerVMLeastConn(ur)
    if err != nil {
        return
    }

    if !found {
        dev.destNode = S.AppSetting.DefaultRedirectHost
    } else {
        dev.destNode = base.ur.IPs[0]
    }

    ip = dev.destNode
    return
}

//Dispatch Dispatch
func (dev *InnerVMDevice) Dispatch() {
    dev.dispatch()
}

func (dev *InnerVMDevice) dispatch() {
    base := dev.base
    ip, err := dev.alloc()
    if err != nil {
        lg.Error(err)
        return
    }

    lg.Info(ip)
    base.nodeIP = ip
    base.Dispatch()
    return
}

func (dev *InnerVMDevice) errorProc(err error, code int) {
    dev.base.errorProcess(err, code)
}

func (dev *InnerVMDevice) ZoneList() {
    dev.base.ZoneList()
}
