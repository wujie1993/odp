package util

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type MachineInfo struct {
	OS         string
	CPUCores   int
	MemorySize int
	DiskSize   int
}

type GPUInfo struct {
	ID     int
	Model  string
	UUID   string
	Memory int
}

func RemoteSshCommand(host string, user string, password string, port uint16, command string) (result string, err error) {
	sshHost := host
	sshUser := user
	sshPassword := password
	sshType := "password" //password 或者 key
	sshKeyPath := ""      //ssh id_rsa.id 路径"
	sshPort := port

	//创建sshp登陆配置
	config := &ssh.ClientConfig{
		Timeout:         time.Second, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}
	if sshType == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(sshPassword)}
	} else {
		method, err := publicKeyAuthFunc(sshKeyPath)
		if err != nil {
			return "", err
		}
		config.Auth = []ssh.AuthMethod{method}
	}

	//dial 获取ssh client
	addr := fmt.Sprintf("%s:%d", sshHost, sshPort)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Error("create ssh client failed.", err)
		return "", err
	}
	if sshClient == nil {
		return "", err
	}
	defer sshClient.Close()

	//创建ssh-session
	session, err := sshClient.NewSession()
	if err != nil {
		log.Error("create ssh session failed.", err)
		return "", err
	}
	defer session.Close()
	//执行远程命令
	combo, err := session.CombinedOutput(command)
	if err != nil {
		log.Errorf("failed to exec remote cmd %s in %s. %s", command, addr, err)
		return "", err
	}
	return string(combo), nil
}

func publicKeyAuthFunc(kPath string) (ssh.AuthMethod, error) {
	keyPath, err := homedir.Expand(kPath)
	if err != nil {
		log.Error("find key's home dir failed", err)
		return nil, err
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Error("ssh key file read failed", err)
		return nil, err
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Error("ssh key signer failed", err)
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func GetMachineInfo(host string, user string, password string, port uint16) (MachineInfo, error) {
	info := MachineInfo{}

	command := "cat /etc/redhat-release | awk '{print $4}'; cat /proc/cpuinfo | grep 'processor' | wc -l; free | grep Mem | awk '{print $2}'; df | grep '^/dev/*' | awk '{s+=$2} END {print s}'"
	result, err := RemoteSshCommand(host, user, password, port, command)
	if err != nil {
		return info, err
	}

	all_result := strings.Split(result, "\n")
	os_version := "Centos" + all_result[0]
	cpu, err := strconv.Atoi(all_result[1])
	if err != nil {
		return info, err
	}
	mem, err := strconv.Atoi(all_result[2])
	if err != nil {
		return info, err
	}
	disk, err := strconv.Atoi(all_result[3])
	if err != nil {
		return info, err
	}

	info.OS = os_version
	info.CPUCores = cpu
	info.MemorySize = mem
	info.DiskSize = disk
	return info, nil
}

func GetGPUInfo(host string, user string, password string, port uint16) ([]GPUInfo, error) {
	info := []GPUInfo{}

	command := "if type nvidia-smi >/dev/null 2>&1; then nvidia-smi --query-gpu=index,gpu_uuid,gpu_name,memory.total --format=csv; fi"
	result, err := RemoteSshCommand(host, user, password, port, command)
	if err != nil {
		return info, err
	}
	r := csv.NewReader(strings.NewReader(result))
	records, err := r.ReadAll()
	if err != nil {
		log.Error(err)
	}

	if len(records) > 0 {
		for _, record := range records[1:] {
			id, _ := strconv.Atoi(record[0])
			mem, _ := strconv.Atoi(strings.TrimSpace(record[3][:len(record[3])-3]))
			info = append(info, GPUInfo{
				ID:     id,
				UUID:   strings.TrimSpace(record[1]),
				Model:  strings.TrimSpace(record[2]),
				Memory: mem,
			})
		}
	}
	return info, nil
}
