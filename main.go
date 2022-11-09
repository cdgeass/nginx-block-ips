package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
)

type Config struct {
	// nginx log 地址
	LogPath string `yaml:"logPath"`
	// 生成文件地址
	FilePath string `yaml:"filePath"`
	// 正则
	RegExp string `yaml:"regExp"`
	// 完成后执行 shell 命令
	Command string `yaml:"command"`
}

func main() {
	config, err := prepareConfig()
	if err != nil {
		return
	}

	err = generateBlockIps(config)
	if err != nil {
		return
	}

	afterGenerate(config)
}

// 读取配置文件
func prepareConfig() (Config, error) {
	var configPath string
	flag.StringVar(&configPath, "c", "./config.yaml", "配置文件路径")
	flag.Parse()

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
		return Config{}, err
	}

	config := Config{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalln(err)
		return Config{}, err
	}
	return config, err
}

// 读取 nginx log 生成黑名单文件
func generateBlockIps(config Config) error {
	logFile, err := os.Open(config.LogPath)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	regExp, err := regexp.Compile(config.RegExp)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	reader := bufio.NewReader(logFile)
	var ips []string
	for {
		line, _, err := reader.ReadLine()
		if err != nil || line == nil {
			break
			return err
		}

		subIps := regExp.FindAllStringSubmatch(string(line), 1)
		if subIps != nil && len(subIps) > 0 {
			ips = append(ips, subIps[0][1])
		}
	}

	if len(ips) > 0 {
		file, err := os.OpenFile(config.FilePath, os.O_CREATE, 0664)
		if err != nil {
			log.Fatalln(err)
			return err
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		for _, ip := range ips {
			_, err := writer.WriteString(fmt.Sprintf("deny %v;\n", ip))
			if err != nil {
				log.Fatalln(err)
				return err
			}
		}
		err = writer.Flush()
		if err != nil {
			log.Fatalln(err)
			return err
		}
	}
	return nil
}

// 生成之后执行
func afterGenerate(config Config) {
	if len(config.Command) > 0 {
		bash := exec.Command("/bin/bash", "-c", config.Command)
		//bash := exec.Command("cmd", "/C", config.Command)
		output, err := bash.CombinedOutput()
		if err != nil {
			log.Fatalln(err)
			return
		}
		log.Println(string(output))
	}
}
