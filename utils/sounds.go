package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
func GetTextToSpeech(text string, filename string, tempo float32) (string, error) {
	client := http.Client{}

	requestJsonBytes, err := json.Marshal(ttsRequest{
		Text:  "body 3 red",
		Voice: "Brian",
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
