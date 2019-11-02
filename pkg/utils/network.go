package utils

import (
    "net"

    lg "lbeng/pkg/logging"
)

var privateNetwork = []string{"127.0.0.0/16", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

var ipnets []*net.IPNet

// isPrivateNetwork is private network check
func isPrivateNetwork(ip string) bool {

    if len(ipnets) == 0 {
        for _, ipn := range privateNetwork {
            ipaddr, ipnet, _ := net.ParseCIDR(ipn)
            ipnets = append(ipnets, ipnet)
            lg.Info("Network address: ", ipn)
            lg.Info("IP address     : ", ipaddr)
            lg.Info("ipnet          : ", ipnet)
        }
    }

    for _, ipnet := range ipnets {
        if ipnet.Contains(net.ParseIP(ip)) {
            return true
        }
    }

    return false
}

//IsPublicNetwork IsPublicNetwork
func IsPublicNetwork(ip string) bool {
    return !isPrivateNetwork(ip)
}
