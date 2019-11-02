/* outer windows vm
 */
package func_device

import (
    E "lbeng/pkg/e"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
)

type TMWinDevice struct {
    base     *DPDevice
    destNode string
}

func (dev *TMWinDevice) Init() (string, string) {
    return dev.base.Init()
}

//ZoneListDispatch ZoneListDispatch
func (dev *TMWinDevice) ZoneListDispatch() {
    dev.base.ZoneListDispatch()
}

//SftpDispatch request sftp
func (dev *TMWinDevice) SftpDispatch() {
    _, err := dev.alloc()
    if err != nil {
        dev.errProc(err, E.ERR_USR_SFTP)
        return
    }
    dev.base.SftpDispatch()
}

//LoginDispatch request screennum
func (dev *TMWinDevice) LoginDispatch() {
    dev.dispatch()
    return
}

func (dev *TMWinDevice) alloc() (ip string, err error) {
    base := dev.base
    ur := dev.base.ur

    found, err := outerVMLeastConn(ur, "tab_rdp_runtime", "rdp_hostname")
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
func (dev *TMWinDevice) Dispatch() {
    dev.dispatch()
}

func (dev *TMWinDevice) dispatch() {
    base := dev.base
    ip, err := dev.alloc()
    lg.Info(ip, err)

    if err != nil {
        lg.Error(err)
        return
    }
    base.Dispatch()
}

func (dev *TMWinDevice) errProc(err error, code int) {
    dev.base.errorProcess(err, code)
}

func (dev *TMWinDevice) ZoneList() {
    dev.base.ZoneList()
}
