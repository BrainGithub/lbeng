/*
inner virtual machine
*/
package devices

type LinuxDevice struct {
    base *InnerVMDevice
}

func (dev *LinuxDevice) Init() (string, string) {
    return dev.base.Init()
}

//ZoneListDispatch ZoneListDispatch
func (dev *LinuxDevice) ZoneListDispatch() {
    dev.base.ZoneListDispatch()
}

//SftpDispatch request sftp
func (dev *LinuxDevice) SftpDispatch() {
    dev.base.SftpDispatch()
}

//LoginDispatch request screennum
func (dev *LinuxDevice) LoginDispatch() {
    dev.base.LoginDispatch()
}

//Dispatch Dispatch
func (dev *LinuxDevice) Dispatch() {
    dev.dispatch()
}

//dispatch default dispatch
func (dev *LinuxDevice) dispatch() {
    base := dev.base
    base.dispatch()
}

func (dev *LinuxDevice) ZoneList() {
    dev.base.ZoneList()
}
