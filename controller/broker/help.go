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
	rawdata, err := c.GetRawData()
	if err != nil {
		logging.Error(err.Error())
		return
	}

	data := make(map[string]interface{})
	json.Unmarshal(rawdata, &data)
	logging.Info(fmt.Sprintf("request:%+v", data))

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "broker server debug",
			"status":  http.StatusOK,
		})
}

//Test
func Test(c *gin.Context) {
	_test_loadbalance(c)
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

func _test_loadbalance(c *gin.Context) {
	req := map[string]interface{}{
		"request":      "zonelist",
		"buildVersion": "5.0.3-4928-20190220-145206",
		"passwd":       "123456",
		"user":         "local.zx1",
		"zonename":     "zone1",
		"hostname":     "192.168.10.33",
	}

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

	c.Data(http.StatusOK, "application/json;charset=UTF-8", respBytes)
}
