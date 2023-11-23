package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const API_URL = "https://us-central1-sunlit-context-217400.cloudfunctions.net/streamlabs-tts"
const RAW_SOUNDS_PATH = "raw-sounds"
const SOUNDS_PATH = "sounds"

type ttsRequest struct {
	Text  string `json:"text"`
	Voice string `json:"voice"`
}

type ttsResponse struct {
	Success  bool   `json:"success"`
	SpeakUrl string `json:"speak_url"`
}

// soundPath, err := utils.GetTextToSpeech("body 2 blue", "hello.mp3", 2)

// if err != nil {
// 	fmt.Println("Error when getting sound", err)
// 	os.Exit(1)
// }

// fmt.Println(soundPath)
func GetTextToSpeech(text string, filename string, voice string) (string, error) {
	client := http.Client{}

	requestJsonBytes, err := json.Marshal(ttsRequest{
		Text:  text,
		Voice: voice,
	})

	if err != nil {
		return "", err
	}

	bodyReader := bytes.NewReader(requestJsonBytes)

	req, err := http.NewRequest(http.MethodPost, API_URL, bodyReader)

	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"Content-Type": {"application/json"},
	}

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	resObj := ttsResponse{}

	json.Unmarshal(resBody, &resObj)

	if !resObj.Success {
		return "", errors.New("couldn't get a speech url")
	}

	soundRes, err := http.Get(resObj.SpeakUrl)

	if err != nil {
		return "", err
	}

	soundBody, err := io.ReadAll(soundRes.Body)

	if err != nil {
		return "", err
	}

	MkDirIfNotExist(RAW_SOUNDS_PATH)
	MkDirIfNotExist(SOUNDS_PATH)

	rawPath := filepath.Join(RAW_SOUNDS_PATH, filename)

	err = os.WriteFile(rawPath, soundBody, 0666)

	if err != nil {
		return "", err
	}

	return rawPath, err

	// path := filepath.Join(SOUNDS_PATH, filename)

	// cmd := exec.Command("ffmpeg", "-n",
	// 	"-i", rawPath, "-filter:a", fmt.Sprintf("\"atempo=%0.1f\"", tempo), "-vn", path)

	// fmt.Println("cmd", cmd)

	// stdout, err := cmd.Output()

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println(stdout)

	// return path, err
}

func GetSocketGroupText(socketGroup string, itemType string) string {
	sockets := make(map[string]int)
	sockets["R"] = 0
	sockets["G"] = 0
	sockets["B"] = 0

	for _, c := range socketGroup {
		sockets[string(c)] += 1
	}

	var ttsArray []string

	maxCount := -1
	minCount := 9999

	for k := range sockets {
		if sockets[k] == 0 {
			continue
		}

		if sockets[k] > maxCount {
			maxCount = sockets[k]
		}

		if sockets[k] < minCount {
			minCount = sockets[k]
		}

		var colour string
		switch k {
		case "R":
			{
				colour = "Red"
			}
		case "G":
			{
				colour = "Green"
			}
		case "B":
			{
				colour = "Blue"
			}
		}

		ttsArray = append(ttsArray, fmt.Sprintf("%d %s", sockets[k], colour))
	}

	if len(ttsArray) == 0 {
		return itemType
	}

	if maxCount == 1 && minCount == 1 {
		return fmt.Sprintf("%s R G B", itemType)
	}

	sort.Slice(ttsArray, func(i, j int) bool {
		return ttsArray[j] < ttsArray[i]
	})

	fmt.Println(sockets)
	fmt.Println(ttsArray)

	return fmt.Sprintf("%s %s", itemType, strings.Join(ttsArray, " "))
}
