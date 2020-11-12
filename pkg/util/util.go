package util

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/wujie1993/waves/pkg/setting"
)

// Setup Initialize the util
func Setup() {
	jwtSecret = []byte(setting.AppSetting.JwtSecret)
}

func CheckIPAddressType(ip string) bool {
	if net.ParseIP(ip) == nil {
		fmt.Printf("Invalid IP Address: %s\n", ip)
		return false
	}
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			fmt.Printf("Given IP Address %s is IPV4 type\n", ip)
			return true
		}
	}
	return false
}

func RunCommand(comm string) {
	command := exec.Command("bash", "-c", comm)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	e := command.Run()
	if e != nil {
		panic(e)
	}

}
