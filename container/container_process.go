package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/root/data/%s"
	ConfigFileName      string = "config.json"
	ContainerLogFile    string = "container.log"
)

type ContainerInfo struct {
	Pid          string   `json:"pid"`        // 容器的 init 进程在宿主机上的 PID
	Id           string   `json:"id"`         // 容器 Id
	Name         string   `json:"name"`       // 容器名
	Command      string   `json:"command"`    // 容器内init运行的命令
	VoluemConfig []string `json:"volumes"`    // 容器 volume 信息
	CreateTime   string   `json:"createTime"` // 创建时间
	Status       string   `json:"status"`     // 容器的状态
}

func NewParentProcess(tty bool, volumeConfigs string, imageName string, envs []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("new pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 配置日志文件
		logDir := fmt.Sprintf(DefaultInfoLocation, ContainerId)
		if err := os.MkdirAll(logDir, 0622); err != nil {
			log.Errorf("mkdir %s error: %v", logDir, err)
			return nil, nil
		}
		stdLogFilePath := filepath.Join(logDir, ContainerLogFile)
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("create file %s error: %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}

	// 传入管道文件读取端的句柄
	cmd.ExtraFiles = []*os.File{readPipe}

	// 配置环境变量
	cmd.Env = append(os.Environ(), envs...)

	// 创建 overlay 文件系统
	imgUrl := filepath.Join("/root/data", "img")
	dataUrl := filepath.Join("/root/data", ContainerId)
	mountUrl := filepath.Join(dataUrl, "merged")
	if err := NewWorkspace(imgUrl, dataUrl, mountUrl, volumeConfigs, imageName); err != nil {
		log.Errorf("overlay filesystem create failed: %v", err)
		return nil, nil
	}

	// 设置根目录为 overlay 文件系统
	cmd.Dir = mountUrl

	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	return os.Pipe()
}
