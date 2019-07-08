package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"

	M "lbeng/models"
	"lbeng/pkg/logging"
	U "lbeng/pkg/utils"
)

//Help, broker server
func Help(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "broker server help manual",
			"status":  http.StatusOK,
		})

	logging.Info()
}

func Debug(c *gin.Context) {
	temp, err := ioutil.ReadFile(".debug.json")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"comments": err.Error()})
		return
	}
	encryBytes := U.ECBEncrypt(temp)
	reader := bytes.NewReader(encryBytes)
	url := "http://localhost:11980/"
	resp, err := http.Post(url, "application/json; charset=UTF-8", reader)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"comments": err.Error()})
		return
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	decryBytes := U.ECBDecrypt(respBytes)
	c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
	return
}

func TestDB(c *gin.Context) {
	var user M.UserReq
	user.LoginName = "local.super"

	tmp := map[string]interface{}{
		"hello": "world",
	}
	ctx, err := json.Marshal(tmp)
	if err != nil {
		logging.Error(err.Error())
	}
	M.UserReqMarshalAndVerify(ctx, &user)

	c.JSON(
		http.StatusOK,
		gin.H{
			"message":      "broker server db test",
			"status":       http.StatusOK,
			user.LoginName: fmt.Sprintf("%+v", user),
		})
}

//Test
func Test(c *gin.Context) {
	// reqZonelist := map[string]interface{}{
	// 	"request":      "zonelist",
	// 	"buildVersion": "5.0.3-4928-20190220-145206",
	// 	"passwd":       "1",
	// 	"user":         "local.test1",
	// }
	// _test_loadbalance(c, reqZonelist)

	reqScreenum := map[string]interface{}{
		"request":               "screennum",
		"buildVersion":          "5.0.3-4968-20190404-172046",
		"auto_login_server":     "",
		"use_remote_app":        0,
		"restart":               0,
		"hostip":                "192.168.10.184",
		"user_login_role":       "",
		"zonename":              "zone1",
		"protocol":              "DPD-ISP",
		"user_login_desktop":    "",
		"hostname":              "192.168.10.184",
		"stop_intranet_desktop": 0,
		"passwd":                "1",
		"network_bw":            "LAN(100Mbps or higher)",
		"geometry":              "1904x1002",
		"client_ip":             "",
		"usb_redirect":          1,
		"user":                  "local.k2",
		"remote_app":            "",
	}

	decryBytes := _test_loadbalance(c, reqScreenum)

	c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
	return
}

func _serveHTTP(c *gin.Context) {
	// http.Redirect(c.Writer, c.Request, "http://localhost:11980", http.StatusFound)
	guest, err := url.Parse("http://localhost:11980")
	if err != nil {
		logging.Error(err.Error())
	}
	proxy := httputil.NewSingleHostReverseProxy(guest)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func _test_loadbalance(c *gin.Context, req map[string]interface{}) (decryBytes []byte) {
	bytesData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	encryBytes := U.ECBEncrypt(bytesData)
	reader := bytes.NewReader(encryBytes)
	url := "http://localhost:11980/"
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

func RacingTestVM(c *gin.Context) {
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
		"user":                  "local.zx1",
		"remote_app":            "",
	}

	decryBytes := _test_loadbalance(c, reqScreenum)

	// for {
	// 	time.Sleep(time.Second)
	// 	lg.Info("sleep 1 s")
	// }

	c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
	return
}

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

	decryBytes := _test_loadbalance(c, reqScreenum)

	// for {
	// 	time.Sleep(time.Second)
	// 	lg.Info("sleep 1 s")
	// }

	c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
	return
}

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

	decryBytes := _test_loadbalance(c, reqScreenum)

	// for {
	// 	time.Sleep(time.Second)
	// 	lg.Info("sleep 1 s")
	// }

	c.Data(http.StatusOK, "application/json;charset=UTF-8", decryBytes)
	return
}
