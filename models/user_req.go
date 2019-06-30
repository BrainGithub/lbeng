package models

import (
	"encoding/json"
	"errors"
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
	DevIDs []string //Allocated dev id
	IPs    []string //Allocated ip
	Prots  []string //User defined protos
	Pools  []string //User defined pools
}

func _getProtocol() string {
	return ""
}

func _getUserIDByUserName(uname string) (string, error) {
	sql := fmt.Sprintf("select user_id from tab_basic_users where loginname = '%s'", uname)
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%s, sql：%s", err, sql)
	defer rows.Close()
	if err != nil {
		lg.Error(err.Error())
		return "", err
	}

	uid := ""
	for rows.Next() {
		rows.Scan(&uid)
	}

	return uid, nil
}

func _getUserNameByUserID(uid string) (string, error) {
	sql := fmt.Sprintf("select loginname from tab_basic_users where user_id = '%s'", uid)
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%s, sql：%s", err, sql)
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
	var tmp string
	var err error

	if user.Request == "" {
		msg := "request is nil"
		lg.Error(msg)
		return errors.New(msg)
	}

	if user.LoginName != "" {
		tmp, err = _getUserIDByUserName(user.LoginName)
		if err != nil {
			return err
		}
		user.ZoneID = tmp
	}

	if user.ZoneName != "" && user.ZoneID == "" {
		sql := fmt.Sprintf("select zone_id from tab_zones where zonename = '%s'", user.ZoneName)
		rows, err := db.Raw(sql).Rows()
		lg.FmtInfo("err:%s, sql：%s", err, sql)
		if err != nil {
			lg.Error(err.Error())
			return err
		}
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&tmp)
		}
		user.ZoneID = tmp
	}

	if user.ZoneName == "" && user.ZoneID != "" {
		sql := fmt.Sprintf("select zonename from tab_zones where zone_id = '%s'", user.ZoneID)
		rows, err := db.Raw(sql).Rows()
		lg.FmtInfo("err:%s, sql：%s", err, sql)
		if err != nil {
			lg.Error(err.Error())
			return err
		}
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&tmp)
		}
		user.ZoneName = tmp
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
func UserReqMarshalAndVerify(ctx []byte, user *UserReq) error {
	err := json.Unmarshal(ctx, user)
	if err != nil {
		lg.Error(err.Error())
		return err
	}

	//2. alignUserMsg
	err = user.alignUserMsg()
	if err != nil {
		lg.Error(err.Error())
		return err
	}

	return nil
}
