package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const letterBytes = "0123456789"
const (
	letterIdxBits = 4                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var ContainerId = randContainerId(6)

func randContainerId(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

const nameBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	nameIdxBits = 6                    // 6 bits to represent a letter index
	nameIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	nameIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandContainerName() string {
	namePart1 := randLetterString(5)
	namePart2 := randLetterString(5)
	return namePart1 + "_" + namePart2
}

func randLetterString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for nameIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), nameIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), nameIdxMax
		}
		if idx := int(cache & nameIdxMask); idx < len(nameBytes) {
			b[i] = nameBytes[idx]
			i--
		}
		cache >>= nameIdxBits
		remain--
	}
	return string(b)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func VolumeConfigExtract(volumeConfigs string) []string {
	old := strings.Split(volumeConfigs, ",")

	var new []string
	for _, conf := range old {
		if len(conf) > 0 {
			new = append(new, conf)
		}
	}

	return new
}

func VolumeUrlExtract(volumeConfig string) ([]string, error) {
	volumeUrls := strings.Split(volumeConfig, ":")
	if len(volumeUrls) == 2 && len(volumeUrls[0]) != 0 && len(volumeUrls[1]) != 0 {
		return volumeUrls, nil
	} else {
		return nil, fmt.Errorf("volume url extract error. volume config is %s", volumeConfig)
	}
}

func GetContainerInfo(containerId string) (*ContainerInfo, error) {
	configFileDir := fmt.Sprintf(DefaultInfoLocation, containerId)
	configFileName := filepath.Join(configFileDir, ConfigFileName)
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		err = fmt.Errorf("read file %s error: %v", configFileName, err)
		return nil, err
	}

	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		err = fmt.Errorf("json unmarshal error: %v", err)
		return nil, err
	}

	return &containerInfo, nil
}

func RecordContainerInfo(containerInfo *ContainerInfo) error {
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(DefaultInfoLocation, ContainerId)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		return fmt.Errorf("mkdir %s error: %v", dirUrl, err)
	}

	fileName := filepath.Join(dirUrl, ConfigFileName)
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
