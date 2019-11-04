package models

import (
    "encoding/json"
    "fmt"
    "strings"

    lg "lbeng/pkg/logging"
    S "lbeng/pkg/setting"
)

type AutoLoginServer struct {
    Prot   string `json:"protocol"`
    IP     string `json:"ip"`
    Domain string `json:"domain"`
}

type Capability struct {
    SSHTunnel int `json:"ssh_tunnel"`
    RDPProxy  int `json:"rdp_proxy"`
}

//Idat, ISPClient request data, virtual table
type UserReq struct {
    Request         string `json:"request"`
    LoginName       string `json:"user"`
    UserID          string `json:"userid"`
    Passwd          string `json:"passwd"`
    ZoneName        string `json:"zonename"`
    ZoneID          int    `json:"zone_id"`
    Restart         int    `json:"restart"`
    Protocol        string `json:"protocol"`
    ClientVer       string `json:"buildVersion"`
    ClientIP        string `json:"clientip"`
    Capability      string `json:"capability"`
    AutoLoginServer string `json:"auto_login_server"`
    RemoteServer    string `json:"remote_server"`

    //-----for allocate
    DevIDs     []string        //Allocated dev id
    IPs        []string        //Allocated ip
    Prots      []string        //User defined protos
    Pools      []string        //User defined pools
    IsSharedVM bool            //shared VM
    ImageIDs   []int           //image id
    Stat       [][]interface{} //devid, ip, num list with ordered, for race cond
    AutoLS     AutoLoginServer //AutoLoginServer

    //-----name insensitivate
    caseIgnore bool
    LoggedOn   bool
    Counter    int             //incr counter
}

func (user *UserReq) GetProtocolAndPools() error {
    return user.getProtocolAndPools()
}

func (user *UserReq) getProtocolAndPools() error {
    sql := fmt.Sprintf(
        "select a.protocol, a.pool_id, a.ip from tab_user_zone_applications a "+
            "where a.loginname = '%s' "+
            "and a.zone_id = %d "+
            "and if('%s' = '', 1, '%s' = a.protocol)",
        user.LoginName,
        user.ZoneID,
        user.Protocol, user.Protocol)

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return err
    }

    var prots, pools, ips []string

    defer rows.Close()
    for rows.Next() {
        var prot, poolId, ip string
        if err := rows.Scan(&prot, &poolId, &ip); err != nil {
            lg.Error("db error:%s", err.Error())
            break
        }

        var isA bool
        if prot != "" {
            isA = false
            for _, item := range prots {
                if prot == item {
                    isA = true
                }
            }
            if !isA {
                prots = append(prots, prot)
            }
        }

        if poolId != "" {
            isA = false
            for _, item := range pools {
                if poolId == item {
                    isA = true
                }
            }
            if !isA {
                pools = append(pools, poolId)
            }
        }

        if ip != "" {
            isA = false
            for _, item := range ips {
                if ip == item {
                    isA = true
                }
            }
            if !isA {
                ips = append(ips, ip)
            }
        }
    }
    lg.Info(prots, pools, ips)
    user.Prots = prots
    user.Pools = pools
    user.IPs = ips

    if len(user.Prots) == 1 {
        user.Protocol = user.Prots[0]
    }

    return nil
}

func getUserIDByUserName(user *UserReq) (string, string, error) {
    uname := user.LoginName
    sql := ""
    rawName := ""
    uid := ""

    user.caseIgnore = userNameCaseIgnore(uname)

    if user.caseIgnore {
        sql = fmt.Sprintf("select loginname, user_id from tab_basic_users where upper(loginname) = upper('%s')", uname)
    } else {
        sql = fmt.Sprintf("select loginname, user_id from tab_basic_users where loginname = '%s'", uname)
    }

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)
    defer rows.Close()
    if err != nil {
        lg.Error(err.Error())
        return "", "", err
    }

    for rows.Next() {
        rows.Scan(&rawName, &uid)
    }
    return rawName, uid, nil
}

func getUserNameByUserID(uid string) (string, error) {
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

func userNameCaseIgnore(name string) bool {
    //AD, LDAP is case insensative
    arr := strings.Split(name, ".")
    if len(arr) > 1 {
        ream := arr[0]
        switch strings.ToUpper(ream) {
        case "LDAP", "AD", "LOCAL":
            return true
        }
    }

    return false
}

//alignUserMsg align user info
//1. user name, id
//2. zone name, id
//3. protocol and VM resourse pool
func (user *UserReq) alignUserMsg() error {
    var err error
    var rawName, uid string

    rawName, uid, err = getUserIDByUserName(user)
    if err != nil {
        lg.Error(rawName, uid, err.Error())
        return err
    }
    user.UserID = uid
    user.LoginName = rawName
    lg.Info("%+v", *user)

    found := false
    if user.ZoneName != "" && user.ZoneID == 0 {
        sql := fmt.Sprintf("select zone_id from tab_zones where zone_name = '%s'", user.ZoneName)
        rows, err := db.Raw(sql).Rows()
        lg.FmtInfo("err:%v, sql:%s", err, sql)
        if err != nil {
            lg.Error(err.Error())
            return err
        }
        defer rows.Close()
        var tmp int
        for rows.Next() {
            rows.Scan(&tmp)
            found = true
        }
        user.ZoneID = tmp
    } else if user.ZoneName == "" && user.ZoneID != 0 {
        sql := fmt.Sprintf("select zone_name from tab_zones where zone_id = %d", user.ZoneID)
        rows, err := db.Raw(sql).Rows()
        lg.FmtInfo("err:%v, sql:%s", err, sql)
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

    if user.ZoneID == 0 || user.ZoneName == "" {
        return nil
    }

    //get protocols and Pools
    if err = user.getProtocolAndPools(); err != nil {
        return err
    }

    if err = user.CheckSharedVM(); err != nil {
        return err
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

//UserReqMarshalAndVerify marshal user request and do simple verify
//序列化请求数据
func UserReqMarshalAndVerify(ctx []byte, user *UserReq) (err error) {
    //lg.FmtInfo("%s", ctx)
    err = json.Unmarshal(ctx, user)
    if err != nil {
        lg.Error(err.Error())

        ctxmap := make(map[string]interface{})
        err = json.Unmarshal(ctx, &ctxmap)
        lg.FmtInfo("%+v, %s", ctxmap, err)
        autoLS, ok := ctxmap["auto_login_server"]
        lg.FmtInfo("%v, %t", autoLS, ok)
        if ok {
            switch autoLS.(type) {
            case string:
                var auto AutoLoginServer
                if user.AutoLoginServer != "" {
                    if err := json.Unmarshal([]byte(user.AutoLoginServer), &auto); err == nil {
                        user.Protocol = auto.Prot
                        user.AutoLS = auto
                    }
                }
                lg.Info("get string protocol:", auto.Prot)
            case map[string]interface{}:
                val := autoLS.(map[string]interface{})["protocol"]
                user.Protocol = val.(string)
                user.AutoLS.Prot = val.(string)

                val = autoLS.(map[string]interface{})["ip"]
                user.AutoLS.IP = val.(string)

                val = autoLS.(map[string]interface{})["domain"]
                user.AutoLS.Domain = val.(string)

                lg.FmtInfo("%+v", user.AutoLS)
            }
        }
    } else {
        var auto AutoLoginServer
        if user.AutoLoginServer != "" {
            if err := json.Unmarshal([]byte(user.AutoLoginServer), &auto); err == nil {
                user.Protocol = auto.Prot
                user.AutoLS = auto
            }
        }
        lg.Info("get string protocol:", auto.Prot)
    }

    //here is wrong, just to compati client error
    if user.RemoteServer != "" && user.AutoLS.IP == "" {
        user.AutoLS.IP = user.RemoteServer
    }

    err = user.alignUserMsg()
    if err != nil {
        lg.Error(err.Error())
        return
    }

    lg.FmtInfo("UserReq:%+v", *user)

    return
}

//GetDescription, GetDescription
func (ur *UserReq) GetDescription(user string, zone int, p string, ip string) (desc string) {
    sql := fmt.Sprintf("select b.description from tab_user_zone_applications a left join tab_auto_login_server b "+
            "on a.pool_id = b.pool_id "+
            "where a.loginname = '%s' and a.zone_id = %d and '%s' = a.protocol and b.ip = '%s'", user, zone, p, ip)

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)
    if err != nil {
        lg.Error(err.Error())
        return
    }
    defer rows.Close()
    for rows.Next() {
        rows.Scan(&desc)
    }

    return
}

//GetAutoLoginSvrIP, GetAutoLoginSvrIP
func (ur *UserReq) GetAutoLoginSvrIP(uname string, zid int, p string) (ips []string) {
    sql := ""
    switch p {
    case "DPD-ISP":
        sql = fmt.Sprintf(
                "select distinct b.ip "+
                " from tab_dock_runtime a join tab_cluster b"+
                " on a.dev_id = b.dev_id " +
                " where a.login_name = '%s'"+
                " and a.zone_id = %d",
                uname, zid)
    case "DPD-WIN", "DPD-WINSVR", "DPD-Linux":
        sql = fmt.Sprintf(
                "select distinct b.ip from tab_vm_runtime a join tab_cluster b "+
                " on a.dev_id = b.dev_id " +
                " where a.login_name = '%s'" +
                " and a.zone_id = %d" +
                " and a.vm_type = '%s'",
                uname, zid, p)
    case "DPD-TM-Win":
        fallthrough
    case "DPD-GRA-TM":
        sql = fmt.Sprintf(
                "select distinct ip "+
                "from tab_user_zone_applications "+
                "where loginname = '%s' "+
                "and zone_id = %d " +
                "and protocol = '%s' ",
                uname, zid, p)
    }

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)
    if err != nil {
        lg.Error(err.Error())
        return
    }
    defer rows.Close()
    for rows.Next() {
        var ip string
        rows.Scan(&ip)
        ips = append(ips, ip)
    }

    return
}

//InnerVMLogedOnMaping InnerVMLogedOnMaping
func (user *UserReq) InnerVMLogedOnMaping() (bool, error) {
    var err error
    found := false
    sql := fmt.Sprintf(
            "select distinct a.dev_id, a.ip "+
                "from tab_cluster a left join tab_vm_runtime b "+
                "on a.dev_id = b.dev_id "+
                "where a.type != 'backup' "+
                "and a.online = 1 "+
                "and b.login_name = '%s' "+
                "and b.zone_id = %d "+
                "and b.pool_id = '%s' "+
                "and b.state != 11",
            user.LoginName,
            user.ZoneID,
            user.Pools[0])

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()
    for rows.Next() {
        var devid, ip string
        if err := rows.Scan(&devid, &ip); err != nil {
            lg.Error("db error:%s", err.Error())
            return false, err
        }
        user.DevIDs = append(user.DevIDs, devid)
        user.IPs = append(user.IPs, ip)
        found = true
    }

    lg.FmtInfo("found:%s, %s", user.DevIDs, user.IPs)
    return found, nil
}

//InnerDockerLogedOnMaping InnerDockerLogedOnMaping
func (user *UserReq) InnerDockerLogedOnMaping() (bool, error) {
    var err error
    found := false
    sql := fmt.Sprintf(
            "select distinct a.dev_id, a.ip "+
                "from tab_cluster a left join tab_dock_runtime b "+
                "on a.dev_id = b.dev_id "+
                "where a.type != 'backup' "+
                "and a.online = 1 "+
                "and b.login_name = '%s' "+
                "and b.zone_id = %d "+
                "and b.pool_id = '%s' ",
            user.LoginName,
            user.ZoneID,
            user.Pools[0])

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()

    var devIDs []string
    var IPs []string
    for rows.Next() {
        var devid, ip string
        if err := rows.Scan(&devid, &ip); err != nil {
            lg.Error("db error:%s", err.Error())
            return false, err
        }
        devIDs = append(devIDs, devid)
        IPs = append(IPs, ip)
        found = true
    }
    user.DevIDs = devIDs
    user.IPs = IPs

    lg.FmtInfo("found:%s, %s", user.DevIDs, user.IPs)
    return found, nil
}

//GetInnerDockerLeastConnStat GetInnerDockerLeastConnStat
func (user *UserReq) GetInnerDockerLeastConnStat() (bool, error) {
    sql := fmt.Sprintf(
        "select a.dev_id, a.ip, a.online, if(b.state is null, -1, b.state), if(b.dev_id is null, 0, count(*)) as num "+
            "from tab_cluster a left join tab_dock_runtime b "+
            "on a.dev_id = b.dev_id "+
            "where a.type != 'backup' "+
            "and a.online = 1 "+
            "group by a.dev_id order by num")

    return user._getLeaseConnStat(sql, true)
}

//OuterVMLogedOnMaping OuterVMLogedOnMaping
func (user *UserReq) OuterVMLogedOnMaping(table string, hostnameItem string) (bool, error) {
    var err error
    found := false
    sql := fmt.Sprintf(
            "select distinct a.dev_id, a.ip "+
                "from tab_cluster a left join %s b "+
                "on a.dev_id = b.dev_id "+
                "where a.type != 'backup' "+
                "and a.online = 1 "+
                "and b.user = '%s' "+
                "and b.zone_id = %d "+
                "and b.pool_id = '%s' "+
                "and b.%s = '%s'",
            table,
            user.LoginName,
            user.ZoneID,
            user.Pools[0],
            hostnameItem,
            user.AutoLS.IP)

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()

    var devids, ips []string
    for rows.Next() {
        var devid, ip string
        if err := rows.Scan(&devid, &ip); err != nil {
            lg.Error("db error:%s", err.Error())
            return false, err
        }
        devids = append(devids, devid)
        ips = append(ips, ip)
        found = true
    }

    if found {
        user.DevIDs = devids
        user.IPs = ips
    }

    lg.FmtInfo("found:%s, %s", user.DevIDs, user.IPs)
    return found, nil
}

func (user *UserReq) GetDefaultBucket() ([][]interface{}, error) {
    var defaultBucket [][]interface{} //defaultBucket
    sql := fmt.Sprintf("select dev_id, ip, 0 as num from tab_cluster where type != 'backup' and online = 1 order by weight DESC")
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)
    if err != nil {
        lg.Error(err.Error())
        return defaultBucket, err
    }
    defer rows.Close()

    for rows.Next() {
        var devid, ip string
        var num int

        if err := rows.Scan(&devid, &ip, &num); err != nil {
            lg.Error("db error:%s", err.Error())
            return defaultBucket, err
        }

        var tmpVar []interface{} //vm variables slice
        tmpVar = append(tmpVar, devid, ip, num)
        defaultBucket = append(defaultBucket, tmpVar)
    }

    return defaultBucket, err
}

//
func (user *UserReq) _getLeaseConnStat(sql string, inverse bool) (bool, error) {
    var found bool
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()

    var stat [][]interface{} //statistic
    for rows.Next() {
        var devid, ip, online, state string
        var num int
        if err := rows.Scan(&devid, &ip, &online, &state, &num); err != nil {
            lg.Error("db error:%s", err.Error())
            return false, err
        }
        user.DevIDs = append(user.DevIDs, devid)
        user.IPs = append(user.IPs, ip)
        found = true

        //statistic for union query
        if inverse {
            num = -num
        }
        var tmpVar []interface{} //vm variables slice
        tmpVar = append(tmpVar, devid, ip, online, state, num)
        stat = append(stat, tmpVar)
    }

    if found == true {
        lg.FmtInfo("found:%t, [devid ip online state num]:%+v", found, stat)
        user.Stat = stat
    }

    return found, nil
}

//GetInnerVMLeastConnStat add statistic of idle onlined VM
func (user *UserReq) GetInnerVMLeastConnStat(vmstate int) (found bool, err error) {
    //abnormal vms
    if vmstate == -1 {

        sql := fmt.Sprintf(
        "select a.dev_id, a.ip, a.online, b.state, if(b.dev_id is null, 0, count(*)) as num "+
            "from tab_cluster a left join tab_vm_runtime b "+
            "on a.dev_id = b.dev_id "+
            "where a.type != 'backup' "+
            "and a.online = 1 "+
            "and b.zone_id = %d "+
            "and b.pool_id = '%s' "+
            "and if(%d < 0, b.state != 11, b.state = %d) "+
            "group by a.dev_id order by b.state, num",
        user.ZoneID,
        user.Pools[0],
        vmstate,
        vmstate)

        found, err = user._getLeaseConnStat(sql, true)
    } else {
        found, err = user.getInnerVMLeastConnStat(vmstate)
    }

    lg.FmtInfo("vmstate:%d, found:%t, err:%v", vmstate, found, err)
    return
}

func (user *UserReq) getInnerVMLeastConnStat(vmstate int) (bool, error) {
    sql := fmt.Sprintf(
        "select a.dev_id, a.ip, a.online, b.state, if(b.dev_id is null, 0, count(*)) as num "+
            "from tab_cluster a left join tab_vm_runtime b "+
            "on a.dev_id = b.dev_id "+
            "where a.type != 'backup' "+
            "and a.online = 1 "+
            "and b.login_name = '' "+
            "and b.zone_id = %d "+
            "and b.pool_id = '%s' "+
            "and if(%d < 0, b.state != 11, b.state = %d) "+
            "group by a.dev_id order by b.state, num",
        user.ZoneID,
        user.Pools[0],
        vmstate,
        vmstate)

    return user._getLeaseConnStat(sql, false)
}

//GetOuterVMLeastConnStat add statistic of idle onlined VM
func (user *UserReq) GetOuterVMLeastConnStat(table string) (bool, error) {
    sql := fmt.Sprintf(
        "select a.dev_id, a.ip, a.online, 2, if(b.dev_id is null, 0, count(*)) as num "+
            "from tab_cluster a left join %s b "+
            "on a.dev_id = b.dev_id "+
            "where a.type != 'backup' "+
            "and a.online = 1 "+
            "and b.pool_id = '%s' "+
            "group by a.dev_id order by num",
        table,
        user.Pools[0])

    return user._getLeaseConnStat(sql, true)
}

func (user *UserReq) NormalLeastConn() (bool, error) {
    var err error
    found := false
    sql := fmt.Sprintf(
        "select a.dev_id, a.ip, a.online, b.state, if(b.dev_id is null, 0, count(*)) as num " +
            "from tab_cluster a left join tab_container_runtime b " +
            "on a.dev_id = b.dev_id " +
            "where a.type != 'backup' " +
            "and a.online = 1 " +
            "group by a.dev_id order by num limit 1")
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()
    for rows.Next() {
        var devid, ip string
        if err := rows.Scan(&devid, &ip); err != nil {
            lg.Error("db error:%s", err.Error())
            break
        }
        user.DevIDs = append(user.DevIDs, devid)
        user.IPs = append(user.IPs, ip)
        found = true
    }

    return found, err
}

func (user *UserReq) NormalLeastConnStat() (bool, error) {
    sql := fmt.Sprintf(
        "select a.dev_id, a.ip, a.online, 0, if(b.dev_id is null, 0, count(*)) as num " +
            "from tab_cluster a left join tab_container_runtime b " +
            "on a.dev_id = b.dev_id " +
            "where a.type != 'backup' " +
            "and a.online = 1 " +
            "group by a.dev_id order by num")

    return user._getLeaseConnStat(sql, true)
}

func (user *UserReq) IsSharedVMClientConfiged() (bool, error) {
    sql := fmt.Sprintf("select count(*) "+
        "from tab_client_login_relation b join tab_vm_runtime c "+
        "on b.machine_alias = c.machine_alias "+
        "and b.zone_id = c.zone_id "+
        "and b.client_ip = '%s'"+
        "and b.zone_id = %d",
        user.ClientIP,
        user.ZoneID)
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)
    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()

    cnt := 0
    found := false
    for rows.Next() {
        if err := rows.Scan(&cnt); err != nil {
            lg.Error("db error:%s", err.Error())
            return false, err
        }

        if cnt > 0 {
            found = true
        }
    }

    return found, nil
}

func (user *UserReq) GetSharedVMHost() (bool, error) {
    sql := fmt.Sprintf("select distinct a.dev_id, a.ip "+
        "from tab_cluster a join tab_client_login_relation b join tab_vm_runtime c "+
        "on b.machine_alias = c.machine_alias "+
        "and b.zone_id = c.zone_id "+
        "and a.dev_id = c.dev_id "+
        "and a.online = 1 "+
        "and b.client_ip = '%s'",
        user.ClientIP)
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()

    var devid string
    var ip string
    var found bool
    for rows.Next() {
        if err := rows.Scan(&devid, &ip); err != nil {
            lg.Error("db error:%s", err.Error())
            return false, err
        }
        user.DevIDs = append(user.DevIDs, devid)
        user.IPs = append(user.IPs, ip)
        found = true
    }
    lg.FmtInfo("sharedVM:%v, %v", user.DevIDs, user.IPs)
    if !found {
        lg.Warn("not cluster, using localhost")
        sli := []string{S.AppSetting.DefaultRedirectHost}
        user.IPs = sli
        found = true
    }
    return found, err
}

func (user *UserReq) CheckSharedVM() error {
    sql := fmt.Sprintf("select a.vm_stype, a.image_id from "+
            "tab_auto_login_server a join tab_user_zone_applications b "+
            "on a.zone_id = b.zone_id "+
            "and a.protocol = b.protocol "+
            "and a.pool_id = b.pool_id "+
            "and a.image_id = b.image_id "+
            "where b.zone_id = %d "+
            "and b.loginname = '%s' "+
            "and b.protocol = '%s'",
            user.ZoneID, user.LoginName, user.Protocol)

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return err
    }
    defer rows.Close()

    var isSha bool
    var imageID int
    for rows.Next() {
        if err := rows.Scan(&isSha, &imageID); err != nil {
            lg.Error("db error:%s", err.Error())
            return err
        }
        user.IsSharedVM = isSha
        user.ImageIDs = append(user.ImageIDs, imageID)
    }
    lg.FmtInfo("isSharedVM:%t, %v", user.IsSharedVM, user.ImageIDs)
    return err
}

func (user *UserReq) IsHostOnline() (bool, error) {
    sql := fmt.Sprintf("select online from tab_cluster where dev_id = '%s'", user.DevIDs[0])
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return false, err
    }
    defer rows.Close()

    var online bool
    for rows.Next() {
        if err := rows.Scan(&online); err != nil {
            lg.Error("db error:%s", err.Error())
            return false, err
        }
    }
    lg.FmtInfo("checkOnline:%+v-%+v, %t", user.DevIDs, user.IPs, online)
    return online, err
}

//GetHAIP GetHAIP
func GetHAIP(ip string) string {
    if ip == "localhost" || ip == "127.0.0.1" || ip == "" {
        return "127.0.0.1"
    }

    sql := fmt.Sprintf("select ha_ip from tab_cluster where ip = '%s'", ip)
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return ip
    }
    defer rows.Close()

    var ha string
    for rows.Next() {
        if err := rows.Scan(&ha); err != nil {
            lg.Error("db error:%s", err.Error())
            return ip
        }
    }

    lg.FmtInfo("ip:%s, ha_ip:%s", ip, ha)

    return ha
}

//GetMasterIP
func GetMasterIP() (isCluster bool, ip string, ha string) {
    sql := fmt.Sprintf("select ip, ha_ip from tab_cluster where type = 'manager' and online = 1")
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)
    for rows.Next() {
        if err := rows.Scan(&ip, &ha); err != nil {
            lg.Error("db error:%s", err.Error())
            return
        }
        isCluster = true
    }

    return
}


func PublicNetworkDetect() (bflag bool) {
    sql := fmt.Sprintf("select value from tab_global_config where name = 'public_network_ip_detection_flag'")
    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return
    }
    defer rows.Close()

    var flag string
    for rows.Next() {
        if err := rows.Scan(&flag); err != nil {
            lg.Error("db error:%s", err.Error())
            return
        }
    }
    lg.FmtInfo("public_network_ip_detection_flag:%s", flag)
    if flag == "1" {
        bflag = true
    }
    return
}

func (user *UserReq) ThisIsWrongJob(tab string, hostip string, guestitem string) {
    sql := fmt.Sprintf(
            "delete from %s "+
            "where hostname = '%s' "+
            "and user = '%s' "+
            "and zone_id = %d "+
            "and protocol = '%s' "+
            "and %s = '%s'",
            tab,
            hostip,
            user.LoginName,
            user.ZoneID,
            user.Protocol,
            guestitem,
            user.AutoLS.IP,
        )

    rows, err := db.Raw(sql).Rows()
    lg.FmtInfo("err:%v, sql:%s", err, sql)

    if err != nil {
        lg.Error(err.Error())
        return
    }
    defer rows.Close()
}
