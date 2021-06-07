package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/YasinZhangX/dockerSrc/container"
	_ "github.com/YasinZhangX/dockerSrc/nsenter"
	log "github.com/sirupsen/logrus"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func ExecContainer(containerId string, cmdArray []string) {
	containerInfo, err := container.GetContainerInfo(containerId)
	if err != nil {
		log.Errorf("exec container get container(%s) info error: %v", containerId, err)
		return
	}
	pid := containerInfo.Pid

	cmdStr := strings.Join(cmdArray, " ")
	log.Infof("container pid %s", pid)
	log.Infof("command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)

	if err := cmd.Run(); err != nil {
		log.Errorf("exec container %s error: %v", containerId, err)
	}
}
