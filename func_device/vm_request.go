package func_device

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

    M "lbeng/models"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
    U "lbeng/pkg/utils"
    F "lbeng/pkg/file"
)

//vmRequest request portal
func vmRequest(c *gin.Context, url string, bytesCtx []byte, ur *M.UserReq) (err error) {

    http.DefaultClient.Timeout = S.AppSetting.DefaultRedirectTimeout

    err = doBrokerRequest(c, url, bytesCtx, ur)
    //err = doVMReq(c, url, bytesCtx, ur)
    return
}

//doBrokerRequest, dispatcher request
func doBrokerRequest(c *gin.Context, url string, bytesCtx []byte, ur *M.UserReq) error {
    data := make(map[string]interface{})
    json.Unmarshal(bytesCtx, &data)
    data["client_ip"] = c.ClientIP()
    data["protocol"] = ur.Protocol
    //lg.Info(fmt.Sprintf("json:%v", data))
    bytesData, err := json.Marshal(data)
    if err != nil {
        lg.Error(err.Error())
        return err
    }

    encryted := U.ECBEncrypt(bytesData)
    reader := bytes.NewReader(encryted)
    lg.FmtInfo("dispatch url:%s, data_len:%d", url, len(bytesData))
    resp, err := http.Post(url, "application/json; charset=UTF-8", reader)
    if err != nil {
        lg.Error(err.Error())
        return err
    }
    defer resp.Body.Close()

    lg.FmtInfo("return from:%s, datalen:%d", url, resp.ContentLength)

    contentType := resp.Header.Get("Content-Type")
    extraHeaders := map[string]string{}

    c.DataFromReader(http.StatusOK, resp.ContentLength, contentType, resp.Body, extraHeaders)
    return nil
}

//doVMReq requset vm directly
func doVMReq(c *gin.Context, baseUrl string, bytesCtx []byte, ur *M.UserReq) error {
    params := fmt.Sprintf("cmd=getVM&username=%s&zonename=%s&passwd=%s&vm_type=%s&image_id=%d&display_type=0&force=false&change_passwd=true",
        ur.LoginName,
        ur.ZoneName,
        ur.Passwd,
        ur.Protocol,
        ur.ImageIDs[0])

    //bytesData, err := json.Marshal(params)
    bytesData := []byte(params)
    reader := bytes.NewReader(bytesData)
    lg.FmtInfo("dispatch url:%s, data:%s", baseUrl, bytesData)
    resp, err := http.Post(baseUrl, "application/json; charset=UTF-8", reader)
    if err != nil {
        lg.Error(err.Error())
        return err
    }
    defer resp.Body.Close()

    contentType := resp.Header.Get("Content-Type")
    extraHeaders := map[string]string{}
    c.DataFromReader(http.StatusOK, resp.ContentLength, contentType, resp.Body, extraHeaders)
    return nil
}

//doBrokerPost doBrokerPost
func doBrokerPost(c *gin.Context, url string, data map[string]interface{}, ur *M.UserReq) error {
    data["client_ip"] = c.ClientIP()
    lg.Info(fmt.Sprintf("json:%v", data))
    bytesData, err := json.Marshal(data)
    if err != nil {
        lg.Error(err.Error())
        return err
    }

    encryted := U.ECBEncrypt(bytesData)
    reader := bytes.NewReader(encryted)
    lg.FmtInfo("dispatch url:%s, data:%s", url, bytesData)
    resp, err := http.Post(url, "application/json; charset=UTF-8", reader)
    if err != nil {
        lg.Error(err.Error())
        return err
    }
    defer resp.Body.Close()

    lg.FmtInfo("user:%s,%d,%s, return from:%s, datalen:%d", ur.LoginName, ur.ZoneID, ur.Protocol, url, resp.ContentLength)

    return nil
}

//doVMRequest requset vm directly
func doVMRequest(c *gin.Context, baseUrl string, bytesCtx []byte, ur *M.UserReq) error {
    data := make(map[string]interface{})
    json.Unmarshal(bytesCtx, &data)
    lg.Info(fmt.Sprintf("json:%v", data))
    bytesData, err := json.Marshal(data)
    if err != nil {
        lg.Error(err.Error())
        return err
    }

    // encryted := U.ECBEncrypt(bytesData)
    reader := bytes.NewReader(bytesData)
    lg.FmtInfo("dispatch url:%s, data:%s", baseUrl, bytesData)
    resp, err := http.Post(baseUrl, "application/json; charset=UTF-8", reader)
    if err != nil {
        lg.Error(err.Error())
        return err
    }
    defer resp.Body.Close()

    lg.FmtInfo("user:%s,%d,%s, return from:%s, datalen:%d", ur.LoginName, ur.ZoneID, ur.Protocol, baseUrl, resp.ContentLength)

    return nil
}


//is cluster stable
func isClusterStable() (isCluster bool, stable bool, ha string) {
    clu := M.GetClusterFromCache()
    if clu.IsCluster && clu.IsStable {
        ha = clu.HA
        isCluster = clu.IsCluster
        stable = clu.IsStable
        return
    }

    isCluster = !F.CheckNotExist("/license/ha_config")
    if !isCluster {
        return
    }

    _, _, ha = M.GetMasterIP()
    if ha == "" {
        var clu M.Cluster
        clu.IsCluster = isCluster
        clu.IsStable = stable
        clu.HA = ha
        M.SetClusterCache(clu)
        return
    }

    cmd := "cmd=cluster_state"
    reader := bytes.NewReader([]byte(cmd))
    url := "http://" + ha + ":" + strconv.Itoa(11913)
    lg.FmtInfo("http post url:%s, cmd:%s", url, cmd)
    http.DefaultClient.Timeout = S.AppSetting.DefaultRedirectTimeout
    resp, err := http.Post(url, "application/x-www-form-urlencoded", reader)
    if err != nil {
        lg.Error(err.Error())
        return
    }
    defer resp.Body.Close()

    respBytes, err := ioutil.ReadAll(resp.Body)
    lg.FmtInfo("%b, %s", err, respBytes)
    if err != nil {
        lg.Error(err.Error())
        return
    }

    ctnMap := make(map[string]interface{})
    err = json.Unmarshal(respBytes, &ctnMap)
    lg.FmtInfo("%s, %+v", err, ctnMap)

    if ctnMap["cluster"].(bool) && ctnMap["stable"].(bool) && ctnMap["result"].(bool) {
        isCluster = true
        stable = true
    }

    var clus M.Cluster
    clus.IsCluster = isCluster
    clus.IsStable = stable
    clus.HA = ha
    M.SetClusterCache(clus)

    return
}

//StopRedundantDesktop stop Redundant Desktop
func StopRedundantDesktop(c *gin.Context, ur *M.UserReq) {
    nodeip := ur.AutoLS.IP
    //stop others
    for _, ip := range ur.IPs {
        if ip == nodeip {
            continue
        }

        stopContainer(c, ip, ur)
    }
}

//StopRedundantDesktop stop Redundant Desktop
func stopContainer(c *gin.Context, ip string, ur *M.UserReq) {
    if ip == "" {
        return
    }

    cmd := "shutdownVM"
    if ur.Protocol == "DPD-ISP" {
        cmd = "shutdownDock"
    }

    lg.FmtInfo("to stop container:%s,%s,%d,%s", ip, ur.LoginName, ur.ZoneID, ur.Protocol)
    expCtx := make(map[string]interface{})
    expCtx["user"] = ur.LoginName
    expCtx["passwd"] = ur.Passwd
    expCtx["request"] = "stop_container"
    expCtx["cmd"] = cmd
    expCtx["zone_id"] = ur.ZoneID
    expCtx["protocol"] = ur.Protocol
    expCtx["image_id"] = ur.ImageIDs[0]
    expCtx["abnormal_flag"] = 1
    expCtx["remote_server"] = ur.AutoLS.IP
    expCtx["enable"] = 1

    ha := M.GetHAIP(ip)
    url := fmt.Sprintf("%s%s:%s", S.AppSetting.PrefixUrl, ha, S.AppSetting.DefaultRedirectPort)
    err := doBrokerPost(c, url, expCtx, ur)
    lg.Info(err)


    //this is totally wrong, but what can I do?
    //delete database
    if ur.Protocol == "DPD-TM-Win" {
        ur.ThisIsWrongJob("tab_rdp_runtime", ip, "rdp_hostname")
    } else if ur.Protocol == "DPD-GRA-TM" {
        ur.ThisIsWrongJob("tab_vnc_runtime", ip, "vnc_hostname")
    }
}
