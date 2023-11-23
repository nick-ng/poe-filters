package utils

import (
	"fmt"
	"sort"
	"strings"
)

type armourSlot struct {
	BaseType   string
	TtsStrting string
	Filename   string
	Icon       string
}

var armourSlots = []armourSlot{
	{BaseType: "Boots", TtsStrting: "boots", Filename: "boots", Icon: "Square"},
	{BaseType: "Gloves", TtsStrting: "gloves", Filename: "gloves", Icon: "Triangle"},
	{BaseType: "Helmets", TtsStrting: "helm", Filename: "helm", Icon: "Circle"},
	{BaseType: "Body Armours", TtsStrting: "body", Filename: "body", Icon: "Pentagon"},
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
		return fmt.Sprintf("%s R G B. %s", itemType, itemType)
	}

	sort.Slice(ttsArray, func(i, j int) bool {
		return ttsArray[j] < ttsArray[i]
	})

	return fmt.Sprintf("%s %s. %s", itemType, strings.Join(ttsArray, " "), itemType)
}

func GetArmourSocketGroupFilter(socketGroup string, level ...string) string {
	var filterBlocks []string

	minLevel := "1"
	maxLevel := "44"

	if len(level) == 2 {
		minLevel = level[0]
		maxLevel = level[1]
	} else if len(level) == 1 {
		maxLevel = level[0]
	}

	for _, slot := range armourSlots {
		ttsString := GetSocketGroupText(socketGroup, slot.TtsStrting)

		_, soundPath, err := GetTextToSpeech(ttsString, fmt.Sprintf("%s-%s.mp3", socketGroup, slot.Filename), "Brian", 2.2)

		if err != nil {
			continue
		}

		filterLines := []string{
			"Show",
			fmt.Sprintf("\tAreaLevel >= %s", minLevel),
			fmt.Sprintf("\tAreaLevel <= %s", maxLevel),
			"\tSockets < 6",
			"\tRarity <= Rare",
			"\tLinkedSockets <= 4",
			fmt.Sprintf("\tSocketGroup = %s", socketGroup),
			fmt.Sprintf("\tClass \"%s\"", slot.BaseType),
			"\t#!LinkText!#",
			"\t#!LinkBorder!#",
			"\t#!LinkBackground!#",
			"\tDisableDropSound",
			fmt.Sprintf("\tCustomAlertSound \"%s\" 300", soundPath),
			fmt.Sprintf("\tMinimapIcon 0 Cyan %s", slot.Icon),
		}

		filterBlocks = append(filterBlocks, strings.Join(filterLines, "\n"))
	}

	return strings.Join(filterBlocks, "\n\n")
}
