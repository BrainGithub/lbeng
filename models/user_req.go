package models

import (
	"encoding/json"
	"fmt"
	lg "lbeng/pkg/logging"
)

//Idat, ISPClient request data, virtual table
type UserReq struct {
	Request    string `json:"request"`
	LoginName  string `json:"user"`
	UserID     string `json:"userid"`
	Passwd     string `json:"passwd"`
	ZoneName   string `json:"zonename"`
	ZoneID     string `json:"zoneid"`
	Protocol   string `json:"protocol"`
	ClientVer  string `json:"buildVersion"`
	ClientIP   string `json:"clientip"`
	Capability string `json:"capability"`
	//-----for allocate
	DevIDs     []string        //Allocated dev id
	IPs        []string        //Allocated ip
	Prots      []string        //User defined protos
	Pools      []string        //User defined pools
	IsSharedVM bool            //shared VM
	ImageIDs   []int           //image id
	Stat       [][]interface{} //devid, ip, num list with ordered, for race cond
}

func (user *UserReq) getProtocolAndPools() error {
	if user.Protocol == "" {
		sql := fmt.Sprintf(
			"select protocol, pool_id from tab_user_zone_applications "+
				"where loginname = '%s' "+
				"and zone_id = %s ",
			user.LoginName,
			user.ZoneID)
		rows, err := db.Raw(sql).Rows()
		lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)

		if err != nil {
			lg.Error(err.Error())
			return err
		}

		defer rows.Close()
		for rows.Next() {
			var prot, poolId string
			if err := rows.Scan(&prot, &poolId); err != nil {
				lg.Error("db error:%s", err.Error())
				break
			}
			user.Prots = append(user.Prots, prot)
			user.Pools = append(user.IPs, poolId)
		}
		lg.Info(user.Prots, user.Pools)

		count := len(user.Prots)
		if count == 1 {
			user.Protocol = user.Prots[0]
		} else if count > 1 {

		}
	} else {
		sql := fmt.Sprintf(
			"select pool_id from tab_user_zone_applications "+
				"where loginname = '%s' "+
				"and zone_id = %s "+
				"and protocol = '%s'",
			user.LoginName,
			user.ZoneID,
			user.Protocol)
		rows, err := db.Raw(sql).Rows()
		lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)

		if err != nil {
			lg.Error(err.Error())
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var poolid string
			if err := rows.Scan(&poolid); err != nil {
				lg.Error("db error:%s", err.Error())
				break
			}
			user.Pools = append(user.Pools, poolid)
		}
	}
	return nil
}

func _getUserIDByUserName(uname string) (string, error) {
	sql := fmt.Sprintf("select user_id from tab_basic_users where loginname = '%s'", uname)
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)
	defer rows.Close()
	if err != nil {
		lg.Error(err.Error())
		return "", err
	}

	uid := ""
	for rows.Next() {
		rows.Scan(&uid)
	}
	lg.FmtInfo("uid:%s", uid)
	return uid, nil
}

func _getUserNameByUserID(uid string) (string, error) {
	sql := fmt.Sprintf("select loginname from tab_basic_users where user_id = '%s'", uid)
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)
	defer rows.Close()
	if err != nil {
		lg.Error(err.Error())
		return "", err
	}

	str := ""
	for rows.Next() {
		rows.Scan(&str)
	}

	return str, nil
}

//alignUserMsg align user info
//1. user name, id
//2. zone name, id
//3. protocol and VM resourse pool
func (user *UserReq) alignUserMsg() error {
	var err error

	if user.LoginName != "" && user.UserID == "" {
		var tmp string
		tmp, err = _getUserIDByUserName(user.LoginName)
		if err != nil {
			return err
		}
		user.UserID = tmp
	}

	found := false
	if user.ZoneName != "" && user.ZoneID == "" {
		sql := fmt.Sprintf("select zone_id from tab_zones where zone_name = '%s'", user.ZoneName)
		rows, err := db.Raw(sql).Rows()
		lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)
		if err != nil {
			lg.Error(err.Error())
			return err
		}
		defer rows.Close()
		var tmp string
		for rows.Next() {
			rows.Scan(&tmp)
			found = true
		}
		user.ZoneID = tmp
	} else if user.ZoneName == "" && user.ZoneID != "" {
		sql := fmt.Sprintf("select zone_name from tab_zones where zone_id = %s", user.ZoneID)
		rows, err := db.Raw(sql).Rows()
		lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)
		if err != nil {
			lg.Error(err.Error())
			return err
		}
		defer rows.Close()
		var tmp string
		for rows.Next() {
			rows.Scan(&tmp)
			found = true
		}
		user.ZoneName = tmp
	}

	found = found

	if user.ZoneID == "" || user.ZoneName == "" {
		return nil
	}

	//get protocols and Pools
	err = user.getProtocolAndPools()

	return err
}

func (user *UserReq) verifyPasswd() error {
	return nil
}

func (user *UserReq) sync() error {
	return nil
}

func (user *UserReq) authorize() error {
	return nil
}

//UserReqMarshalAndVerify marshal user request and do simple verify
//序列化请求数据
func UserReqMarshalAndVerify(ctx []byte, user *UserReq) (err error) {
	lg.FmtInfo("%s", ctx)
	err = json.Unmarshal(ctx, user)
	if err != nil {
		lg.Error(err.Error())
		return
	}

	err = user.alignUserMsg()
	if err != nil {
		lg.Error(err.Error())
		return
	}

	lg.FmtInfo("UserReq:%+v", *user)

	return
}
