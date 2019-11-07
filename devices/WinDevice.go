/*
inner virtual machine
*/
package devices

type WinDevice struct {
    base *InnerVMDevice
}

func (dev *WinDevice) Init() (string, string) {
    return dev.base.Init()
}

//ZoneListDispatch ZoneListDispatch
func (dev *WinDevice) ZoneListDispatch() {
    dev.base.ZoneListDispatch()
}

//SftpDispatch request sftp
func (dev *WinDevice) SftpDispatch() {
    dev.base.SftpDispatch()
}

//LoginDispatch request screennum
func (dev *WinDevice) LoginDispatch() {
    dev.base.LoginDispatch()
}

//Dispatch Dispatch
func (dev *WinDevice) Dispatch() {
    dev.dispatch()
}

//dispatch default dispatch
func (dev *WinDevice) dispatch() {
    base := dev.base
    base.dispatch()
}

func (dev *WinDevice) ZoneList() {
    dev.base.ZoneList()
}
