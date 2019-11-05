package models

import (
	"encoding/json"
	lg "lbeng/pkg/logging"
)

type cluster struct {
	IsCluster     bool     `json:"cluster"`
	IsStable      bool     `json:"stable"`
	IsResult      bool     `json:"result"`
	HA            string   `json:"ha"`
}


func GetClusterFromCache() (clu cluster) {
	k := "cluster"

	res := redb.HGetAll(k).Val()
	lg.FmtInfo("%+v", res)
	if err := json.Unmarshal([]byte(res[k]), &clu); err != nil {
		lg.Info(err.Error())
	}
	lg.FmtInfo("%+v", clu)

	return clu
}

func SetClusterCache(v map[string]interface{}) error {
	k := "cluster"
	lg.FmtInfo("%+v", v)
	return setMapStatus(k, v)
}



func setMapStatus(k string, val map[string]interface{}) error {
	if k == "" {
		lg.Warn("null value")
		return nil
	}

	var vv = make(map[string]interface{})
	vv["hello"] = 1
	vv["hello2"] = 2

	var v = make(map[string]interface {})
	v[k] = vv
	str, err := redb.HMSet(k, vv).Result()
	lg.Info(str, err.Error())

	return err
}