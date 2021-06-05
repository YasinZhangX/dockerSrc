package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/YasinZhangX/dockerSrc/cgroups"
	"github.com/YasinZhangX/dockerSrc/container"
)

func RemoveContainer(containerId string) error {
	if err := deleteCgroup(containerId); err != nil {
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

// 删除文件系统
func deleteFileSystem(containerId string) error {
	dataUrl := filepath.Join("/root/data", containerId)
	mountUrl := filepath.Join(dataUrl, "merged")
	if err := container.DeleteWorkspace(dataUrl, mountUrl, ""); err != nil {
		return fmt.Errorf("overlay filesystem delete failed: %v", err)
	}

	return nil
}
