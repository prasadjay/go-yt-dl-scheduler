package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type AppConfig struct {
	ForceRun            bool `json:"force_run"`
	StartHour           int  `json:"start_hour"`
	EndHour             int  `json:"end_hour"`
	DownloadWaitSeconds int  `json:"download_wait_seconds"`
	TimeCheckSeconds    int  `json:"time_check_seconds"`
}

var config AppConfig

const (
	logsDir          = "logs"
	configDir        = "config.json"
	urlListDir       = "list.txt"
	completedListDir = "completed_list.json"
	videosDir        = "downloaded_videos"
)

func init() {
	var err error
	_, err = os.Stat(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(logsDir, 0777)
		}
	}
	_, err = os.Stat(videosDir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(videosDir, 0777)
		}
	}
}

func main() {
	downloadFiles()
}

func downloadFiles() {
	nowTime := time.Now()

	// Read current config.
	configBytes, err := os.ReadFile(configDir)
	if err != nil {
		log.Fatalln("Error reading list:\n\t", err.Error())
	}
	fmt.Println("\nCURRENT CONFIG : \n", string(configBytes))

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatalln("Error unmarshalling", configDir, "\n\t", err)
	}

	// Read input list.
	inputList, err := os.ReadFile(urlListDir)
	if err != nil {
		log.Fatalln("Error reading", urlListDir, "\n\t", err.Error())
	}

	//urlList := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(inputList)), "\r", "\n"), "\n")
	urlList := strings.Split(strings.TrimSpace(string(inputList)), "\n")

	// Open and read a list of completed downloads.
	completedLog, err := os.OpenFile(completedListDir, os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", "\n\t", err)
	}
	defer completedLog.Close()

	completedListBytes, err := os.ReadFile(completedListDir)
	if err != nil {
		log.Fatalln("Error reading", completedListDir, "\n\t", err.Error())
	}

	completedList := make(map[string]string)
	err = json.Unmarshal(completedListBytes, &completedList)
	if err != nil {
		log.Println("Error unmarshalling", completedListDir, "\n\t", err.Error())
		return
	}

	if len(urlList) == 1 && urlList[0] == "" {
		log.Println("[ERROR]:", urlListDir, "is empty!")
		return
	}

	urlMap := make(map[string]string)
	for _, value := range urlList {
		//value = strings.TrimPrefix(strings.Split(value, "\n")[0], "https://www.youtube.com/watch?v=")
		value = strings.Split(strings.TrimSpace(value), "\n")[0]
		if _, ok := completedList[value]; !ok {
			urlMap[value] = value
		}
		if value == "" {
			delete(urlMap, value)
		}
	}

	if len(urlMap) == 0 {
		log.Fatalln("Nothing to download! Add link(s) to the", urlListDir, "file!")
	}

	// Set event log.
	logFileName := filepath.FromSlash(logsDir + "/log-" + nowTime.Format("2006-01-02T15-04-05") + ".txt")
	eventLog, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer eventLog.Close()

	log.SetOutput(eventLog)
	log.Println("Process started at:", nowTime.Format(time.RFC3339))

	var isMidnightSet bool
	for {
		nowTimeHour := time.Now().Hour()
		if nowTimeHour >= config.StartHour && nowTimeHour < config.EndHour || config.ForceRun {
			isMidnightSet = false
			if len(urlMap) == 0 {
				fmt.Println("All downloads completed at", nowTime.Format("2006-01-02 15:04:05"))
				log.Println("All downloads completed at", nowTime.Format("2006-01-02 15:04:05"))
				os.Exit(0)
			}

			var videoUrl string
			for k, _ := range urlMap {
				videoUrl = k
				break
			}

			// Download video
			fmt.Println("Downloading:", videoUrl)
			if len(urlMap)-1 > 0 {
				fmt.Println("Pending URLs:", len(urlMap)-1)
			}

			os.Chdir(videosDir)
			err := exec.Command("youtube-dl", "-i", videoUrl, "-f", "137+140").Run()
			os.Chdir("..")
			if err != nil {
				fmt.Println("Error downloading", videoUrl, "\n\t", err.Error(), "\n")
				log.Println("Error downloading", videoUrl, "\n\t", err.Error())

				delete(urlMap, videoUrl) // Delete broken link from map.

				if len(urlMap) > 0 {
					waitForNextDownload(len(urlMap))
					continue
				} else {
					return
				}
			}

			delete(urlMap, videoUrl) // Delete link from map.
			fmt.Println("Downloaded successfully.\n")

			// Add completed download to list.
			completedList[videoUrl] = time.Now().Format("2006-01-02T15:04:05")
			completedDownload, err := json.Marshal(completedList)
			if err != nil {
				fmt.Println("Error marshalling", completedList, "\n\t", err.Error())
				log.Println("Error marshalling", completedList, "\n\t", err.Error())
			}

			// Write completed download to file.
			completedLog.Truncate(0)
			completedLog.Seek(0, 0)
			_, err = completedLog.Write(completedDownload)
			if err != nil {
				log.Println("Error writing to", completedList, "\n\t", err.Error())
				log.Println("Error writing to", completedList, "\n\t", err.Error())
			}
			waitForNextDownload(len(urlMap))

		} else if !isMidnightSet {
			fmt.Println("\nWaiting till midnight...")
			isMidnightSet = true
			time.Sleep(time.Duration(config.TimeCheckSeconds) * time.Second)
			continue
		}
	}
}

func waitForNextDownload(mapLength int) {
	if mapLength > 1 {
		fmt.Println("Waiting for the next download...")
		time.Sleep(time.Duration(config.DownloadWaitSeconds) * time.Second)
		fmt.Println("Proceed...\n")
	}
}

/*func joinFiles() {
	fmt.Println("STARTING....\n")
	files, err := os.ReadDir(videosDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), "mp4") {
			videoPath := videosDir + f.Name()
			audioPath := videosDir + strings.TrimSuffix(f.Name(), "f137.mp4") + "f140.m4a"
			outputName := videosDir + strings.TrimSuffix(f.Name(), "f137.mp4") + "mp4"
			err := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-c", "copy", outputName).Run()
			if err != nil {
				log.Fatalln("ERROR:", videoPath, ":", err.Error())
			}
		}
	}
}*/
