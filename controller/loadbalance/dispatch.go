package loadbalance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	M "lbeng/models"
	lg "lbeng/pkg/logging"
	S "lbeng/pkg/setting"
	U "lbeng/pkg/utils"
	"math"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Counter struct {
	lock sync.Mutex // protects following fields
	C    map[string]int
}

var counter = &Counter{
	C: make(map[string]int),
}

func (cnt *Counter) incr(k string) {
	cnt.lock.Lock()
	cnt.C[k]++
	cnt.lock.Unlock()
}

func (cnt *Counter) incrTotalCounter() {
	cnt.lock.Lock()
	cnt.C["TotalConn"]++
	cnt.lock.Unlock()
}

func (cnt *Counter) incrTotalReq() {
	cnt.lock.Lock()
	cnt.C["TotalReq"]++
	cnt.lock.Unlock()
}

func (cnt *Counter) decr(k string) {
	cnt.lock.Lock()
	if cnt.C[k] > 0 {
		cnt.C[k]--
	}
	cnt.lock.Unlock()
}

func _union_least_conn_under_RaceCond(ur *M.UserReq) (found bool, err error) {
	lg.FmtInfo("befor:%+v", *ur)

	//1. 数据库和增量缓存联合查询
	//2. 竞态查找最小连接
	//how to do:
	//InnerVM:     最大空闲 - 缓存已经使用的 = 可用空闲的，     求最大值并分配之
	//Docker&外置: 已登录的 + 缓存已经使用的 = 已经或正在使用的，求最小值并分配之, （数据库查询时取反，算法同上）
	stat := ur.Stat //[][]string{devid, ip, online, vmstate, vmIdleNum}
	var ip, devid string
	idleNum := math.MinInt32
	poolID := ur.Pools[0]

	for i, item := range stat {
		hostip := item[1].(string)
		num := item[4].(int)
		num -= counter.C[poolID+hostip]
		ur.Stat[i][4] = num //for tracking
		if idleNum < num {
			idleNum = num
			ip = hostip
			devid = item[0].(string)
		}
	}

	if ip != "" && devid != "" {
		tmpips := []string{ip}
		tmpids := []string{devid}
		ur.IPs = tmpips
		ur.DevIDs = tmpids
		found = true
	}

	lg.FmtInfo("after:%+v", *ur)
	return
}

//查找单pool的内置虚机的最小连接节点，docker除外
//1. find the logged on mapping, if found return
//2. statistic and sort the unlogged list, from database
//3. query the unfinished connection
//4. repeat step1, to sovle the scenario: the database changed between the step2 finished and the step3 finished
func _inner_vm_least_conn(ur *M.UserReq) (found bool, err error) {
	// return ur.GetInnerVMLeastConn()
	found, err = ur.InnerVMLogedOnMaping()
	if err != nil || found == true {
		//already logged on, return
		return
	}

	found, err = ur.GetInnerVMLeastConnStat()
	if err != nil || found == false {
		return
	}

	return _union_least_conn_under_RaceCond(ur)
}

func _inner_docker_least_conn(ur *M.UserReq) (bool, error) {
	return false, nil
}

func _extra_least_conn(ur *M.UserReq) (bool, error) {

	return false, nil
}

func _normal_least_conn(ur *M.UserReq) (found bool, err error) {
	lg.Info("after:%+v", *ur)
	found, err = ur.NormalLeastConnStat()
	if err != nil || found == false {
		lg.Info("after:%+v", *ur)
		return
	}

	return _union_least_conn_under_RaceCond(ur)
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

	return ur.GetLogedOnMapping()
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
	return leastConnection(ur)
}

func allocate(ur *M.UserReq) error {
	var err error
	var found bool

	DefaultHOST := S.AppSetting.DefaultRedirectHost

	if ur.ZoneID == "" {
		lg.FmtInfo("loginname:%s, ZoneID:%s, may be request zonelist", ur.LoginName, ur.ZoneID)
		ur.IPs = append(ur.IPs, DefaultHOST)
	} else {
		if ur.LoginName != "" && ur.Protocol != "" && len(ur.Pools) > 0 {
			if found, err = _doAlloc(ur); err != nil {
				return err
			}
			if found != true {
				lg.Error("found nothing error, using localhost")
				ur.IPs = append(ur.IPs, DefaultHOST)
			}
		}
	}

	return err
}

//to Dispatch,
//do below 3 things:
//1. check protocol
//   multi or zero protocols, return to user selection
//2. check return nodes
//   for multi nodes, return to user selection
//3. only 1 node, HEAD on
func _doDisp(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
	//allocated nodes ip num
	hostIPNum := len(ur.IPs)
	//user not defined
	if hostIPNum == 0 || hostIPNum > 1 {
		//user defined protocols
		userDefinedProtocolsNum := len(ur.Prots)
		var odat map[string]interface{}
		if userDefinedProtocolsNum == 0 {
			odat = BuildMsgDisplay(ur, "No available resource pool to assign for this user!")
		} else {
			odat = BuildMsgAutoLoginServer(ur)
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
	data["client_ip"] = c.ClientIP()
	lg.Info(fmt.Sprintf("json:%v", data))
	bytesData, err := json.Marshal(data)
	if err != nil {
		lg.Error(err.Error())
		return err
	}

	//Counter increase
	poolID := ""
	if len(ur.Pools) > 0 {
		poolID = ur.Pools[0]
	}
	counter.incr(poolID + nodeip)
	counter.incr("TotalDispat-" + nodeip)
	lg.FmtInfo("before:%+v", *counter)

	encryted := U.ECBEncrypt(bytesData)
	reader := bytes.NewReader(encryted)
	url := fmt.Sprintf("http://%s:%s/", nodeip, S.AppSetting.DefaultRedirectPort)
	lg.FmtInfo("dispatch url:%s, data:%v", url, bytesData)
	resp, err := http.Post(url, "application/json; charset=UTF-8", reader)
	if err != nil {
		lg.Error(err.Error())
		return err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	extraHeaders := map[string]string{}

	c.DataFromReader(http.StatusOK, resp.ContentLength, contentType, resp.Body, extraHeaders)

	//decrease
	counter.decr(poolID + nodeip)
	lg.FmtInfo("after:%+v", *counter)

	return nil
}

//Dispatch handler
func dispatch(c *gin.Context, bytesCtx []byte, ur *M.UserReq) error {
	var err error

	ur.ClientIP = c.ClientIP()
	if err = allocate(ur); err != nil {
		lg.Info(err.Error())
		return err
	}

	return _doDisp(c, bytesCtx, ur)

}
