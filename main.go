package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// Repository git 仓库名称
type Repository struct {
	Name string `json:"name"`
}

// GithubJSON github 发送的webhooks 中的结构
type GithubJSON struct {
	Repository Repository `json:"repository"`
	Ref        string     `json:"ref"`
}

// Config 存放多个仓库配置
type Config struct {
	Hooks []Hook
}

// Hook 对应 hooks.json 中的配置
type Hook struct {
	Repo   string
	Branch string
	Shell  string
}

func loadConfig(configFile *string) {
	var config Config
	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(config.Hooks); i++ {
		addHandler(config.Hooks[i].Repo, config.Hooks[i].Branch, config.Hooks[i].Shell)
	}
}

func setLog(logFile *string) {
	log_handler, err := os.OpenFile(*logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		panic("cannot write log")
	}

	log.SetOutput(log_handler)
	log.SetFlags(5)
}

func startWebserver() {
	log.Println("starting webserver on port: " + *port)
	http.ListenAndServe(":"+*port, nil)
}

func addHandler(repo, branch, shell string) {
	uri := branch
	branch = "refs/heads/" + branch
	http.HandleFunc("/"+repo+"/"+uri, func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var data GithubJSON
		err := decoder.Decode(&data)

		if err != nil {
			log.Println(err)
		}

		if data.Repository.Name == repo && data.Ref == branch {
			executeShell(shell)
		}
	})
}

func executeShell(shell string) {
	fmt.Println(shell)
	out, err := exec.Command(shell).Output()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Shell output was: %s\n", out)
}

var (
	// Set port number
	port = flag.String("port", "7654", "port to listen on")
	// Set config path
	configFile = flag.String("config", "./example.json", "config")
	// Set log path
	logFile = flag.String("log", "./log", "log file")
)

func init() {
	flag.Parse()
	setLog(logFile)
	loadConfig(configFile)
}

func main() {
	startWebserver()
}
