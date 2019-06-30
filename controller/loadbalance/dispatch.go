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

func _inner_vm_least_conn(ur *M.UserReq) error {
	ur.GetInnerVMLeastConn()
	return nil
}

func _inner_docker_least_conn(ur *M.UserReq) error {
	return nil
}

func _extra_least_conn(ur *M.UserReq) error {

	return nil
}

func _normal_least_conn(ur *M.UserReq) error {

	return nil
}

func _checkOnline(ur *M.UserReq) error {
	return nil
}

func weight_round_robin(ur *M.UserReq) error {
	return nil
}

func hashMap(ur *M.UserReq) error {
	return ur.GetLogedOnMapping()
}

func leastConnection(ur *M.UserReq) error {
	InnnerVM := []string{"DPD-Win", "DPD-Linux", "DPD-WINSVR"}
	InnnerDocker := []string{"DPD-ISP"}
	ExternalVM := []string{"DPD-TM-Win", "DPD-GRA-TM", "SecureRDP", "XDMCP", "VNCProxy"}

	var msg string
	prot := ur.Protocol
	if prot == "" {
		msg = "protocol is null"
		lg.Error(msg)
		return errors.New(msg)
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
	return errors.New(msg)
}

//to Dispatch,
//do below 3 things:
//1. check protocol
//   multi or zero protocols, return to user selection
//2. check return nodes
//   for multi nodes, return to user selection
//3. only 1 node, HEAD on
func _doDisp(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
	//1.
	if ur.Protocol == "" {
		lg.Info("zero or multi protocols")
		//return to user
	}

	//2.
	if len(ur.IPs) != 1 {
		lg.Info("zero or multi VM nodes:%s", ur.IPs)
		//return
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
	lg.Info(fmt.Sprintf("client request:%+v", data))

	lg.Info(fmt.Sprintf("json:%v", data))
	bytesData, err := json.Marshal(data)
	if err != nil {
		lg.Error(err.Error())
		return err
	}
	encryted := U.ECBEncrypt(bytesData)
	reader := bytes.NewReader(encryted)
	url := fmt.Sprintf("http://%s:11900/", nodeip)
	resp, err := http.Post(url, "application/json; charset=UTF-8", reader)
	lg.Info(err, fmt.Sprintf("dispatch resp:%+v", resp))
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
func _doAlloc(ur *M.UserReq) error {
	var err error

	//maping success, return
	if err = hashMap(ur); err != nil {
		//least connection
		err = leastConnection(ur)
	}

	if err == nil && len(ur.IPs) == 0 {
		ur.IPs = append(ur.IPs, "localhost")
	}

	lg.Error(err, err.Error())

	return err
}

//preAlloc, before allocate node, we do below things
//1. get user defined protocols
//2. check protocols
//   a. if only 1 protocol, go on allocate
//   b. if multi protocols, go back to user for selection
func preAlloc(ur *M.UserReq) error {
	if err := ur.GetProtocols(); err != nil {
		return err
	}

	if len(ur.Prots) == 1 {
		ur.Protocol = ur.Prots[0]
	}

	return nil
}

func allocate(ur *M.UserReq) error {
	var err error
	if ur.LoginName != "" && ur.ZoneName != "" && ur.Protocol != "" {
		emsg := fmt.Sprintf("user name/zone or protocol missed:%+v", ur)
		lg.Error(emsg)
		return nil
	}

	if err = _doAlloc(ur); err == nil {
		err = _checkOnline(ur)
	}
	lg.Error(err.Error())
	return err
}

//Dispatch handler
func dispatch(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
	var err error
	if err = preAlloc(ur); err == nil {
		if err = allocate(ur); err == nil {
			err = _doDisp(c, bytesCtx, ur)
		}
	}

	return err
}
