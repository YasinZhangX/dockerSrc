package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/tabwriter"

	"github.com/YasinZhangX/dockerSrc/container"
	log "github.com/sirupsen/logrus"
)

func ListContainers() {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirs, err := os.ReadDir(dirUrl)
	if err != nil {
		log.Errorf("read dir %s error: %v", dirUrl, err)
		return
	}

	var containers []*container.ContainerInfo
	for _, dir := range dirs {
		tmpContainer, err := getContainerInfoInDir(dir)
		if err != nil {
			log.Errorf("Get container info error %v", err)
			continue
		}
		if tmpContainer != nil {
			containers = append(containers, tmpContainer)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 2, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreateTime)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("flush error: %v", err)
		return
	}
}

func getContainerInfoInDir(dir fs.DirEntry) (*container.ContainerInfo, error) {
	containerId := dir.Name()

	r, _ := regexp.Compile(`^\d{6}$`)
	if r.MatchString(containerId) {
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
	} else {
		return nil, nil
	}
}
