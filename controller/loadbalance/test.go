package loadbalance

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "net/http"

    "github.com/gin-gonic/gin"

    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
    U "lbeng/pkg/utils"
)

var debugFile = ".lb.debug.json"

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
