package func_device

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/gin-gonic/gin"
    "math"

    "lbeng/controller/port_pre_alloc"
    M "lbeng/models"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
    U "lbeng/pkg/utils"
)

var counter = U.GetCounter()

//getLeastConnUnderRaceCond
func getLeastConnUnderRaceCond(ur *M.UserReq) (found bool, err error) {
    //1. 数据库和增量缓存联合查询
    //2. 竞态查找最小连接
    //how to do:
    //InnerVM:     最大空闲 - 缓存已经使用的 = 可用空闲的，     求最大值并分配之
    //Docker&外置: 已登录的 + 缓存已经使用的 = 已经或正在使用的，求最小值并分配之, （数据库查询时取反，算法同上）
    stat := ur.Stat //[][]string{devid, ip, online, vmstate, vmIdleNum}
    var ip, devid string
    idleNum := math.MinInt32
    poolID := ur.Pools[0]

    defaultBucket, _ := ur.GetDefaultBucket()

    lg.FmtInfo("%s befor:%+v", ur.LoginName, defaultBucket)

    counter.Lock()
    lg.FmtInfo("enlock")

    for i, bucItem := range defaultBucket {
        iDevID := bucItem[0].(string)
        iIP := bucItem[1].(string)
        iNum := bucItem[2].(int)
        statNum := 0

        for _, item := range stat {
            hostip := item[1].(string)
            statNum = item[4].(int)
            if iIP == hostip {
                iNum += statNum
            }
        }

        iNum -= counter.C[poolID+iIP]
        defaultBucket[i][2] = iNum
        if idleNum < iNum {
            idleNum = iNum
            ip = iIP
            devid = iDevID
        }
    }

    if ip != "" && devid != "" {
        tmpips := []string{ip}
        tmpids := []string{devid}
        ur.IPs = tmpips
        ur.DevIDs = tmpids

        found = true

        counter.IncrUnsafe(ur.Pools[0] + ip)
        ur.Counter++
    }



    lg.FmtInfo("unlock")
    counter.Unlock()

    lg.FmtInfo("%s after:%+v", ur.LoginName, defaultBucket)
    lg.FmtInfo("%s found:%t, idleNum:%d", ur.LoginName, found, idleNum)
    return
}

//InnerVMLeastConn innerVMLeastConn
func InnerVMLeastConn(ur *M.UserReq) (found bool, err error) {
    return innerVMLeastConn(ur)
}

//查找单pool的内置虚机的最小连接节点，docker除外
//1. find the logged on mapping, if found return
//2. statistic and sort the unlogged list, from database
//3. get least conn, union join the unfinished connection
func innerVMLeastConn(ur *M.UserReq) (found bool, err error) {
    found, err = ur.InnerVMLogedOnMaping()
    if err != nil || found == true {
        //already logged on, return
        ur.LoggedOn = true
        return
    }

    vmstates := []int {2, 1, -1} //idle, creating, not abnormal
    for _, vmstate := range vmstates {

        found, err = ur.GetInnerVMLeastConnStat(vmstate)
        if err != nil {
            return
        }

        if !found && vmstate != -1 {
            continue
        }

        found, err = getLeastConnUnderRaceCond(ur)
        if err != nil {
            return
        }

        if found {
            break
        }
    }

    //search end, has to be found
    lg.FmtInfo("%s found: ", ur.LoginName, found)
    return
}

//InnerDockerLeastConn InnerDockerLeastConn
func InnerDockerLeastConn(ur *M.UserReq) (found bool, err error) {
    return innerDockerLeastConn(ur)
}

//innerDockerLeastConn, for docker
func innerDockerLeastConn(ur *M.UserReq) (found bool, err error) {
    found, err = ur.InnerDockerLogedOnMaping()
    if err != nil {
        return
    } else if found {
        //already logged on, return
        ur.LoggedOn = true
        return
    }

    found, err = ur.GetInnerDockerLeastConnStat()
    if err != nil {
        return
    }

    return getLeastConnUnderRaceCond(ur)
}

//OuterVMLeastConn OuterVMLeastConn
func OuterVMLeastConn(ur *M.UserReq, table string, item string) (found bool, err error) {
    return outerVMLeastConn(ur, table, item)
}

//查找单pool的内置虚机的最小连接节点，docker除外
//1. find the logged on mapping, if found return
//2. statistic and sort the unlogged list, from database
//3. get least conn, union join the unfinished connection
func outerVMLeastConn(ur *M.UserReq, table string, item string) (found bool, err error) {
    found, err = ur.OuterVMLogedOnMaping(table, item)
    if err != nil {
        return false, err
    } else if found == true {
        //already logged on, return
        ur.LoggedOn = true
        return
    }

    found, err = ur.GetOuterVMLeastConnStat(table)
    if err != nil {
        return
    }

    return getLeastConnUnderRaceCond(ur)
}

func externalVMLeastConn(ur *M.UserReq, tab string, item string) (bool, error) {
    return outerVMLeastConn(ur, tab, item)
}

func defaultLeastConn(ur *M.UserReq) (found bool, err error) {
    //lg.Info("after:%+v", *ur)
    found, err = ur.NormalLeastConnStat()
    if err != nil {
        //lg.Info("after:%+v", *ur)
        lg.Error(err.Error())
        return
    }

    return getLeastConnUnderRaceCond(ur)
}

func checkOnline(ur *M.UserReq) (bool, error) {
    return ur.IsHostOnline()
}

func weightRoundRobin(ur *M.UserReq) error {
    return nil
}


func hashMap(ur *M.UserReq) (bool, error) {
    return true, nil
}

//doLeastConn
//1. inner
//2. outer
func doLeastConn(ur *M.UserReq) (bool, error) {
    InnnerVM := []string{"DPD-WIN", "DPD-Linux", "DPD-WINSVR"}
    InnnerDocker := []string{"DPD-ISP"}
    //ExternalVM := []string{"DPD-TM-Win", "DPD-GRA-TM", "SecureRDP", "XDMCP", "VNCProxy"}

    var msg string
    prot := ur.Protocol
    if prot == "" {
        msg = "protocol is null"
        lg.Info(msg)
        return false, nil
    }

    //Inner Docker
    for _, v := range InnnerDocker {
        if v == prot {
            if M.PublicNetworkDetect() {
                lg.Info("do PublicNetworkDetect")
                if U.IsPublicNetwork(ur.ClientIP) {
                    lg.Info("is PublicNetwork do default redirect")
                    ur.IPs = []string{S.AppSetting.DefaultRedirectHost}
                    return true, nil
                }
            }
            return innerDockerLeastConn(ur)
        }
    }

    //Inner vm process
    for _, v := range InnnerVM {
        if v == prot {
            return innerVMLeastConn(ur)
        }
    }

    //External
    {
        if "DPD-TM-Win" == prot {
            return externalVMLeastConn(ur, "tab_rdp_runtime", "rdp_hostname")
        } else if "DPD-GRA-TM" == prot {
            return externalVMLeastConn(ur, "tab_vnc_runtime", "vnc_hostname")
        }
    }

    msg = fmt.Sprintf("wrong protocol:%s", prot)
    lg.Error(msg)
    return false, errors.New(msg)
}

func leastConnection(ur *M.UserReq) (found bool, err error) {
    found, err = doLeastConn(ur)
    lg.Info(found, err)
    return
}

func sharedVMAlloc(ur *M.UserReq) (bool, error) {
    if !ur.IsSharedVM {
        return false, nil
    }

    if cfged, _ := ur.IsSharedVMClientConfiged(); cfged == false {
        return false, errors.New("Please double check the Shared-VM client login configuration")
    }

    //is shared VM and configured
    return ur.GetSharedVMHost()
}

//_doAlloc
//1. shared vm
//2. hash map, logged on mapping
//3. least connection
func doAlloc(ur *M.UserReq) (bool, error) {
    var err error
    var found bool

    defaultBucket, _ := ur.GetDefaultBucket()
    if len(defaultBucket) == 0 {
        return false, nil
    }

    //shared vm
    found, err = sharedVMAlloc(ur)
    if err != nil || found == true {
        return found, err
    }

    //least connection
    return leastConnection(ur)
}

//allocate to do load balance, localhost as default
func allocate(ur *M.UserReq) error {
    var err error
    var found bool

    found, err = doAlloc(ur)
    if err != nil {
        return err
    }

    if !found {
        lg.Warn("none cluster, using localhost")
        defaultHost := S.AppSetting.DefaultRedirectHost
        ur.IPs = []string{defaultHost}
        return nil
    }

    return nil
}

//doDispatch
//1. check protocol, multi or zero protocols, return to user selection
//2. check return nodes, for multi nodes, return to user selection
//3. HEAD on, to the only 1
func doDispatch(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
    var nodeip string
    //allocated nodes ip num
    hostIPNum := len(ur.IPs)
    userDefinedProtocolsNum := len(ur.Prots)

    //1. user not defined
    if hostIPNum == 0 || (userDefinedProtocolsNum != 1 && ur.ZoneID != 0) {
        //user defined protocols
        if userDefinedProtocolsNum == 0 {
            BuildAndReturnMsg(c, ur, "Please double check the available resource pool assigned to this user!")
        } else if userDefinedProtocolsNum > 1 {
            _, odat := OneBuildMsgAutoLoginServer(ur)
            BuildAndReturnMsg2(c, odat)
        }
        return nil
    } else if hostIPNum > 1 {
        if ur.AutoLS.IP == "" {
            _, odat := OneBuildMsgAutoLoginServer(ur)
            BuildAndReturnMsg2(c, odat)
            return nil
        } else { //to stop others
            lg.Info("process outer multi-desktop")
            lg.FmtInfo("auto_login_server:%s", ur.AutoLS.IP)
            lg.FmtInfo("multi-desktops:%v", ur.IPs)

            if ur.Protocol == "DPD-TM-Win" || ur.Protocol == "DPD-GRA-TM" {
                nodeip = ur.IPs[0]
            } else {
                nodeip = ur.AutoLS.IP
            }

            StopRedundantDesktop(c, ur)
        }
    }

    if nodeip == "" {
        //3 HEAD ON
        nodeip = ur.IPs[0]
    }

    poolID := ""
    if len(ur.Pools) > 0 {
        poolID = ur.Pools[0]
    }

    loginID := ""
    freePortK := ""

    if ur.LoginName != "" && poolID != "" {
        loginID = ur.LoginName + poolID

        if counter.HasKey(loginID) {
            BuildAndReturnMsgUnfinishedLogging(c, ur)
            return nil
        }

        if ur.Protocol == "DPD-GRA-TM" {
            freePortK = fmt.Sprintf("%s_%d_%s_%s", ur.LoginName, ur.ZoneID, ur.Protocol, ur.AutoLS.IP)
        }
    }

    ha := "127.0.0.1"
    if isClu, stable, haIP := isClusterStable(); isClu {
        if stable {
            ha = haIP
        } else {
            return errors.New("Cluster is not stable, please try later")
        }
    }

    if ur.Request == "screennum" && ur.Protocol == "DPD-ISP" && ur.Restart == 1 {
        stopContainer(c, nodeip, ur)
        ur.Restart = 0
        return nil
    }

    //license check
    check_license_k := ""
    if ur.Request == "screennum" && poolID != "" && !ur.IsSharedVM {
        loggedOn, err := CheckLicense(ur);
        if err != nil {
            lg.Info(err.Error())
            return err
        }
        if loggedOn {
            check_license_k = ""
        } else {
            check_license_k = GetLicenseK(ur.LoginName, ur.ZoneID, ur.Protocol, poolID)
        }
    }

    //Counter increase
    counter.Incr("TotalDispat-" + nodeip)
    counter.Incr(loginID)

    url := fmt.Sprintf("%s%s:%s", S.AppSetting.PrefixUrl, ha, S.AppSetting.DefaultRedirectPort)
    err := vmRequest(c, url, bytesCtx, ur)

    //decrease
    FreeLicenseCounter(check_license_k)

    port_pre_alloc.FreePort(ur.AutoLS.IP, freePortK)
    counter.Decr(loginID)

    return err
}

func decrAllocCounter(ur *M.UserReq) {
    if len(ur.IPs[0]) <= 0 || len(ur.Pools) <= 0 {
        return
    }

    k := ur.Pools[0] + ur.IPs[0]
    for i:=0; i<ur.Counter; i++ {
        counter.Decr(k)
    }
}

//Dispatch handler
//1. resource allocate
//2. do dispatch request
func dispatch(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
    var err error

    ur.ClientIP = c.ClientIP()

    if ur.Request == "screennum" && ur.Protocol == "DPD-ISP" && ur.Restart == 1 {
        if err = allocate(ur); err != nil {
            lg.Info(err.Error())
            return err
        }

        doDispatch(c, bytesCtx, ur)

        ctxmap := make(map[string]interface{})
        err = json.Unmarshal(bytesCtx, &ctxmap)
        if err != nil {
            lg.Error(err.Error())
            return err
        }

        ctxmap["restart"] = 0
        bytesCtx, err = json.Marshal(ctxmap)
        if err != nil {
            lg.Error(err.Error())
            return err
        }
    }

    if err = allocate(ur); err != nil {
        lg.Info(err.Error())
        return err
    }

    err = doDispatch(c, bytesCtx, ur)

    decrAllocCounter(ur)

    counter.Log("end disp")

    return err
}

