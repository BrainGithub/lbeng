package models

import (
    "fmt"

    lg "github.com/sirupsen/logrus"
)

type BrokerRequest struct {
    Request             string          `json:"request"`
    LoginName           string          `json:"user"`
    UserID              string          `json:"userid"`
    Passwd              string          `json:"passwd"`
    ZoneName            string          `json:"zonename"`
    ZoneID              int             `json:"zone_id"`
    Protocol            string          `json:"protocol"`
    Cap                 Capability      `json:"capability"`
    AutoLS              AutoLoginServer `json:"auto_login_server"`
    PoolID              string          `json:"pool_id"`
    ClientVer           string          `json:"buildVersion"`
    ClientIP            string          `json:"client_ip"`
    ClientMAC           string          `json:"client_mac"`
    ClientPort          int             `json:"client_port"`
    UserLoginRole       string          `json:"user_login_role"`
    StopIntranetDesktop int             `json:"stop_intranet_desktop"`
}

func (br *BrokerRequest) GetZoneList() ([]string, error) {
    var zoenlist []string

    sql := fmt.Sprintf("select a.zone_name from tab_zones a left join tab_user_zone_applications b "+
        "on a.zone_id = b.zone_id "+
        "where b.user_id = %s", br.UserID)
    rows, err := db.Raw(sql).Rows()
    lg.Infof("err:%v, sql:%s", err, sql)
    lg.Error("sdfsa")

    if err != nil {
        lg.Error(err.Error())
        return zoenlist, err
    }
    defer rows.Close()
    for rows.Next() {
        var zonename string
        if err := rows.Scan(&zonename); err != nil {
            lg.Error("db error:%s", err.Error())
            return zoenlist, err
        }
        zoenlist = append(zoenlist, zonename)
    }

    return zoenlist, err
}
