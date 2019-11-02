package broker

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    M "lbeng/models"
    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
    U "lbeng/pkg/utils"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "os/signal"
    "sync"
    "syscall"

    "github.com/gin-gonic/gin"
)

type LoginTest struct {
    User    string `json:"user"`
    ZoneID     int `json:"zone_id"`
    Protocol   string `json:"protocol"`
    Host       string `json:"host"`
    Start      int `json:"start"`
    Num        int `json:"num"`
}

var debugFile = ".broker.debug.json"

//Debug for on lined last connection
func Debug(c *gin.Context) {
    lg.Info("last request debug")
    temp, err := ioutil.ReadFile(debugFile)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"comments": err.Error()})
        return
    }
    encryBytes := U.ECBEncrypt(temp)
    reader := bytes.NewReader(encryBytes)
    url := fmt.Sprintf("%s%s:%s", S.AppSetting.PrefixUrl, "localhost", S.ServerSetting.HttpPort)
    lg.FmtInfo("last request debug:url:%s", url)

    resp, err := http.Post(url, "application/json; charset=UTF-8", reader)
    if err != nil {
        lg.Error(err.Error())
        c.JSON(http.StatusOK, gin.H{"comments": err.Error()})
        return
    }
    defer resp.Body.Close()

    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        lg.Error(err.Error())
        return
    }

    decryBytes := U.ECBDecrypt(respBytes)

    showBytes := make([]byte, 0, len(respBytes)+1+len([]byte("\n"))+len(decryBytes))
    showBytes = append(showBytes, "\n"...)
    showBytes = append(showBytes, decryBytes...)

    lg.FmtInfo("debugBytes:%s", showBytes)

    c.Data(http.StatusOK, "application/json;charset=UTF-8", showBytes)
    return
}

//TestDB test database
func TestDB(c *gin.Context) {
    var user M.UserReq
    user.LoginName = "local.super"

    tmp := map[string]interface{}{
        "hello": "world",
    }
    ctx, err := json.Marshal(tmp)
    if err != nil {
        lg.Error(err.Error())
    }
    M.UserReqMarshalAndVerify(ctx, &user)

    c.JSON(
        http.StatusOK,
        gin.H{
            "message":      "query db local.super",
            "status":       http.StatusOK,
            user.LoginName: fmt.Sprintf("%+v", user),
        })
}

func waitForSignal() {
    sigs := make(chan os.Signal)
    signal.Notify(sigs, os.Interrupt)
    signal.Notify(sigs, syscall.SIGTERM)
    <-sigs
}

//TestScreenum Test screen num
func TestMultiLoginVMLinux(c *gin.Context) {
    testMultiLogin(c, "DPD-Linux")
}

//TestScreenum Test screen num
func TestMultiLoginVMDocker(c *gin.Context) {
    testMultiLogin(c, "DPD-ISP")
}

//TestScreenum Test screen num
func TestMultiLoginVMWin(c *gin.Context) {
    testMultiLogin(c, "DPD-WIN")
}

func testMultiLogin(c *gin.Context, prot string) {
    var test LoginTest

    count := 60
    testHost := "127.0.0.1"
    rawuser := "local.zx"
    zoneID := 11
    start := 1

    if err := c.BindJSON(&test); err == nil {
        if test.Num > 0 {
            count = test.Num
        }

        if test.Start > 0 {
            start = test.Start
        }

        if test.Host != "" {
            testHost = test.Host
        }

        if test.User != "" {
            rawuser = test.User
        }

        if zoneID > 0 {
            zoneID = test.ZoneID
        }

        if test.Protocol != "" {
            prot = test.Protocol
        }
    }

    lg.FmtInfo("%+v", test)
    lg.Info(rawuser, zoneID, prot, count)


    var result []byte

    var wg sync.WaitGroup

    outch := make(chan []byte, 1024)
    defer close(outch)

    for i := start; i <= count; i++ {
        wg.Add(1)

        user := fmt.Sprintf("%s%03d", rawuser, i)

        go func(user string, zone int, prot string, host string) {
            defer wg.Done()

            reqScreenum := map[string]interface{}{
                "auto_login_server": fmt.Sprintf("{\"protocol\":\"%s\",\"ip\":\"\",\"domain\":\"local\"}", prot),
                "buildVersion":"5.1h1_simulator_test",
                "capability":"{\"ssh_tunnel\":1,\"rdp_proxy\":1}",
                "client_ip":"192.168.2.191",
                "geometry":"800x600",
                "hostip":host,
                "hostname":host,
                "network_bw":"LAN(100Mbps or higher)",
                "passwd":"1",
                "remote_app":"",
                "request":"screennum",
                "restart":0,
                "stop_intranet_desktop":0,
                "usb_redirect":1,
                "use_remote_app":0,
                "user":user,
                "user_login_desktop":"",
                "user_login_role":"",
                "zone_id":zone,
            }

            decryBytes := testLoadbalance(c, reqScreenum)
            outch <- decryBytes
            result = append(result, <-outch...)
            result = append(result, []byte("<br>")...)
            //lg.FmtInfo("-------result:%s", result)
        }(user, zoneID, prot, testHost)
    }

    wg.Wait()
    c.Data(http.StatusOK, "application/json;charset=UTF-8", result)
    return
}

func serveHTTP(c *gin.Context) {
    // http.Redirect(c.Writer, c.Request, "http://localhost:11980", http.StatusFound)
    guest, err := url.Parse("http://localhost:11980")
    if err != nil {
        lg.Error(err.Error())
    }
    proxy := httputil.NewSingleHostReverseProxy(guest)
    proxy.ServeHTTP(c.Writer, c.Request)
}

func testLoadbalance(c *gin.Context, req map[string]interface{}) (decryBytes []byte) {
    bytesData, err := json.Marshal(req)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    encryBytes := U.ECBEncrypt(bytesData)
    reader := bytes.NewReader(encryBytes)
    url := fmt.Sprintf("%s%s:%s/", S.AppSetting.PrefixUrl, S.AppSetting.DefaultRedirectHost, S.ServerSetting.HttpPort)
    request, err := http.NewRequest("POST", url, reader)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    request.Header.Set("Content-Type", "application/json;charset=UTF-8")
    client := http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    decryBytes = U.ECBDecrypt(respBytes)
    return
}

//RacingTestVM for concurrency
func RacingTestVM(c *gin.Context) {
    reqScreenum := map[string]interface{}{
        "request":               "screennum",
        "buildVersion":          "5.0.3-4968-20190404-172046",
        "auto_login_server":     "",
        "use_remote_app":        0,
        "restart":               0,
        "hostip":                "192.168.60.43",
        "user_login_role":       "",
        "zonename":              "zone1",
        "protocol":              "DPD-WIN",
        "user_login_desktop":    "",
        "hostname":              "192.168.60.43",
        "stop_intranet_desktop": 0,
        "passwd":                "1",
        "network_bw":            "LAN(100Mbps or higher)",
        "geometry":              "1904x1002",
        "client_ip":             "",
        "usb_redirect":          1,
        "user":                  "local.zx1",
        "remote_app":            "",
    }

    decryBytes := testLoadbalance(c, reqScreenum)

    // for {
    //     time.Sleep(time.Second)
    //     lg.Info("sleep 1 s")
    // }

    c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
    return
}

//RacingTestDocker for concurrency
func RacingTestDocker(c *gin.Context) {
    reqScreenum := map[string]interface{}{
        "request":           "screennum",
        "buildVersion":      "5.0.3-4968-20190404-172046",
        "auto_login_server": "",
        "use_remote_app":    0,
        "restart":           0,
        "hostip":            "192.168.10.184",
        "user_login_role":   "",
        "zonename":          "zone1",
        // "protocol":              "DPD-ISP",
        "user_login_desktop":    "",
        "hostname":              "192.168.10.184",
        "stop_intranet_desktop": 0,
        "passwd":                "1",
        "network_bw":            "LAN(100Mbps or higher)",
        "geometry":              "1904x1002",
        "client_ip":             "",
        "usb_redirect":          1,
        "user":                  "local.zx2",
        "remote_app":            "",
        "capability":            "{\"ssh_tunnel\":1,\"rdp_proxy\":1}",
    }

    decryBytes := testLoadbalance(c, reqScreenum)

    // for {
    //     time.Sleep(time.Second)
    //     lg.Info("sleep 1 s")
    // }

    c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
    return
}

//RacingTestShareVM for concurrency
func RacingTestShareVM(c *gin.Context) {
    reqScreenum := map[string]interface{}{
        "request":               "screennum",
        "buildVersion":          "5.0.3-4968-20190404-172046",
        "auto_login_server":     "",
        "use_remote_app":        0,
        "restart":               0,
        "hostip":                "192.168.10.184",
        "user_login_role":       "",
        "zonename":              "zone1-保护区",
        "protocol":              "DPD-WIN",
        "user_login_desktop":    "",
        "hostname":              "192.168.10.184",
        "stop_intranet_desktop": 0,
        "passwd":                "1",
        "network_bw":            "LAN(100Mbps or higher)",
        "geometry":              "1904x1002",
        "client_ip":             "",
        "usb_redirect":          1,
        "user":                  "local.zx6",
        "remote_app":            "",
    }

    decryBytes := testLoadbalance(c, reqScreenum)

    // for {
    //     time.Sleep(time.Second)
    //     lg.Info("sleep 1 s")
    // }

    c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
    return
}
