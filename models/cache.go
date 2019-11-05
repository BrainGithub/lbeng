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


func GetClusterFromCache() (clu Cluster) {
	k := "cluster"

	res, _ := redb.Do("GET", k).Result()
	lg.FmtInfo("%+v", res)
	if err := json.Unmarshal(res.([]byte), &clu); err != nil {
		lg.Info(err.Error())
	}
	lg.FmtInfo("%+v", clu)

	return clu
}

func SetClusterCache(v Cluster) (err error) {
	k := "cluster"
	lg.FmtInfo("%+v", v)

	v.IsStable = true
	v.HA = "1.1.1.1"

	dat, _ := json.Marshal(v)
	if err = redb.Do("SET", k, dat).Err(); err != nil {
		lg.Error(err.Error())
	}

	return err
}