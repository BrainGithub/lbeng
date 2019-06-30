package models

import (
	"errors"
	"fmt"
	lg "lbeng/pkg/logging"
)

//LogedOnMapping
func (user *UserReq) GetLogedOnMapping() error {
	var err error
	if user.LoginName != "" && user.ZoneName != "" && user.Protocol != "" {
		sql := fmt.Sprintf(
			"select dev_id, hostname from tab_container_runtime a join tab_cluster b "+
				"on a.dev_id = b.dev_id "+
				"where b.type != 'backup' "+
				"and user = '%s' "+
				"and zonename = '%s' "+
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
	} else {
		emsg := fmt.Sprintf("user name/zone or protocol missed:%+v", user)
		lg.Error(emsg)
		err = errors.New(emsg)
	}

	return err
}

func (user *UserReq) GetInnerVMLeastConn() error {
	// var err error
	return nil
}

func (user *UserReq) GetInnerDockerLeastConn() error {
	return nil
}

func (user *UserReq) GetProtocols() error {
	sql := fmt.Sprintf(
		"select protocol, pool_id from tab_user_zone_applications "+
			"and user = '%s' "+
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
	return nil
}
