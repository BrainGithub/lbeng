package models

import (
    "fmt"
    lg "lbeng/pkg/logging"
)

type PortPreAlloc struct {
    User       string `json:"user"`
    Zone       int `json:"zone"`
    Protocol   string `json:"protocol"`
    GuestIP    string `json:"guest_ip"`
    Port       int `json:"port"`
    Start      int `json:"start"`
    Step       int `json:"step"`
    Cap        int `json:"cap"`

    Ports      map[int]bool
}

func (pa *PortPreAlloc) GetUsedPort() error {
    sql := fmt.Sprintf(
        "select vnc_port from tab_vnc_runtime "+
        "where vnc_hostname = '%s' and "+
        "protocol = '%s'", pa.GuestIP, pa.Protocol)
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return err
    }
    defer rows.Close()

    ports := make(map[int]bool)
    for rows.Next() {
        var port int
        if err = rows.Scan(&port); err != nil {
            lg.Error("db error:%s", err.Error())
            pa.Ports = ports
            return err
        }
        ports[port] = true
    }

    pa.Ports = ports
    lg.FmtInfo("%+v", pa)
    return nil
}