package port_pre_alloc

import (
    "net/http"
    "time"
    "sync"
    "fmt"

    "github.com/gin-gonic/gin"

    M "lbeng/models"
    lg "lbeng/pkg/logging"
)

//port pre-alloc mapping
type PortMap struct {
    lock sync.Mutex // protects following fields
    p    map[string]map[string]int
}

var portMap = PortMap {
    p: make(map[string]map[string]int),
}

func GetPort(c *gin.Context) {
    var pa M.PortPreAlloc

    tstart := time.Now()

    c.BindJSON(&pa)

    pa.Port = -1

    portMap.lock.Lock()
    lg.Info("enlock")

    err := pa.GetUsedPort()
    if err != nil {
        lg.Error(err.Error())
    } else {

        submap := portMap.getSubMap(pa.GuestIP)

        for i:=pa.Start; i<(pa.Start+pa.Cap); i+=pa.Step {
            if !pa.Ports[i] {
                used := false

                for k,v := range submap {
                    if i == v {
                        used = true
                        lg.Info("used, k:%s, v:%d", k, v)
                        break
                    }
                    lg.Info("used, k:%s, v:%d", k, v)
                }

                if !used {
                    pa.Port = i
                    k := fmt.Sprintf("%s_%d_%s_%s", pa.User, pa.Zone, pa.Protocol, pa.GuestIP)
                    submap[k] = i
                    break
                }
            }
        }
    }

    lg.Info("unlock")
    portMap.lock.Unlock()

    c.JSON(http.StatusOK, pa)
    elapsed := time.Since(tstart)
    lg.FmtInfo("%+v", portMap)
    lg.FmtInfo("request elapsed:%d ms, %+v", elapsed/time.Millisecond, pa)
    return
}

func (pm PortMap) getSubMap(ip string) map[string]int {
    submap, ok := portMap.p[ip]
    if !ok {
        submap = make(map[string]int)
        portMap.p[ip] = submap
    }

    return submap
}

func FreePort(ip string, subk string) {
    if ip == "" || subk == "" {
        lg.FmtInfo("k:%s, subk:%s, %+v", ip, subk, portMap)
        return
    }

    portMap.lock.Lock()
    lg.Info("enlock")

    submap := portMap.getSubMap(ip)
    _, ok := submap[subk]
    if ok {
        delete(submap, subk)
    }

    lg.Info("unlock")
    portMap.lock.Unlock()

    lg.FmtInfo("%+v", portMap)

    return
}
