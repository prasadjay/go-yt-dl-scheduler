package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("youtube-dl", "-i", "tZFasLMz7Z0", "-f", "137+140")
	err := cmd.Run()
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
