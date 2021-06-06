package container

import (
	"fmt"
	"math/rand"
	"os"
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
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(nameBytes) {
			b[i] = nameBytes[idx]
			i--
		}
		cache >>= letterIdxBits
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
