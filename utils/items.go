package utils

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var defaultDropLevels = []int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80}

// returns drop levels and "isExact", whether the drop levels correspond to
// actual items in the game
// class is the in-game name of the class. e.g. "Two Hand Maces", "Body Armours"
// group is evasion, energyshield, fast, phys, etc.
func getDropLevels(itemClass string, group string) ([]int, bool) {
	itemClassLowerCase := strings.ToLower(itemClass)
	groupLowerCase := strings.ToLower(group)

	switch itemClassLowerCase {
	case ("body armours"):
		{
			evasionLevels := []int{1, 11, 16, 22, 26, 33, 36, 45, 48, 51, 55, 59, 62, 65, 70, 75}
			evasionESLevels := []int{1, 11, 16, 28, 33, 39, 45, 48, 51, 55, 59, 62, 65, 70, 75}
			temp := [][]int{evasionLevels, evasionESLevels}
			var allLevels []int

			for _, dropLevels := range temp {
				for _, level := range dropLevels {
					if slices.Contains(allLevels, level) {
						allLevels = append(allLevels, level)
					}
				}
			}

			switch groupLowerCase {
			case "dex":
				fallthrough
			case "ev":
				fallthrough
			case "evasion":
				{
					return evasionLevels, true
				}
			case "dexint":
				fallthrough
			case "eves":
				fallthrough
			case "evasionenergyshield":
				{
					return evasionESLevels, true
				}
			default:
				{
					return allLevels, true
				}
			}
		}
	default:
		{
			return defaultDropLevels, false
		}
	}
}

// @todo(nick-ng): since you have custom styles, the item class can also be part of the custom styles...
func GetDropLevelFilter(rawCommand string, customStyles []string, bigStyles []string) (string, error) {
	rawFlags := ParseFlags(rawCommand)

	var itemClass string
	var itemGroup string
	flags := map[string]int{
		"min":  0,
		"max":  99,
		"big":  1,
		"show": 7,
	}

	var levels []int
	isCustomLevels := false
	isExact := false

	for _, flag := range rawFlags {
		switch flag.Name {
		case "class":
			{
				itemClass = flag.Value
			}
		case "group":
			{
				itemGroup = flag.Value
			}
		case "levels":
			{
				levelsStrings := strings.Split(flag.Value, ",")
				var tempLevels []int
				for _, levelString := range levelsStrings {
					dropLevel, err := strconv.ParseInt(levelString, 10, 0)

					if err != nil {
						continue
					}

					tempLevels = append(tempLevels, int(dropLevel))
				}

				if len(tempLevels) == 0 {
					continue
				}

				slices.Sort(tempLevels)

				levels = tempLevels
				isCustomLevels = true
				isExact = true
			}
		default:
			{
				flagValueInt, err := strconv.ParseInt(flag.Value, 10, 0)
				if err != nil {
					continue
				}

				flags[flag.Name] = int(flagValueInt)
			}
		}
	}

	if !isCustomLevels {
		levels, isExact = getDropLevels(itemClass, itemGroup)
	}

	customStyleChunk := "\tRarity <= Rare\n\tSetFontSize 25\n\t#!LevelingBorder!# 230\n\tMinimapIcon 1 Pink Cross\n"
	if len(customStyles) > 0 {
		temp := strings.Join(customStyles, "\n\t")
		customStyleChunk = fmt.Sprintf("\t%s\n", temp)
	}

	bigStyleChunk := "\tRarity <= Rare\n\tSetFontSize 45\n\t#!LevelingBorder!# 230\n\tMinimapIcon 2 Pink Cross"
	if len(bigStyles) > 0 {
		var tempBigPropertyNames []string
		for _, style := range bigStyles {
			temp := strings.Split(style, " ")
			tempBigPropertyNames = append(tempBigPropertyNames, temp[0])
		}
		var existingStyles []string
		for _, style := range customStyles {
			temp := strings.Split(style, " ")
			if !slices.Contains(tempBigPropertyNames, temp[0]) {
				existingStyles = append(existingStyles, style)
			}
		}

		temp1 := strings.Join(existingStyles, "\n\t")
		temp2 := strings.Join(bigStyles, "\n\t")
		bigStyleChunk = fmt.Sprintf("\t%s\n\t%s\n", temp1, temp2)
	}

	var filterStrings []string
	for i, dropLevel := range levels {
		upperDropLevel := 86

		if dropLevel > flags["max"] {
			continue
		}

		if dropLevel < flags["min"] {
			continue
		}

		dropLevelChunk := fmt.Sprintf("\tDropLevel == \"%d\"", dropLevel)
		if !isExact {
			if i < (len(levels) - 1) {
				upperDropLevel = levels[i+1]
			}

			dropLevelChunk = fmt.Sprintf("\tDropLevel >= \"%d\"\n\tDropLevel < \"%d\"", dropLevel, upperDropLevel)
		}

		filterLines := []string{
			// big highlight
			"Show",
			fmt.Sprintf("\tAreaLevel <= %d", dropLevel+flags["big"]),
			fmt.Sprintf("\tClass == \"%s\"", itemClass),
			dropLevelChunk,
			bigStyleChunk,
			// small highlight
			"Show",
			fmt.Sprintf("\tAreaLevel <= %d", dropLevel+flags["show"]),
			fmt.Sprintf("\tClass == \"%s\"", itemClass),
			dropLevelChunk,
			customStyleChunk,
		}

		filterStrings = append(filterStrings, strings.Join(filterLines, "\n"))
	}

	return strings.Join(filterStrings, "\n\n"), nil
}
