package main

import (
	"fmt"
	// "os"
	// "os/exec"
	"path/filepath"
	"strings"
	// "time"

	"github.com/YasinZhangX/dockerSrc/cgroups"
	"github.com/YasinZhangX/dockerSrc/container"
)

func RemoveContainer(containerId string) error {
	containerInfo, err := container.GetContainerInfo(containerId)
	if err != nil {
		return err
	}

	if containerInfo.Status != container.STOP {
		return fmt.Errorf("can not remove running container")
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
