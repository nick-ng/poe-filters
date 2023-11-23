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
}

// armourSlots := []armourSlot{{BaseType: "Boots", TtsStrting: "boots", Filename: "boots"}}

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

func GetArmourSocketGroupFilter(socketGroup string) string {
	return ""
}
