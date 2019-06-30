package models

import (
	lg "lbeng/pkg/logging"
)

//tab_vm_runtime
type TabVmRuntime struct {
	ID       int    `gorm:"primary_key" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func tabVmRuntime(user UserReq) (bool, error) {
	lg.Info("in tab_vm_runtime")

	return false, nil

}
