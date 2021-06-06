package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/YasinZhangX/dockerSrc/cgroups"
	"github.com/YasinZhangX/dockerSrc/container"
)

func RemoveContainer(containerId string) error {
	containerInfo, err := getContainerInfo(containerId)
	if err != nil {
		return err
	}

	if containerInfo.Status != container.Exit {
		cmd := exec.Command("kill", containerInfo.Pid)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("kill container process %s error: %v", containerInfo.Pid, err)
		}

		time.Sleep(time.Millisecond * 100)
	}

	if err := deleteCgroup(containerId); err != nil {
		return err
	}

	if err := deleteContainerInfo(containerId); err != nil {
		return err
	}

	if err := deleteFileSystem(containerId); err != nil {
		return err
	}

	return nil
}

// 删除 Cgroup 配置
func deleteCgroup(containerId string) error {
	cgroupManager := cgroups.NewCgroupManager(strings.Join([]string{"mydocker-cgroup", containerId}, "_"))
	return cgroupManager.Destory()
}

// 删除容器信息文件
func deleteContainerInfo(containerId string) error {
	return DeleteContainerInfo(containerId)
}

// 删除文件系统
func deleteFileSystem(containerId string) error {
	dataUrl := filepath.Join("/root/data", containerId)
	mountUrl := filepath.Join(dataUrl, "merged")
	if err := container.DeleteWorkspace(dataUrl, mountUrl, ""); err != nil {
		return fmt.Errorf("overlay filesystem delete failed: %v", err)
	}

	return nil
}

func getContainerInfo(containerId string) (*container.ContainerInfo, error) {
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFileName := filepath.Join(configFileDir, container.ConfigFileName)
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		err = fmt.Errorf("read file %s error: %v", configFileName, err)
		return nil, err
	}

	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		err = fmt.Errorf("json unmarshal error: %v", err)
		return nil, err
	}

	return &containerInfo, nil
}
