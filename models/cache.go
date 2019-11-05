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
	return setMapStatus(k, v)
}



func setMapStatus(k string, val map[string]interface{}) error {
	if k == "" || val == nil {
		lg.Warn("null value")
		return nil
	}

	str, err := redb.HMSet(k, val).Result()
	if err != nil {
		lg.Info(str, err.Error())
	}

	return err
}