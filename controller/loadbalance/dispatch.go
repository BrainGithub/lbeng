package loadbalance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	M "lbeng/models"
	lg "lbeng/pkg/logging"
	U "lbeng/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func _counter() {

}

func _inner_vm_least_conn(ur *M.UserReq) (bool, error) {
	return ur.GetInnerVMLeastConn()
}

func _inner_docker_least_conn(ur *M.UserReq) (bool, error) {
	return false, nil
}

func _extra_least_conn(ur *M.UserReq) (bool, error) {

	return false, nil
}

func _normal_least_conn(ur *M.UserReq) (bool, error) {

	return ur.NormalLeastConn()
}

func _checkOnline(ur *M.UserReq) (bool, error) {
	return ur.IsHostOnline()
}

func weight_round_robin(ur *M.UserReq) error {
	return nil
}

func hashMap(ur *M.UserReq) (bool, error) {
	if ur.LoginName == "" || ur.ZoneName == "" || ur.Protocol == "" {
		emsg := fmt.Sprintf("user name/zone or protocol missed:%+v", ur)
		lg.Error(emsg)
		return false, nil
	}

	found := false
	err := ur.GetLogedOnMapping()
	if err == nil && len(ur.IPs) > 0 {
		found = true
	}

	return found, err
}

func _doLeastConn(ur *M.UserReq) (bool, error) {
	InnnerVM := []string{"DPD-WIN", "DPD-Linux", "DPD-WINSVR"}
	InnnerDocker := []string{"DPD-ISP"}
	ExternalVM := []string{"DPD-TM-Win", "DPD-GRA-TM", "SecureRDP", "XDMCP", "VNCProxy"}

	var msg string
	prot := ur.Protocol
	if prot == "" {
		msg = "protocol is null"
		lg.Info(msg)
		return false, nil
	}

	//Inner vm process
	for _, v := range InnnerVM {
		if v == prot {
			return _inner_vm_least_conn(ur)
		}
	}

	//Inner Docker
	for _, v := range InnnerDocker {
		if v == prot {
			return _inner_docker_least_conn(ur)
		}
	}

	//External
	for _, v := range ExternalVM {
		if v == prot {
			return _extra_least_conn(ur)
		}
	}

	msg = fmt.Sprintf("wrong protocol:%s", prot)
	lg.Error(msg)
	return false, errors.New(msg)
}

func leastConnection(ur *M.UserReq) (bool, error) {
	found, err := _doLeastConn(ur)
	lg.Info(found, err)
	if err != nil {
		return false, err
	}

	if found == false {
		found, err = _normal_least_conn(ur)
	}

	return found, err
}

func sharedVMAlloc(ur *M.UserReq) (bool, error) {
	if err := ur.CheckSharedVM(); err != nil {
		return false, err
	}

	if !ur.IsSharedVM {
		return false, nil
	}

	//is shared VM
	return ur.GetSharedVMHost()
}

//to Dispatch,
//do below 3 things:
//1. check protocol
//   multi or zero protocols, return to user selection
//2. check return nodes
//   for multi nodes, return to user selection
//3. only 1 node, HEAD on
func _doDisp(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
	len := len(ur.IPs)
	//user not defined
	if len == 0 {
		//return
		odat := map[string]interface{}{
			"request":  ur.Request,
			"status":   "dispatch failed",
			"comments": "User not defined, please check configuration",
		}
		bytesData, err := json.Marshal(odat)
		if err != nil {
			lg.Error(err.Error())
			return err
		}
		encryted := U.ECBEncrypt(bytesData)
		c.Data(http.StatusOK, "application/json; charset=UTF-8", encryted)
		return nil
	} else if len > 1 {
		odat := map[string]interface{}{
			"request":  ur.Request,
			"status":   "dispatch failed",
			"comments": fmt.Sprintf("Multi nodes error:%s", ur.IPs),
		}
		bytesData, err := json.Marshal(odat)
		if err != nil {
			lg.Error(err.Error())
			return err
		}
		encryted := U.ECBEncrypt(bytesData)
		c.Data(http.StatusOK, "application/json; charset=UTF-8", encryted)
		return nil
	}

	//3 HEAD on
	nodeip := ur.IPs[0]
	if nodeip == "" {
		lg.Error("allocation error")
		return errors.New("allocation error")
	}
	data := make(map[string]interface{})
	json.Unmarshal(bytesCtx, &data)
	data["clientip"] = c.ClientIP()
	data["hostname"] = nodeip

	lg.Info(fmt.Sprintf("json:%v", data))
	bytesData, err := json.Marshal(data)
	if err != nil {
		lg.Error(err.Error())
		return err
	}
	encryted := U.ECBEncrypt(bytesData)
	reader := bytes.NewReader(encryted)
	url := fmt.Sprintf("http://%s:11900/", nodeip)
	lg.FmtInfo("dispatch url:%s", url)
	resp, err := http.Post(url, "application/json; charset=UTF-8", reader)
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

//node allocation
func _doAlloc(ur *M.UserReq) (bool, error) {
	var err error
	var found bool

	//shared vm
	found, err = sharedVMAlloc(ur)
	if err != nil || found == true {
		return found, err
	}

	//maping success, return
	if found, err = hashMap(ur); err != nil || found == true {
		return found, err
	}

	//least connection
	found, err = leastConnection(ur)
	if err != nil {
		return false, err
	}

	return found, err
}

func allocate(ur *M.UserReq) error {
	var err error
	var found bool

	Default_HOST := "192.168.10.184"

	if ur.ZoneID == "" {
		lg.FmtInfo("loginname:%s, ZoneID:%s, may be request zonelist", ur.LoginName, ur.ZoneID)
		ur.IPs = append(ur.IPs, Default_HOST)
	} else {
		if err := ur.GetProtocolsAndPools(); err == nil {

			len := len(ur.Prots)
			switch len {
			case 0:
			case 1:
				ur.Protocol = ur.Prots[0]
				if found, err = _doAlloc(ur); err != nil {
					return err
				}

				if found != true {
					lg.Error("found nothing error, using localhost")
					ur.IPs = append(ur.IPs, Default_HOST)
				}
			default:
			}
		}
	}

	return err
}

//Dispatch handler
func dispatch(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
	var err error

	if err = allocate(ur); err != nil {
		lg.Info(err.Error())
		return err
	}

	return _doDisp(c, bytesCtx, ur)

}
