package utils

import (
    lg "lbeng/pkg/logging"
    "sync"
)

//Counter for connection
type counterMap struct {
    lock sync.Mutex // protects following fields
    C    map[string]int
}

func GetCounter() *counterMap {
    var counter = &counterMap{
        C: make(map[string]int),
    }
    return counter
}

//Log counter
func (cnt *counterMap) Log(str string) {
    cnt.lock.Lock()
    lg.FmtInfo("%s: %+v", str, *cnt)
    cnt.lock.Unlock()
}

func (cnt *counterMap) Incr(k string) {
    if k == "" {
        return
    }

    cnt.lock.Lock()
    cnt.C[k]++
    cnt.lock.Unlock()
}

func (cnt *counterMap) IncrUnsafe(k string) {
    if k == "" {
        return
    }

    cnt.C[k]++
}

func (cnt *counterMap) Lock() {
    cnt.lock.Lock()
}

func (cnt *counterMap) Unlock() {
    cnt.lock.Unlock()
}

func (cnt *counterMap) Decr(k string) {
    if k == "" {
        return
    }

    cnt.lock.Lock()
    if cnt.C[k] > 0 {
        cnt.C[k]--
        if cnt.C[k] == 0 {
            delete(cnt.C, k)
        }
    }
    cnt.lock.Unlock()
}

func (cnt *counterMap) HasKey(k string) bool {
    if k == "" {
        return false
    }

    ok := false

    cnt.lock.Lock()
    _, ok = cnt.C[k]
    cnt.lock.Unlock()

    return ok
}
