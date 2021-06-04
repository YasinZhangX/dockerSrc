package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func NewParentProcess(tty bool, volumeConfigs string) (*exec.Cmd, *os.File) {
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
	}
	// 传入管道文件读取端的句柄
	cmd.ExtraFiles = []*os.File{readPipe}

	// 创建 overlay 文件系统
	imgUrl := filepath.Join("/root/data", "img")
	dataUrl := filepath.Join("/root/data", ContainerId)
	mountUrl := filepath.Join(dataUrl, "merged")
	if err := NewWorkspace(imgUrl, dataUrl, mountUrl, volumeConfigs); err != nil {
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

func NewWorkspace(imgUrl string, dataUrl string, mountUrl string, volumeConfigs string) error {
	// create container data directory
	if err := os.Mkdir(dataUrl, 0777); err != nil {
		return fmt.Errorf("mkdir %s error: %v", dataUrl, err)
	}

	err := createLowerLayer(imgUrl, dataUrl)
	if err != nil {
		return err
	}

	err = createUpperLayer(dataUrl)
	if err != nil {
		return err
	}

	err = createMountPoint(imgUrl, dataUrl, mountUrl)
	if err != nil {
		return err
	}

	if len(volumeConfigs) != 0 {
		volumeConfigArr := volumeConfigExtract(volumeConfigs)
		if len(volumeConfigArr) > 0 {
			for _, conf := range volumeConfigArr {
				if volumeUrls, err := volumeUrlExtract(conf); err == nil {
					log.Infof("volume mount: %q", volumeUrls)
					err = mountVolume(mountUrl, volumeUrls)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
		} else {
			return fmt.Errorf("volume parameter input is not correct")
		}
	}

	return nil
}

func createLowerLayer(imgUrl string, dataUrl string) error {
	lowerDir := filepath.Join(imgUrl, "busybox")
	busyboxTarURL := "/root/busybox.tar"
	exist, err := pathExists(lowerDir)
	if err != nil {
		return fmt.Errorf("fail to judge whether dir %s exists: %v", lowerDir, err)
	} else {
		if !exist {
			if err := os.Mkdir(lowerDir, 0777); err != nil {
				return fmt.Errorf("mkdir %s error: %v", lowerDir, err)
			}
			if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", lowerDir).CombinedOutput(); err != nil {
				return fmt.Errorf("untar %s to %s error: %v", busyboxTarURL, lowerDir, err)
			}
		}
	}

	return nil
}

func createUpperLayer(dataUrl string) error {
	upperDir := filepath.Join(dataUrl, "upper")
	if err := os.Mkdir(upperDir, 0777); err != nil {
		return fmt.Errorf("mkdir %s error: %v", upperDir, err)
	}

	return nil
}

func createMountPoint(imgUrl string, dataUrl string, mountUrl string) error {
	// create mount directory
	if err := os.Mkdir(mountUrl, 0777); err != nil {
		return fmt.Errorf("mkdir %s error: %v", mountUrl, err)
	}

	// create worker directory
	workDir := filepath.Join(dataUrl, "worker")
	if err := os.Mkdir(workDir, 0777); err != nil {
		return fmt.Errorf("mkdir %s error: %v", workDir, err)
	}

	// mount overlay filesystem
	lowerDir := filepath.Join(imgUrl, "busybox")
	upperDir := filepath.Join(dataUrl, "upper")
	dirs := "lowerdir=" + lowerDir + ",upperdir=" + upperDir + ",workdir=" + workDir
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mountUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mount error: %v", err)
	}

	return nil
}

func DeleteWorkspace(dataUrl string, mountUrl string, volumeConfigs string) error {
	// process volume configs
	if len(volumeConfigs) != 0 {
		volumeConfigArr := volumeConfigExtract(volumeConfigs)
		if len(volumeConfigArr) > 0 {
			for _, conf := range volumeConfigArr {
				if volumeUrls, err := volumeUrlExtract(conf); err == nil {
					log.Infof("volume umount: %q", volumeUrls)
					err = umountVolume(mountUrl, volumeUrls)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
		} else {
			return fmt.Errorf("volume parameter input is not correct")
		}
	}

	err := deleteMountPoint(dataUrl, mountUrl)
	if err != nil {
		return err
	}

	err = deleteUpperDir(dataUrl)
	if err != nil {
		return err
	}

	// delete container data directory
	if err := os.RemoveAll(dataUrl); err != nil {
		return fmt.Errorf("remove dir %s error: %v", dataUrl, err)
	}

	return nil
}

func umountVolume(mountUrl string, volumeUrls []string) error {
	// umount volume
	containerVolumeUrl := filepath.Join(mountUrl, volumeUrls[1])
	cmd := exec.Command("umount", containerVolumeUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("umount %s error: %v", mountUrl, err)
	}

	return nil
}

func deleteMountPoint(dataUrl string, mountUrl string) error {
	// umount overlay filesystem
	cmd := exec.Command("umount", mountUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("umount %s error: %v", mountUrl, err)
	}

	// delete worker directory
	workDir := filepath.Join(dataUrl, "worker")
	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf("remove dir %s error: %v", workDir, err)
	}

	// delete mount directory
	if err := os.RemoveAll(mountUrl); err != nil {
		return fmt.Errorf("remove dir %s error: %v", mountUrl, err)
	}

	return nil
}

func deleteUpperDir(dataUrl string) error {
	upperDir := filepath.Join(dataUrl, "upper")
	if err := os.RemoveAll(upperDir); err != nil {
		return fmt.Errorf("remove dir %s error: %v", upperDir, err)
	}

	return nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func volumeConfigExtract(volumeConfigs string) []string {
	old := strings.Split(volumeConfigs, ",")

	var new []string
	for _, conf := range old {
		if len(conf) > 0 {
			new = append(new, conf)
		}
	}

	return new
}

func volumeUrlExtract(volumeConfig string) ([]string, error) {
	volumeUrls := strings.Split(volumeConfig, ":")
	if len(volumeUrls) == 2 && len(volumeUrls[0]) != 0 && len(volumeUrls[1]) != 0 {
		return volumeUrls, nil
	} else {
		return nil, fmt.Errorf("volume url extract error. volume config is %s", volumeConfig)
	}
}

func mountVolume(mountUrl string, volumeUrls []string) error {
	// check parent url
	parentUrl := volumeUrls[0]
	if exist, err := pathExists(parentUrl); err == nil {
		if !exist {
			if err := os.Mkdir(parentUrl, 0777); err != nil {
				return fmt.Errorf("mkdir %s error: %v", parentUrl, err)
			}
		}
	} else {
		return err
	}

	// check container url
	containerUrl := volumeUrls[1]
	containerVolumeUrl := filepath.Join(mountUrl, containerUrl)
	if exist, err := pathExists(containerVolumeUrl); err == nil {
		if !exist {
			if err := os.Mkdir(containerVolumeUrl, 0777); err != nil {
				return fmt.Errorf("mkdir %s error: %v", containerVolumeUrl, err)
			}
		}
	} else {
		return err
	}

	// bind mount volume
	if err := syscall.Mount(parentUrl, containerVolumeUrl, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount volume(%s to %s) error: %v", parentUrl, containerVolumeUrl, err)
	}

	return nil
}
