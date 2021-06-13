package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	// "os/signal"
	"path/filepath"
	"strings"

	"github.com/YasinZhangX/dockerSrc/cgroups"
	"github.com/YasinZhangX/dockerSrc/cgroups/subsystems"
	"github.com/YasinZhangX/dockerSrc/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, exitRemove bool, volumeConfigs string, res *subsystems.ResourceConfig,
	containerName string, imageName string, cmdArray []string) {
	// 初始化父进程和写管道
	parent, writePipe := container.NewParentProcess(tty, volumeConfigs, imageName)
	if parent == nil {
		log.Errorf("new parent process error")
		return
	}

	// 启动父进程
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	log.Infof("parent process(%d) start", parent.Process.Pid)

	// 保存容器信息
	if err := recordContainerInfo(parent.Process.Pid, cmdArray, containerName, volumeConfigs); err != nil {
		log.Errorf("record container info error: %v", err)
		return
	}

	// 配置 cgroup
	cgroupManager := cgroups.NewCgroupManager(strings.Join([]string{"mydocker-cgroup", container.ContainerId}, "_"))
	if exitRemove {
		defer cgroupManager.Destory()
	}
	log.Infof("cgroup name is %s", cgroupManager.Path)
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	// 发送用户命令
	sendInitCommand(cmdArray, writePipe)

	if tty {
		parent.Wait()

		if exitRemove {
			// 删除容器信息文件
			DeleteContainerInfo(container.ContainerId)

			// 删除文件系统
			dataUrl := filepath.Join("/root/data", container.ContainerId)
			mountUrl := filepath.Join(dataUrl, "merged")
			if err := container.DeleteWorkspace(dataUrl, mountUrl, volumeConfigs); err != nil {
				log.Errorf("overlay filesystem delete failed: %v", err)
			}
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

func recordContainerInfo(containerPID int, commandArray []string, containerName string, volumeConfigs string) error {
	if len(containerName) == 0 {
		containerName = container.RandContainerName()
	}

	containerInfo := &container.ContainerInfo{
		Id:           container.ContainerId,
		Pid:          strconv.Itoa(containerPID),
		Command:      strings.Join(commandArray, " "),
		VoluemConfig: container.VolumeConfigExtract(volumeConfigs),
		CreateTime:   time.Now().Format("2006-01-02 15:04:05"),
		Status:       container.RUNNING,
		Name:         containerName,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return fmt.Errorf("record container info error: %v", err)
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, container.ContainerId)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		return fmt.Errorf("mkdir %s error: %v", dirUrl, err)
	}

	fileName := filepath.Join(dirUrl, container.ConfigFileName)
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("create file %s error: %v", fileName, err)
	}
	defer file.Close()
	if _, err := file.WriteString(jsonStr); err != nil {
		return fmt.Errorf("file write error: %v", err)
	}

	return nil
}

func DeleteContainerInfo(containerId string) error {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	fileName := filepath.Join(dirUrl, container.ConfigFileName)
	if err := os.Remove(fileName); err != nil {
		return fmt.Errorf("remove file %s error: %v", fileName, err)
	}

	return nil
}
