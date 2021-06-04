package main

import (
	"fmt"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func CommitContainer(containerId string, imageName string) {
	mntURL := filepath.Join("/root/data", containerId, "merged")
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
