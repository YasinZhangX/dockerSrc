package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/YasinZhangX/dockerSrc/container"
)

func printContainerLog(containerId string) error {
	logDir := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	logFilePath := filepath.Join(logDir, container.ContainerLogFile)
	logFile, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("log container open file %s error: %v", logFilePath, err)
	}
	defer logFile.Close()

	log, err := ioutil.ReadAll(logFile)
	if err != nil {
		return fmt.Errorf("log container read file %s error: %v", logFile.Name(), err)
	}

	fmt.Fprint(os.Stdout, string(log))
	fmt.Println()

	return nil
}
