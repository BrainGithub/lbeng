package models

import (
	"encoding/json"
	"fmt"
	lg "lbeng/pkg/logging"
)

//Idat, ISPClient request data, virtual table
type UserReq struct {
	Request   string `json:"request"`
	LoginName string `json:"user"`
	UserID    string `json:"userid"`
	Passwd    string `json:"passwd"`
	ZoneName  string `json:"zonename"`
	ZoneID    string `json:"zoneid"`
	Protocol  string `json:"protocol"`
	ClientVer string `json:"buildVersion"`
	ClientIP  string `json:"clientip"`
	//-----for allocate
	DevIDs     []string        //Allocated dev id
	IPs        []string        //Allocated ip
	Prots      []string        //User defined protos
	Pools      []string        //User defined pools
	IsSharedVM bool            //shared VM
	ImageIDs   []int           //image id
	Stat       [][]interface{} //devid, ip, num list with ordered, for race cond
}

func _getProtocol() string {
	return ""
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
		}
		user.ZoneName = tmp
	} else {
		return nil
	}

	if user.Protocol == "" {
		user.Protocol = _getProtocol()
	}

	return nil
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

//UserVerify
//1. verifyPasswd openam
//2. alignUserMsg
//3. syncUser
//4. authorize
func UserReqMarshalAndVerify(ctx []byte, user *UserReq) (err error) {
	lg.FmtInfo("%s", ctx)
	err = json.Unmarshal(ctx, user)
	if err != nil {
		lg.Error(err.Error())
		return
	}

	//2. alignUserMsg
	err = user.alignUserMsg()
	if err != nil {
		lg.Error(err.Error())
		return
	}

	lg.FmtInfo("UserReq:%+v", *user)

	return
}
