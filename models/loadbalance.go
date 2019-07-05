package models

import (
	"errors"
	"fmt"
	lg "lbeng/pkg/logging"
)

//LogedOnMapping
func (user *UserReq) GetLogedOnMapping() (found bool, err error) {
	sql := fmt.Sprintf(
		"select distinct a.dev_id, a.hostname from tab_container_runtime a join tab_cluster b "+
			"on a.dev_id = b.dev_id "+
			"where b.type != 'backup' "+
			"and b.online = 1 "+
			"and a.user = '%s' "+
			"and a.zonename = '%s' "+
			"and if('%s' = '', 1, a.protocol = '%s')",
		user.LoginName,
		user.ZoneName,
		user.Protocol,
		user.Protocol)
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%v, sql:%s", err, sql)

	if err != nil {
		lg.Error(err.Error())
		return
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

	return
}

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
			"and b.zone_id=%s "+
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
	for rows.Next() {
		var devid, ip string
		var num int
		if err := rows.Scan(&devid, &ip, &num); err != nil {
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

//GetInnerVMLeastConn
func (user *UserReq) GetInnerVMLeastConn() (bool, error) {
	var found bool
	var err error
	if found, err = user.InnerVMLogedOnMaping(); found == true || err != nil {
		return found, err
	}

	sql := fmt.Sprintf(
		"select a.dev_id, a.ip, if(b.dev_id is null, 0, count(*)) as num "+
			"from tab_cluster a left join tab_vm_runtime b "+
			"on a.dev_id = b.dev_id "+
			"where a.type != 'backup' "+
			"and a.online = 1 "+
			"and b.login_name = '' "+
			"and b.zone_id=%s "+
			"and b.pool_id = '%s' "+
			"and if(b.state = 2, 1, if(b.state=4, 1, 1)) "+ //1. idle, 2. stopped, 3. starting
			"group by a.dev_id order by num limit 1",
		user.ZoneID,
		user.Pools[0])
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)

	if err != nil {
		lg.Error(err.Error())
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var devid, ip, num string
		if err := rows.Scan(&devid, &ip, &num); err != nil {
			lg.Error("db error:%s", err.Error())
			return false, err
		}
		user.DevIDs = append(user.DevIDs, devid)
		user.IPs = append(user.IPs, ip)
		found = true
		lg.FmtInfo("found:%t, devid:%v, ip:%v", found, user.DevIDs, user.IPs)
	}

	return found, nil
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

//GetInnerVMLeastConnStat
//add statistic of idle onlined VM
func (user *UserReq) GetInnerVMLeastConnStat() (bool, error) {
	sql := fmt.Sprintf(
		"select a.dev_id, a.ip, a.online, b.state, if(b.dev_id is null, 0, count(*)) as num "+
			"from tab_cluster a left join tab_vm_runtime b "+
			"on a.dev_id = b.dev_id "+
			"where a.type != 'backup' "+
			"and a.online = 1 "+
			"and b.login_name = '' "+
			"and b.zone_id=%s "+
			"and b.pool_id = '%s' "+
			"group by a.dev_id order by b.state, num",
		user.ZoneID,
		user.Pools[0])

	return user._getLeaseConnStat(sql, false)
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
	lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)

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
		err = errors.New("shared vm not configured")
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
		"where b.zone_id = %s "+
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
	lg.FmtInfo("err:%v, sql:%s, rows:%+v", err, sql, rows)

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

//GetProtocolsAndPools
func (user *UserReq) GetProtocolsAndPools() error {
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
		if len(user.Prots) == 1 {
			user.Protocol = user.Prots[0]
		}
		lg.Info(user.Prots, user.Pools)
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
