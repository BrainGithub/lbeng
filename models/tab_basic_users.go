package models

import (
	"fmt"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	lg "lbeng/pkg/logging"
)

//tab_vm_runtime
type TabBasicUsers struct {
	UserID    int    `gorm:"primary_key" json:"user_id"`
	Loginname string `json:"loginname"`
	Password  string `json:"password"`
}

func GetTabBasicUsers(user *UserReq) (bool, error) {
	lg.Info("in tab_basic_users")

	sql := fmt.Sprintf("select user_id from tab_basic_users where loginname = '%s'", user.LoginName)
	lg.Info(sql)
	rows, err := db.Raw(sql).Rows()

	defer rows.Close()
	for rows.Next() {
		var uid string
		rows.Scan(&uid)
	}

	// var basicUser TabBasicUsers

	// err := db.Select("*").Where(TabBasicUsers{Loginname: user.LoginName}).First(&basicUser).Error
	// lg.FmtInfo("%s, %s, %d", basicUser.Loginname, basicUser.Password, basicUser.UserID)
	return true, err
}
