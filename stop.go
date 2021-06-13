package main

import (
	"regexp"
	"strconv"
	"syscall"

	"github.com/YasinZhangX/dockerSrc/container"
	log "github.com/sirupsen/logrus"
)

func stopContainer(containerId string) {
	containerInfo, err := container.GetContainerInfo(containerId)
	if err != nil {
		log.Errorf("get container info error: %v", err)
		return
	}
	pid := containerInfo.Pid
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.Errorf("convert pid from string to int error: %v", err)
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		r, _ := regexp.Compile(`.*no such process.*`)
		if !r.MatchString(err.Error()) {
			log.Errorf("stop container %s error: %v", containerId, err)
			return
		}
	}

	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	if err := container.RecordContainerInfo(containerInfo); err != nil {
		log.Errorf("record container info error: %v", err)
	}
}
