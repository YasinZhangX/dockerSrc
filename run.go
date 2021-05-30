package main

import (
	"os"
	"os/signal"
	"strings"

	"github.com/YasinZhangX/dockerSrc/cgroups"
	"github.com/YasinZhangX/dockerSrc/cgroups/subsystems"
	"github.com/YasinZhangX/dockerSrc/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig) {
	// 加入管道接受系统信号，实现优雅退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// 初始化父进程和写管道
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("new parent process error")
		return
	}

	// 启动父进程
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// 配置 cgroup
	cgroupManager := cgroups.NewCgroupManager(strings.Join([]string{"mydocker-cgroup", RandStringBytesMaskImpr(6)}, "_"))
	defer cgroupManager.Destory()
	log.Infof("cgroup name is %s", cgroupManager.Path)
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	// 发送用户命令
	sendInitCommand(cmdArray, writePipe)

	parent.Wait()
}

// 发送用户命令
func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Infof("all command is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
