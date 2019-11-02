package models

import (
    "fmt"
    lg "lbeng/pkg/logging"
)

//LicReq license request
type LicReq struct {
    User       string `json:"user"`
    UserID     int `json:"user_id"`
    Zone       int `json:"zone"`
    Protocol   string `json:"protocol"`
    PoolID     string `json:"pool_id"`
    GuestIP    string `json:"guest_ip"`
}

func (lic *LicReq) getNum(sqlStr string) (map[string]bool, error) {
    var vms = make(map[string]bool)
    rows, err := db.Raw(sqlStr).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sqlStr)

    if err != nil {
        lg.Error(err.Error())
        return vms, err
    }
    defer rows.Close()

    for rows.Next() {
        key := ""
        if err = rows.Scan(&key); err != nil {
            lg.Error("db error:%s", err.Error())
            return vms, err
        }

        vms[key] = true
    }

    return vms, nil
}

func (lic *LicReq) GetISPLicensed() (map[string]bool, error) {
    sql := fmt.Sprintf("select concat(login_name, '_', zone_id, '_DPD-ISP', '_', pool_id) "+
        "from tab_dock_runtime where pool_id = '%s'", lic.PoolID)
    return lic.getNum(sql)
}

func (lic *LicReq) GetInnerVMLicensed() (map[string]bool, error) {
    sql := fmt.Sprintf("select concat(login_name, '_', zone_id, '_', vm_type, '_', pool_id) from tab_vm_runtime "+
        "where pool_id = '%s' and login_name != ''", lic.PoolID)

    return lic.getNum(sql)
}

func (lic *LicReq) GetOuterWinLicensed() (map[string]bool, error) {
    sql := fmt.Sprintf("select concat(user, '_', zone_id, '_', protocol, '_', pool_id) "+
        "from tab_container_runtime where protocol='DPD-TM-Win' and pool_id = '%s'", lic.PoolID)
    return lic.getNum(sql)
}

func (lic *LicReq) GetOuterLinLicensed() (map[string]bool, error) {
    sql := fmt.Sprintf("select concat(user, '_', zone_id, '_', protocol, '_', pool_id) "+
        "from tab_container_runtime where protocol = 'DPD-GRA-TM' and pool_id = '%s'", lic.PoolID)
    return lic.getNum(sql)
}

func (lic *LicReq) GetPoolLicenseNum() (num int) {
    sql := fmt.Sprintf(
        "select b.user_number from tab_user_zone_applications a left join tab_auto_login_server b "+
        "on a.pool_id = b.pool_id "+
        "where a.loginname = '%s' "+
        "and a.zone_id = %d "+
        "and a.protocol = '%s'",
        lic.User,
        lic.Zone,
        lic.Protocol)

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        num = -1
        return
    }
    defer rows.Close()

    for rows.Next() {
        if err = rows.Scan(&num); err != nil {
            lg.Error("db error:%s", err.Error())
            return -1
        }
    }
    return
}

func (lic *LicReq) UserLoggedOn() bool {
    var sql string
    switch lic.Protocol {
    case "DPD-ISP":
        sql = fmt.Sprintf("select * from tab_dock_runtime "+
            "where login_name = '%s' and zone_id = %d",
            lic.User, lic.Zone)

    case "DPD-WIN", "DPD-WINSVR", "DPD-Linux":
        sql = fmt.Sprintf("select * from tab_vm_runtime "+
            "where login_name = '%s' and zone_id = %d and vm_type = '%s'",
            lic.User, lic.Zone, lic.Protocol)

    case "DPD-TM-Win":
        sql = fmt.Sprintf("select * from tab_rdp_runtime "+
            "where user = '%s' and zone_id = %d and protocol = '%s'",
            lic.User, lic.Zone, lic.Protocol)

    case "DPD-GRA-TM":
        sql = fmt.Sprintf("select * from tab_vnc_runtime "+
            "where user = '%s' and zone_id = %d and protocol = '%s'",
            lic.User, lic.Zone, lic.Protocol)

    default:
        return false
    }

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false
    }
    defer rows.Close()

    for rows.Next() {
        return true
    }

    return false
}
