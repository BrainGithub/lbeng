/* outer Linux vm
 */
package devices

import (
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
)

type TMLinDevice struct {
    base     *DPDevice
    destNode string
}

func (dev *TMLinDevice) Init() (string, string) {
    return dev.base.Init()
}

//ZoneListDispatch ZoneListDispatch
func (dev *TMLinDevice) ZoneListDispatch() {
    dev.base.ZoneListDispatch()
}

//SftpDispatch request sftp
func (dev *TMLinDevice) SftpDispatch() {
    dev.alloc()
    dev.base.SftpDispatch()
}

//LoginDispatch request screennum
func (dev *TMLinDevice) LoginDispatch() {
    dev.dispatch()
    return
}

func (dev *TMLinDevice) alloc() (ip string, err error) {
    base := dev.base
    ur := dev.base.ur

    found, err := OuterVMLeastConn(ur, "tab_vnc_runtime", "vnc_hostname")
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
func (dev *TMLinDevice) Dispatch() {
    dev.dispatch()
}

func (dev *TMLinDevice) dispatch() {
    base := dev.base
    ip, err := dev.alloc()
    lg.Info(ip, err)

    if err != nil {
        lg.Error(err)
        return
    }
    base.Dispatch()
}

func (dev *TMLinDevice) ZoneList() {
    dev.base.ZoneList()
}
