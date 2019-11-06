package models

import (
	"encoding/json"
	lg "lbeng/pkg/logging"
)

type Cluster struct {
	IsCluster     bool     `json:"cluster"`
	IsStable      bool     `json:"stable"`
	IsResult      bool     `json:"result"`
	HA            string   `json:"ha"`
}

const TIMEOUT = 5


func GetClusterFromCache() (clu Cluster) {
	k := "cluster"

	res, err := redb.Do("GET", k).String()
	lg.FmtInfo("%+v, %v", res, err)
	if err == nil {
		if err := json.Unmarshal([]byte(res), &clu); err != nil {
			lg.Info(err.Error())
		}
	}
	lg.FmtInfo("%+v", clu)

	return clu
}

func SetClusterCache(v Cluster) (err error) {
	k := "cluster"
	lg.FmtInfo("%+v", v)

	var dat []byte
	dat, err = json.Marshal(v)
	if err != nil {
		return
	}

	if err = redb.Do("SET", k, dat, "EX", TIMEOUT).Err(); err != nil {
		lg.Error(err.Error())
	}

	return err
}