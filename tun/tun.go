package tun

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/net-byte/vtun/common/config"
	"github.com/songgao/water"
)

func CreateTun(config config.Config) (iface *water.Interface) {
	c := water.Config{DeviceType: water.TUN}
	iface, err := water.New(c)
	if err != nil {
		log.Fatalln("failed to allocate TUN interface:", err)
	}
	log.Println("interface allocated:", iface.Name())
	configTun(config, iface)
	return iface
}

func configTun(config config.Config, iface *water.Interface) {
	os := runtime.GOOS
	ip, ipNet, err := net.ParseCIDR(config.CIDR)
	if err != nil {
		log.Panicf("error cidr %v", config.CIDR)
	}
	if os == "linux" {
		execCmd("/sbin/ip", "link", "set", "dev", iface.Name(), "mtu", "1500")
		execCmd("/sbin/ip", "addr", "add", config.CIDR, "dev", iface.Name())
		execCmd("/sbin/ip", "link", "set", "dev", iface.Name(), "up")
		if config.Route != "" {
			execCmd("/sbin/ip", "route", "add", config.Route, "dev", iface.Name())
		}
	} else if os == "darwin" {
		execCmd("ifconfig", iface.Name(), "inet", ip.String(), config.Gateway, "up")
		if config.Route != "" {
			execCmd("route", "add", "-net", config.Route, "-interface", iface.Name())
		}
	} else if os == "windows" {
		execCmd("netsh", "interface", "ip", "set", "address", "name="+iface.Name(), "source=static", "addr="+ip.String(), "mask="+ipMask(ipNet.Mask), "gateway=none")
		if config.Route != "" {
			execCmd("netsh", "interface", "ip", "delete", "route", "prefix="+config.Route, "interface="+iface.Name(), "store=active")
			execCmd("netsh", "interface", "ip", "add", "route", "prefix="+config.Route, "interface="+iface.Name(), "store=active")
		}
	} else {
		log.Printf("not support os:%v", os)
	}
}

func execCmd(c string, args ...string) {
	log.Printf("exec cmd: %v %v:", c, args)
	cmd := exec.Command(c, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		log.Println("failed to exec cmd:", err)
	}
}

func ipMask(mask net.IPMask) string {
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}
