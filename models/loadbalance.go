package models

import (
	"fmt"
	lg "lbeng/pkg/logging"
)

//LogedOnMapping
func (user *UserReq) GetLogedOnMapping() error {
	var err error

	sql := fmt.Sprintf(
		"select a.dev_id, a.hostname from tab_container_runtime a join tab_cluster b "+
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
	lg.FmtInfo("err:%s, sql：%s", err, sql)
	if err != nil {
		lg.Error(err.Error())
		return err
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
	}

	return err
}

//GetInnerVMLeastConn
func (user *UserReq) GetInnerVMLeastConn() (bool, error) {
	var err error
	found := false
	sql := fmt.Sprintf(
		"select a.dev_id, a.ip, if(b.dev_id is null, 0, count(*)) as num "+
			"from tab_cluster a left join tab_vm_runtime b "+
			"on a.dev_id = b.dev_id "+
			"where a.type != 'backup' "+
			"and a.online = 1 "+
			"and b.login_name = '%s' "+
			"and b.zone_id=%s "+
			"and b.pool_id = '%s' "+
			"group by a.dev_id order by num limit 1",
		user.LoginName,
		user.ZoneID,
		user.Pools[0])
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%s, sql：%s", err, sql)
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

	if found == true {
		lg.Info("found:%s, %s", user.DevIDs, user.IPs)
		return found, nil
	}

	sql = fmt.Sprintf(
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
	rows, err = db.Raw(sql).Rows()
	lg.FmtInfo("err:%s, sql：%s", err, sql)
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
		lg.FmtInfo("found:%s, devid:%s, ip:%s", found, user.DevIDs, user.IPs)
	}

	return found, nil
}

func (user *UserReq) NormalLeastConn() (bool, error) {
	var err error
	found := false
	sql := fmt.Sprintf(
		"select a.dev_id, a.ip, if(b.dev_id is null, 0, count(*)) as num " +
			"from tab_cluster a left join tab_container_runtime b " +
			"on a.dev_id = b.dev_id " +
			"where a.type != 'backup' " +
			"and a.online = 1 " +
			"group by a.dev_id order by num limit 1")
	rows, err := db.Raw(sql).Rows()
	lg.FmtInfo("err:%s, sql：%s", err, sql)
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
	lg.FmtInfo("err:%s, sql：%s", err, sql)
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
	lg.FmtInfo("err:%s, sql：%s", err, sql)
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
	lg.FmtInfo("err:%s, sql：%s", err, sql)
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
	lg.FmtInfo("checkOnline:%v-%v, %t", user.DevIDs, user.IPs, online)
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
		lg.FmtInfo("err:%s, sql：%s", err, sql)
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
	} else {
		sql := fmt.Sprintf(
			"select pool_id from tab_user_zone_applications "+
				"where user = '%s' "+
				"and zone_id = %s "+
				"and protocol = %s",
			user.LoginName,
			user.ZoneID,
			user.Protocol)
		rows, err := db.Raw(sql).Rows()
		lg.FmtInfo("err:%s, sql：%s", err, sql)
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
