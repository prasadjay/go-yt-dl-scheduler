package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	downloadFiles()
}

type AppConfig struct {
	ForceRun            bool `json:"force_run"`
	StartHour           int  `json:"start_hour"`
	EndHour             int  `json:"end_hour"`
	DownloadWaitSeconds int  `json:"download_wait_seconds"`
	TimeCheckSeconds    int  `json:"time_check_seconds"`
}

var config AppConfig

func downloadFiles() {
	nowTime := time.Now()

	//read config
	configBytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Println("error reading list:", err.Error())
		os.Exit(1)
	}
	fmt.Println("CONFIG : ", string(configBytes))
	_ = json.Unmarshal(configBytes, &config)
	fmt.Println("PARSED CONFIG : ", config)

	//open completed log
	completedLogName := "./completed_list.json"
	completeLog, err := os.OpenFile(completedLogName, os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer completeLog.Close()

	//read input list
	listBytes, err := ioutil.ReadFile("./list.txt")
	if err != nil {
		log.Println("error reading list:", err.Error())
		os.Exit(1)
	}
	itemsList := strings.Split(strings.ReplaceAll(string(listBytes), "\r", "\n"), "\n")

	//read completedList
	completedListBytes, err := ioutil.ReadFile(completedLogName)
	if err != nil {
		log.Println("error reading list:", err.Error())
		os.Exit(1)
	}
	completedList := make(map[string]string)
	err = json.Unmarshal(completedListBytes, &completedList)
	if err != nil {
		fmt.Println("Error reading completed list : ", err.Error())
		return
	}

	itemMap := make(map[string]string)

	if len(itemsList) == 1 && itemsList[0] == "" {
		fmt.Println("Empty list... exiting in 10 seconds...")
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}

	for _, value := range itemsList {
		//value = strings.TrimPrefix(strings.Split(value, "&")[0], "https://www.youtube.com/watch?v=")
		value = strings.Split(value, "&")[0]
		if _, ok := completedList[value]; !ok {
			itemMap[value] = value
		}
	}

	if len(itemMap) == 0 {
		fmt.Println("Nothing to download... exiting in 10 seconds...")
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}

	//set event log
	logFileName := filepath.FromSlash("logs/logs-" + nowTime.Format("2006-01-02T15-04-05") + ".txt")
	eventLog, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer eventLog.Close()
	log.SetOutput(eventLog)

	log.Println("Process started at:", nowTime.Format(time.RFC3339))

	for true {
		nowTimeHour := time.Now().Hour()
		if nowTimeHour >= config.StartHour && nowTimeHour < config.EndHour || config.ForceRun {
			if len(itemMap) == 0 {
				fmt.Println("completed all downloads.. exiting @ ", nowTime.Format("2006-01-02T15:04:05"))
				log.Println("completed all downloads.. exiting @ ", nowTime.Format("2006-01-02T15:04:05"))
				os.Exit(0)
			}

			fmt.Println("Pending items: ", itemMap)
			item := ""
			for k := range itemMap {
				item = k
				break
			}
			//do the thing
			fmt.Println("Downloading item : ", item)
			cmd := exec.Command("youtube-dl", "-i", "\""+item+"\"", "-f", "137+140")
			err = cmd.Run()
			if err != nil {
				fmt.Println("Download error: ", err.Error())
				log.Println("Download error : " + item + " : " + err.Error())
				//delete from map
				delete(itemMap, item)
				waitForNextDownload()
				continue
			}
			//delete from map
			delete(itemMap, item)
			//add to completed map
			completedList[item] = time.Now().Format("2006-01-02T15:04:05")
			b, err := json.Marshal(completedList)
			if err != nil {
				log.Println("exiting reason @ list marshal : ", err.Error())
			}
			//write to completed file
			_ = completeLog.Truncate(0)
			_, _ = completeLog.Seek(0, 0)
			_, err = completeLog.Write(b)
			if err != nil {
				log.Println("exiting reason @ list write to log : ", err.Error())
			}

			waitForNextDownload()
		} else {
			fmt.Println("Waiting till midnight...")
			time.Sleep(time.Duration(config.TimeCheckSeconds) * time.Second)
			continue
		}
	}
}

func waitForNextDownload() {
	fmt.Println("waiting...")
	time.Sleep(time.Duration(config.DownloadWaitSeconds) * time.Second)
	fmt.Println("Proceed to next download...\n")
}

func joinFiles() {
	fmt.Println("STARTING....\n")
	files, err := ioutil.ReadDir("./videos")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), "mp4") {
			videoPath := "./videos/" + f.Name()
			audioPath := "./videos/" + strings.TrimSuffix(f.Name(), "f137.mp4") + "f140.m4a"
			outputName := "./" + strings.TrimSuffix(f.Name(), "f137.mp4") + "mp4"
			cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-c", "copy", outputName)
			err = cmd.Run()
			if err != nil {
				fmt.Println("ERROR: ", videoPath, " : ", err.Error())
				os.Exit(0)
			}
		}
	}
}
