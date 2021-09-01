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

const (
	START_HOUR            int = 00
	END_HOUR              int = 8
	DOWNLOAD_WAIT_SECONDS     = 5 //120
	TIME_CHECK_SECONDS        = 5 //60
)

func main() {
	downloadFiles()
}

func downloadFiles() {
	nowTime := time.Now()

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
	logFileName := filepath.FromSlash("./logs/logs-" + nowTime.Format("2006-01-02T15:04:05") + ".txt")
	eventLog, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer eventLog.Close()
	log.SetOutput(eventLog)

	log.Println("Process started at:", nowTime.Format(time.RFC3339))

	for true {
		//nowTimeHour := time.Now().Hour()
		//if nowTimeHour >= START_HOUR && nowTimeHour < END_HOUR {
		if true {
			if len(itemMap) == 0 {
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
			cmd := exec.Command("youtube-dl", "-i", item, "-f", "137+140")
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
			time.Sleep(TIME_CHECK_SECONDS * time.Second)
			continue
		}
	}
}

func waitForNextDownload() {
	fmt.Println("waiting...")
	time.Sleep(DOWNLOAD_WAIT_SECONDS * time.Second)
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
