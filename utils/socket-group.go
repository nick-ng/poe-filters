package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type armourSlot struct {
	BaseType  string
	TtsString string
	Filename  string
	Icon      string
}

const SOCKET_GROUP_TEMPO = 1.8

var armourSlots = []armourSlot{
	{BaseType: "Boots", TtsString: "boots", Filename: "boots", Icon: "Square"},
	{BaseType: "Gloves", TtsString: "gloves", Filename: "gloves", Icon: "Triangle"},
	{BaseType: "Helmets", TtsString: "helm", Filename: "helm", Icon: "Circle"},
	{BaseType: "Body Armours", TtsString: "body", Filename: "body", Icon: "Pentagon"},
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

	if maxCount == 1 && minCount == 1 && len(ttsArray) == 3 {
		return fmt.Sprintf("%s R G B", itemType)
	}

	sort.Slice(ttsArray, func(i, j int) bool {
		return ttsArray[j] < ttsArray[i]
	})

	return fmt.Sprintf("%s %s", itemType, strings.Join(ttsArray, " "))
}

func GetSocketGroupFilter(socketGroup string, args ...string) (string, error) {
	ttsStringPart := " "
	filenamePart := "_"

	if len(args) == 2 {
		filenamePart = args[0]
		ttsStringPart = args[1]
	} else if len(args) == 1 {
		filenamePart = args[0]
		ttsStringPart = args[0]
	}

	ttsString := GetSocketGroupText(socketGroup, ttsStringPart)
	ttsFileName := fmt.Sprintf("%s-%s.mp3", socketGroup, filenamePart)
	_, soundPath, err := GetTextToSpeech(ttsString, ttsFileName, "Brian", SOCKET_GROUP_TEMPO)
	if runtime.GOOS != "windows" {
		gameDir := GetPoe1Path("tts/")
		cmd := exec.Command(
			"cp",
			"--update=none",
			soundPath,
			gameDir,
		)
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err = cmd.Run()
		if err != nil {
			fmt.Println("error copying sound file to game directory", err, outb.String(), errb.String())
		}

		soundPath = strings.ToLower(filepath.Join("tts", ttsFileName))
	}

	if err != nil {
		return "", err
	}

	filterLines := []string{
		"Show",
		"\tSockets < 6",
		"\tRarity <= Rare",
		"\tLinkedSockets <= 4",
		fmt.Sprintf("\tSocketGroup = %s", socketGroup),
		"\t#!LinkBorder!# 230",
		"\t#!LinkBackground!# 255",
		"\tDisableDropSound",
		fmt.Sprintf("\tCustomAlertSound \"%s\" 300", soundPath),
	}

	return strings.Join(filterLines, "\n"), nil
}

func GetArmourSocketGroupFilter(socketGroup string, level ...string) string {
	var filterBlocks []string

	minLevel := "1"
	maxLevel := fmt.Sprintf("%d", (len(socketGroup)*11)-1) // happens to work as a default

	if len(level) == 2 {
		minLevel = level[0]
		maxLevel = level[1]
	} else if len(level) == 1 {
		maxLevel = level[0]
	}

	for _, slot := range armourSlots {
		someFilterLines, err := GetSocketGroupFilter(socketGroup, slot.Filename, slot.TtsString)

		if err != nil {
			continue
		}

		filterLines := []string{
			someFilterLines,
			fmt.Sprintf("\tAreaLevel >= %s", minLevel),
			fmt.Sprintf("\tAreaLevel <= %s", maxLevel),
			fmt.Sprintf("\tClass \"%s\"", slot.BaseType),
			"\tSetFontSize 45",
			fmt.Sprintf("\tMinimapIcon 0 Cyan %s", slot.Icon),
		}

		filterBlocks = append(filterBlocks, strings.Join(filterLines, "\n"))
	}

	return strings.Join(filterBlocks, "\n\n")
}
