package client

import (
	"fmt"
	pb "github.com/bytegang/pb"
	"net"
	"strconv"
	"time"
)

func init() {
	registerAction(new(actionSshScan))
}

var _ ActionDoer = new(actionSshScan)

type actionSshScan struct{}

func (a actionSshScan) Help() (cmd, alias, log string) {
	return "scan", "s", "扫描网段中开放的端口 scan 172.24.11.104/16 22, 扫描CIDR网段 192.168.1.1/24 端口22"
}
func (a actionSshScan) Allow(role pb.UserRole) bool {
	return true
}
func (a actionSshScan) Exec(c *Client, args []string) error {
	cidr := args[1]
	port := args[2]

	ips, err2 := ipArray(cidr)
	if err2 != nil {
		return err2
	}
	var list []string
	for _, ip := range ips {
		_, err := net.DialTimeout("tcp", ip+":"+port, time.Millisecond*2)
		if err != nil {
			//c.Danger(err.Error())
			//return err
		} else {
			list = append(list, ip)
			c.Success(ip)
		}
	}
	msg := fmt.Sprintf("扫描%s(%d个IP),其中%d个IP开启了,端口%s", cidr, len(ips), len(list), port)
	c.Primary(msg)
	return nil
}

func (a actionSshScan) Hint(args *[]string) string {
	if len(*args) < 3 {
		return "参数长度错误 必须包含 CIDR/IP PORT:eg scan 10.13.84.200 22"
	}
	_, _, err := net.ParseCIDR((*args)[1])
	if err != nil {
		return "第一个参数必须是 CIDR 或者 IP"
	}
	i, err := strconv.Atoi((*args)[2])
	if err != nil {
		return "第二个参数必须是是数字"
	}
	if i < 1 || i > 1<<16 {
		return "无效端口"
	}
	return ""
}

func ipArray(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
