package loadbalance

import (
	M "lbeng/models"
	"strings"
)

func BuildMsgAutoLoginServer(user *M.UserReq) (odat map[string]interface{}) {

	var info []interface{}
	arr := strings.Split(user.LoginName, ".")
	domain := arr[0]
	for i, item := range user.Prots {
		ip := ""
		if len(user.IPs) == len(user.Prots) {
			ip = user.IPs[i]
		}
		m := map[string]string{"domain": domain, "server": ip, "protocol": item, "description": ""}
		info = append(info, m)
	}

	odat = map[string]interface{}{
		"request":                user.Request,
		"return":                 "ok",
		"auto_login_server_list": info,
		"user":                   user.LoginName,
		"zonename":               user.ZoneName,
		"comments":               "multi protocols, please make choice",
	}
	return
}

func BuildMsgDisplay(user *M.UserReq, msg string) (odat map[string]interface{}) {
	odat = map[string]interface{}{
		"request":  user.Request,
		"return":   "show-display",
		"display":  msg,
		"user":     user.LoginName,
		"zonename": user.ZoneName,
		"comments": "notice display",
	}
	return
}

func BuildMsgZoneNameNotMatch(user *M.UserReq) (odat map[string]interface{}) {
	odat = map[string]interface{}{
		"request":  user.Request,
		"return":   "zonename-not-match",
		"user":     user.LoginName,
		"zonename": user.ZoneName,
		"comments": "notice display",
	}
	return
}
