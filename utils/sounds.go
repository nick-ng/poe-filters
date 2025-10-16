package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const API_URL = "https://api.streamelements.com/kappa/v2/speech"
const RAW_TTS_SOUNDS_PATH = "raw-tts-sounds"
const TTS_SOUNDS_PATH = "tts-sounds"
const SILENCE_PATH = "sounds/silence.mp3"
const DELAY_SECONDS = 5

var SoundCount int
var MissingSounds map[string]bool
var SilencePath string
var NextTimestamp int64

// type ttsRequest struct {
// 	Text  string `json:"text"`
// 	Voice string `json:"voice"`
// }

// type ttsResponse struct {
// 	Success  bool   `json:"success"`
// 	SpeakUrl string `json:"speak_url"`
// }

func init() {
	SoundCount = 0
	MissingSounds = make(map[string]bool)
	NextTimestamp = 0

	temp, err := filepath.Abs(SILENCE_PATH)

	if err != nil {
		fmt.Println("Something went wrong getting absolute path to silence.mp3", err)
		os.Exit(1)
	}

	SilencePath = temp
}

func ttsDelay() {
	nowSeconds := time.Now().Unix()
	waitSeconds := NextTimestamp - nowSeconds

	if waitSeconds > 0 {
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	NextTimestamp = time.Now().Unix() + DELAY_SECONDS
}

func GetTextToSpeech(text string, filename string, voice string, tempo float64) (string, string, error) {
	MkDirIfNotExist(RAW_TTS_SOUNDS_PATH)
	MkDirIfNotExist(TTS_SOUNDS_PATH)

	rawPath, err := filepath.Abs(filepath.Join(RAW_TTS_SOUNDS_PATH, filename))

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	rawPath = strings.ToLower(rawPath)

	finalPath, err := filepath.Abs(filepath.Join(TTS_SOUNDS_PATH, filename))

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	finalPath = strings.ToLower(finalPath)

	fileStats, err := os.Stat(finalPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", "", err
	}

	if fileStats != nil && fileStats.Size() > 0 {
		return rawPath, finalPath, err
	}

	ttsDelay()

	fmt.Println("getting \"", text, "\"")
	text = strings.ReplaceAll(text, " ", "+")

	client := http.Client{}

	// requestJsonBytes, err := json.Marshal(ttsRequest{
	// 	Text:  text,
	// 	Voice: voice,
	// })

	// if err != nil { fmt.Println(err)
	// 	return "", "", err
	// }

	// bodyReader := bytes.NewReader(requestJsonBytes)

	// req, err := http.NewRequest(http.MethodPost, API_URL, bodyReader)
	url := fmt.Sprintf("%s/?voice=%s&text=%s", API_URL, voice, text)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	// req.Header = http.Header{
	// 	// "Content-Type": {"application/json"},
	// 	"Host":       {"api.streamelements.com"},
	// 	"Accept":     {"*/*"},
	// 	"User-Agent": {"Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0"},
	// }

	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	if res.StatusCode == 400 {
		fmt.Println(req)
		fmt.Println(res)
		os.Exit(1)
	}

	resBody, err := io.ReadAll(res.Body)

	// if err != nil { fmt.Println(err)
	// 	return "", "", err
	// }

	// resObj := ttsResponse{}

	// json.Unmarshal(resBody, &resObj)

	soundBody := resBody

	// fmt.Println(temp)

	// if !resObj.Success {
	// 	return "", "", errors.New("couldn't get a speech url")
	// }

	// soundRes, err := http.Get(resObj.SpeakUrl)

	// if err != nil { fmt.Println(err)
	// 	return "", "", err
	// }

	// soundBody, err := io.ReadAll(soundRes.Body)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	err = os.WriteFile(rawPath, soundBody, 0666)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	// Something happens when re-encoding the file which causes Path of Exile to
	// cut off the last part of the sound. Append 2 seconds of silence to fix.
	fileList := strings.Join([]string{
		fmt.Sprintf("file '%s'", rawPath),
		fmt.Sprintf("file '%s'", SilencePath),
	}, "\n")

	err = os.WriteFile("templist.txt", []byte(fileList), 0666)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", "templist.txt",
		"-filter:a", fmt.Sprintf("atempo=%0.1f,volume=2.1", tempo),
		"-vn",
		finalPath,
	)

	// debug stuff
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	err = cmd.Run()

	SoundCount++

	return rawPath, finalPath, err
}

func MakeTts(rawCommand string, game string) string {
	args := strings.Split(rawCommand, "\"")

	if len(args) < 3 {
		output := fmt.Sprintf("%s\n# warning: need at least text and filename", rawCommand)
		return output
	}

	text := args[1]

	args2 := strings.Split(strings.TrimSpace(args[2]), " ")

	filename := strings.ToLower(args2[0])

	voice := "Brian"
	volume := "300"
	tempo := 1.0

	if len(args2) >= 2 {
		voice = args2[1]
	}
	if len(args2) >= 3 {
		volume = args2[2]
	}
	if len(args2) >= 4 {
		tempTempo, err := strconv.ParseFloat(args2[3], 32)

		if err == nil {
			tempo = tempTempo
		}
	}

	if !strings.HasSuffix(filename, ".mp3") {
		filename = fmt.Sprintf("%s.mp3", filename)
	}

	voicePrefix := fmt.Sprintf("%s-", strings.ToLower(voice))
	if !strings.HasPrefix(filename, voicePrefix) {
		filename = fmt.Sprintf("%s%s", voicePrefix, filename)
	}

	_, path, err := GetTextToSpeech(text, filename, voice, tempo)

	if err != nil {
		output := fmt.Sprintf("%s\n# error: couldn't get sound file\n# %s", rawCommand, err)
		return output
	}

	if runtime.GOOS != "windows" {
		gameDir := GetPoe2SteamPath("tts/")
		if game != "poe2" {
			gameDir = GetPoe1SteamPath("tts/")
		}

		cmd := exec.Command(
			"cp",
			"--update=none",
			path,
			gameDir,
		)

		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()

		if err != nil {
			fmt.Println("error copying sound file to game directory", err, outb.String(), errb.String())
		}

		if game != "poe2" {
			// lutris version
			gameDirLutris := GetPoe1LutrisPath("tts/")
			cmd := exec.Command(
				"cp",
				"--update=none",
				path,
				gameDirLutris,
			)
			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err = cmd.Run()

			if err != nil {
				fmt.Println("error copying sound file to lutris game directory", err, outb.String(), errb.String())
			}
		}

		path = filepath.Join("tts", filename)
	}

	output := fmt.Sprintf("\tCustomAlertSound \"%s\" %s", path, volume)
	return output
}

// @todo(nick-ng): change the path based on operating system
func FixSoundPath(originalLine string) string {
	if runtime.GOOS != "windows" {
		return originalLine
	}
	// I think file names can't contain "
	arguments := strings.Split(originalLine, "\"")

	actualFilePath := arguments[1]

	fileStats, err := os.Stat(actualFilePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Println("a", err)
		MissingSounds[actualFilePath] = true

		arguments[0] = "\tCustomAlertSoundOptional "
		return strings.Join(arguments, "\"")
	}

	if fileStats == nil || fileStats.Size() == 0 {
		MissingSounds[actualFilePath] = true

		arguments[0] = "\tCustomAlertSoundOptional "
		return strings.Join(arguments, "\"")
	}

	absPath, err := filepath.Abs(actualFilePath)

	if err != nil {
		MissingSounds[actualFilePath] = true

		arguments[0] = "\tCustomAlertSoundOptional "
		return strings.Join(arguments, "\"")
	}

	arguments[1] = absPath

	return strings.Join(arguments, "\"")
}

func PrintSoundStats() {
	if SoundCount > 0 {
		fmt.Printf("Sounds requested: %d\n", SoundCount)
	}

	if len(MissingSounds) > 0 {
		fmt.Println("Missing sounds:")

		for missingSound, value := range MissingSounds {
			if value {
				fmt.Printf("- %s\n", missingSound)
			}
		}
	}
}

func GetSoundStats() (int, int) {
	return SoundCount, len(MissingSounds)
}

func ResetSoundStats() {
	SoundCount = 0
	MissingSounds = map[string]bool{}
}
