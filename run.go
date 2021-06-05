package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/YasinZhangX/dockerSrc/cgroups"
	"github.com/YasinZhangX/dockerSrc/cgroups/subsystems"
	"github.com/YasinZhangX/dockerSrc/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, volumeConfigs string, res *subsystems.ResourceConfig, containerName string, cmdArray []string) {
	// 加入 channel 接受系统信号，实现优雅退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// 初始化父进程和写管道
	parent, writePipe := container.NewParentProcess(tty, volumeConfigs)
	if parent == nil {
		log.Errorf("new parent process error")
		return
	}

	// 启动父进程
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	log.Infof("parent process(%d) start", parent.Process.Pid)

	// 配置 cgroup
	cgroupManager := cgroups.NewCgroupManager(strings.Join([]string{"mydocker-cgroup", container.ContainerId}, "_"))
	if tty {
		defer cgroupManager.Destory()
	}
	log.Infof("cgroup name is %s", cgroupManager.Path)
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	// 发送用户命令
	sendInitCommand(cmdArray, writePipe)

	if tty {
		parent.Wait()

		// 删除文件系统
		dataUrl := filepath.Join("/root/data", container.ContainerId)
		mountUrl := filepath.Join(dataUrl, "merged")
		if err := container.DeleteWorkspace(dataUrl, mountUrl, volumeConfigs); err != nil {
			log.Errorf("overlay filesystem delete failed: %v", err)
		}
	}
}

// 发送用户命令
func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Infof("all command is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
