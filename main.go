package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	START_HOUR int = 00
	END_HOUR   int = 8
)

func main() {
	downloadFiles()
}

func downloadFiles() {
	nowTime := time.Now()

	//open completed log
	completedLogName := "./completed_list.json"
	completeLog, err := os.OpenFile(completedLogName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer completeLog.Close()

	//read input list
	listBytes, err := os.ReadFile("./list.txt")
	if err != nil {
		log.Println("error reading list:", err.Error())
		os.Exit(1)
	}
	itemsList := strings.Split(strings.ReplaceAll(string(listBytes), "\r", "\n"), "\n")

	//read completedList
	completedListBytes, err := os.ReadFile("./list.txt")
	if err != nil {
		log.Println("error reading list:", err.Error())
		os.Exit(1)
	}
	completedList := make(map[string]string)
	_ = json.Unmarshal(completedListBytes, &completedList)

	itemMap := make(map[string]string)

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
	logFileName := "logs-" + nowTime.Format("2006-01-02T15:04:05") + ".txt"
	eventLog, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer eventLog.Close()
	log.SetOutput(eventLog)

	log.Println("Process started at:", nowTime.Format(time.RFC3339))

	for true {
		nowTimeHour := time.Now().Hour()
		if nowTimeHour >= START_HOUR && nowTimeHour < END_HOUR {
			fmt.Println("GOT ITEMS: ", itemMap)
			item := ""
			for k := range itemMap {
				item = k
				break
			}
			//do the thing
			fmt.Println("download the file...")
			//delete from map
			delete(itemMap, item)
			//add to completed map
			completedList[item] = time.Now().Format("2006-01-02T15:04:05")
			b, err := json.Marshal(completedList)
			if err != nil {
				log.Println("exiting reason @ list marshal : ", err.Error())
			}
			_, err = completeLog.Write(b)
			if err != nil {
				log.Println("exiting reason @ list write to log : ", err.Error())
			}
			fmt.Println("waiting...")
			time.Sleep(5 * time.Second)
			fmt.Println("Proceed to next download...")
		} else {
			fmt.Println("Waiting till midnight...")
			time.Sleep(5 * time.Second)
			continue
		}
	}

	fmt.Println(time.Now().Hour(), time.Now().Minute())
	os.Exit(1)
	cmd := exec.Command("youtube-dl", "-i", "tZFasLMz7Z0", "-f", "137+140")
	err = cmd.Run()
	if err != nil {
		fmt.Println("ERROR: ", err.Error())
		os.Exit(0)
	}
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
