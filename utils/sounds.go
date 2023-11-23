package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
func GetTextToSpeech(text string, filename string, voice string, tempo float32) (string, string, error) {
	client := http.Client{}

	requestJsonBytes, err := json.Marshal(ttsRequest{
		Text:  text,
		Voice: voice,
	})

	if err != nil {
		return "", "", err
	}

	bodyReader := bytes.NewReader(requestJsonBytes)

	req, err := http.NewRequest(http.MethodPost, API_URL, bodyReader)

	if err != nil {
		return "", "", err
	}

	req.Header = http.Header{
		"Content-Type": {"application/json"},
	}

	res, err := client.Do(req)

	if err != nil {
		return "", "", err
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		return "", "", err
	}

	resObj := ttsResponse{}

	json.Unmarshal(resBody, &resObj)

	if !resObj.Success {
		return "", "", errors.New("couldn't get a speech url")
	}

	soundRes, err := http.Get(resObj.SpeakUrl)

	if err != nil {
		return "", "", err
	}

	soundBody, err := io.ReadAll(soundRes.Body)

	if err != nil {
		return "", "", err
	}

	MkDirIfNotExist(RAW_SOUNDS_PATH)
	MkDirIfNotExist(SOUNDS_PATH)

	rawPath, err := filepath.Abs(filepath.Join(RAW_SOUNDS_PATH, filename))

	if err != nil {
		return "", "", err
	}

	err = os.WriteFile(rawPath, soundBody, 0666)

	if err != nil {
		return "", "", err
	}

	path, err := filepath.Abs(filepath.Join(SOUNDS_PATH, filename))

	if err != nil {
		return "", "", err
	}

	cmd := exec.Command("ffmpeg", "-y",
		"-i", rawPath, "-filter:a", fmt.Sprintf("\"atempo=%0.1f\"", tempo), "-vn", path)

	fmt.Println("cmd", cmd.String())
	// cmd.Start()
	// cmd.Wait()

	batPath, err := filepath.Abs("temp.bat")

	if err != nil {
		return "", "", err
	}

	fmt.Println(batPath)

	os.WriteFile(batPath, []byte(cmd.String()), 0700)

	cmd = exec.Command("CMD", "/C", batPath)

	err = cmd.Run()

	fmt.Println(err)

	// @todo(nick-ng): figure out how to speed up sounds
	return rawPath, path, err
}
