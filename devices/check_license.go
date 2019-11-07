package devices

import (
    "net/http"
    "time"
    "sync"
    "fmt"
    "errors"
    "os/exec"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"

    M "lbeng/models"
    lg "lbeng/pkg/logging"
)

type SysLicense struct {
    valid               bool
    expired             bool


    isCluster           bool
    rdpProxy            bool
    strictSecurityMode  bool

    vds_num             int
    erdp_num            int
    standard_vm_num     int
    enterprise_vm_num   int
    enhanced_vm_num     int
    vm_3d_num           int
    accl_3d_num         int
}

//license pre-alloc mapping
type LicenseCounter struct {
    lock sync.Mutex // protects following fields

    C    map[string]int         //mem temp license, will be free after logged on
}

var licenseCounter = &LicenseCounter {
    C:              make(map[string]int),
}

var sysLic *SysLicense   //system license > pooled_license > logged on license

func CheckLicenseWebInterface(c *gin.Context) {
    tstart := time.Now()

    var lic M.LicReq
    c.BindJSON(&lic)

    lg.FmtInfo("%+v", lic)

    c.JSON(http.StatusOK, lic)
    elapsed := time.Since(tstart)
    lg.FmtInfo("request elapsed:%d ms", elapsed/time.Millisecond)
    return
}

func doLicenseCheck(lr *M.LicReq) (err error) {
    if sysLic == nil {
        InitSysLicense()
        err = errors.New("System License not inited")
    }

    if !sysLic.valid || sysLic.expired {
        err = errors.New("License is not valid or expired")
        return
    }

    sysLicNum := 0
    count := 0


    switch lr.Protocol {
    case "DPD-ISP":
        count, err = getVMSetUsed(lr, lr.GetISPLicensed)
        sysLicNum = sysLic.vds_num

    case "DPD-WIN", "DPD-WINSVR", "DPD-Linux":
        count, err = getVMSetUsed(lr, lr.GetInnerVMLicensed)
        sysLicNum = sysLic.standard_vm_num + sysLic.enterprise_vm_num + sysLic.enhanced_vm_num + sysLic.vm_3d_num

    case "DPD-TM-Win":
        count, err = getVMSetUsed(lr, lr.GetOuterWinLicensed)
        sysLicNum = sysLic.erdp_num

    case "DPD-GRA-TM":
        count, err = getVMSetUsed(lr, lr.GetOuterLinLicensed)
        sysLicNum = sysLic.accl_3d_num
    }

    if err != nil || count < 0 {
        err = errors.New("License check error, please try later.")
        return
    }


    poolLicNum := lr.GetPoolLicenseNum()
    lg.Info(count, poolLicNum, sysLicNum)
    if count < poolLicNum && poolLicNum <= sysLicNum {
        k := getTempLicenseK(lr.User, lr.Zone, lr.Protocol, lr.PoolID)
        incrLicenseCounterUnsafe(k)
    } else {
        lg.FmtInfo("%+v", lr)
        lg.FmtInfo("%+v", sysLic)

        err = errors.New("License reach limit\n please double confirm the license of current user pool")

        return
    }

    return
}

func getVMSetUsed(lr *M.LicReq, callback func() (map[string]bool, error)) (int, error) {
    vms, err := callback()
    lg.FmtInfo("%+v", vms)
    if err != nil {
        return -1, err
    }

    poolID := lr.PoolID

    //licenseCounter.lock.Lock()

    for k, _ := range licenseCounter.C {
        if poolID != "" && strings.Contains(k, poolID) {
            vms[k] = true
        }
    }

    //licenseCounter.lock.Unlock()

    return len(vms), nil
}

func GetLicenseK(user string, zone int, prot string, poolID string) (k string) {
    return getTempLicenseK(user, zone, prot, poolID)
}

func getTempLicenseK(user string, zone int, prot string, poolID string) (k string) {

    if user != "" && zone > 0 && prot != "" {
        k = fmt.Sprintf("%s_%d_%s_%s", user, zone, prot, poolID)
    }
    return
}

//InitSysLicense InitSysLicense
func InitSysLicense() {
    lg.Info("InitSysLicense start")

    go getSysLicense()

    lg.Info("InitSysLicense end")

    return
}

//getSysLicense, Infinite loop for system license
func getSysLicense() {
    var ret []byte
    var err error

    if sysLic == nil {
        sysLic = &SysLicense{}
    }

    for {
        cmd := "sbox_verify.sh"
        eCmd := exec.Command("/bin/bash", "-c", cmd)
        ret, err = eCmd.Output()
        //lg.FmtInfo("%s, %s, %s", ret, err, cmd)
        if err == nil {
            sysLic.valid = true
        }

        cmd = "sbox_verify.sh -q"
        eCmd = exec.Command("/bin/bash", "-c", cmd)
        ret, err = eCmd.Output()
        //lg.FmtInfo("%s, %s, %s", ret, err, cmd)
        if err == nil {
            sysLic.expired = false
        }

        licInfoCmd := exec.Command("/bin/bash", "-c", "sbox_show_license.sh")
        ret, err = licInfoCmd.Output()
        //lg.FmtInfo("%s, %s, %s", ret, err, licInfoCmd)

        if err == nil {
            arr := strings.Split(string(ret[:]), ",")
            for _, i := range arr {
                kv := strings.Split(i, "=")
                k := kv[0]
                v := kv[1]

                num, err := strconv.Atoi(v)
                if err != nil {
                    num = 0
                }

                if k == "Clustering" && v == "on" {
                    sysLic.isCluster = true
                } else if k == "RDP Proxy" && v == "on" {
                    sysLic.rdpProxy = true
                } else if k == "Strict Security" && v == "on" {
                    sysLic.strictSecurityMode = true
                } else if k == "Maximum Concurrent VDs" {
                    sysLic.vds_num = num
                } else if k == "Express RDP Number" {
                    sysLic.erdp_num = num
                } else if k == "Standard VMs Number" {
                    sysLic.standard_vm_num = num
                } else if k == "Enterprise VMs Number" {
                    sysLic.enterprise_vm_num = num
                } else if k == "Enhanced-VM Number" {
                    sysLic.enhanced_vm_num = num
                } else if k == "3D VMs Number" {
                    sysLic.vm_3d_num = num
                } else if k == "3D acceleration VMs Number" {
                    sysLic.accl_3d_num = num
                } else {
                    //lg.Warn("something LICENSE MISSED. %s, %s", k, v)
                }
            }
        }

        time.Sleep(10 * time.Second)
    }
}

func incrLicenseCounterUnsafe(k string) {

    //licenseCounter.lock.Lock()
    //lg.FmtInfo("before incr: %+v", *licenseCounter)
    //lg.Info("licenseCounter enlock")

    licenseCounter.C[k]++

    //lg.Info("licenseCounter unlock")
    //licenseCounter.lock.Unlock()

    return
}

func FreeLicenseCounter(k string) {
    if k == "" {
        return
    }



    licenseCounter.lock.Lock()
    lg.FmtInfo("before free: %+v", *licenseCounter)

    lg.Info("licenseCounter enlock")

    if licenseCounter.C[k] > 0 {
        licenseCounter.C[k]--
    }

    if licenseCounter.C[k] <= 0 {
        delete(licenseCounter.C, k)
    }

    lg.Info("licenseCounter unlock")

    lg.FmtInfo("after free: %+v", *licenseCounter)
    licenseCounter.lock.Unlock()

    return
}



func CheckLicense(ur *M.UserReq) (bool, error) {
    var lic M.LicReq

    lic.User       = ur.LoginName
    lic.Zone       = ur.ZoneID
    lic.Protocol   = ur.Protocol
    lic.GuestIP    = ur.AutoLS.IP
    lic.PoolID     = ur.Pools[0]

    lg.FmtInfo("%+v", lic)

    if lic.UserLoggedOn() {
        return true, nil
    }


    tstart := time.Now()
    licenseCounter.lock.Lock()
    lg.Info("licenseCounter enlock")

    err := doLicenseCheck(&lic)

    lg.Info("licenseCounter unlock")
    licenseCounter.lock.Unlock()
    elapsed := time.Since(tstart)
    lg.FmtInfo("check license check elapsed:%d ms", elapsed/time.Millisecond)

    return false, err
}
