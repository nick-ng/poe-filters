package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type weapon struct {
	BaseType      string `json:"baseType"`
	RequiresLevel int    `json:"requiresLevel"`
	Serious       bool   `json:"serious"`
}

var (
	weapons map[string][]weapon
)

func init() {
	weaponJsonBytes, err := os.ReadFile(filepath.Join("utils", "weapons.json"))

	if err != nil {
		fmt.Println("error loading weapons.json", err)
		os.Exit(1)
	}

	err = json.Unmarshal(weaponJsonBytes, &weapons)

	if err != nil {
		fmt.Println("invalid weapons.json file", err)
	}
}

// Creates filters for each weapon in the specified group e.g. fast-2h-axes.
// If the area level is below the weapon's drop level plus seriousAdjustment,
// the weapon will have a larger font and drop sound. Some weapons are not
// "serious" and won't be highlighted. See weapons.json for weapon groups.
func GetWeaponGroupFilter(weaponGroupName string, maxLevel int, minLevel int, seriousAdjustment int, showAdjustment int) (string, error) {
	weaponGroup := weapons[weaponGroupName]

	filterStrings := []string{}

	for _, weapon := range weaponGroup {
		if weapon.RequiresLevel > maxLevel {
			continue
		}

		if weapon.RequiresLevel < minLevel {
			continue
		}

		filterLines := []string{}

		if weapon.Serious && seriousAdjustment >= 0 {
			filterLines = []string{
				"Show",
				fmt.Sprintf("\tAreaLevel <= %d", weapon.RequiresLevel+seriousAdjustment),
				"\tSockets < 6",
				"\tRarity <= Rare",
				fmt.Sprintf("\tBaseType == \"%s\" # requires level %d", weapon.BaseType, weapon.RequiresLevel),
				"\tCorrupted False",
				"\tSetFontSize 45",
				"\t#!LevelingBorder!# 230",
				"\tMinimapIcon 2 Pink Cross",
				"\tCustomAlertSound \"sounds/ben-finegold-this-is-serious.mp3\"",
			}
		}

		if !weapon.Serious || seriousAdjustment < 0 || seriousAdjustment < showAdjustment {
			filterLines = append(filterLines,
				"Show",
				fmt.Sprintf("\tAreaLevel <= %d", weapon.RequiresLevel+showAdjustment),
				"\tSockets < 6",
				"\tRarity <= Rare",
				fmt.Sprintf("\tBaseType == \"%s\" # requires level %d", weapon.BaseType, weapon.RequiresLevel),
				"\tCorrupted False",
				"\tSetFontSize 25",
				"\t#!LevelingBorder!# 230",
				"\tMinimapIcon 1 Pink Cross",
			)
		}

		filterStrings = append(filterStrings, strings.Join(filterLines, "\n"))
	}

	return strings.Join(filterStrings, "\n\n"), nil
}

// @todo(nick-ng): move to "items" file
// @todo(nick-ng): add a way to add custom lines to each filter. #! custom and #! custombig
func GetDropLevelFilter(rawCommand string) (string, error) {
	rawFlags := ParseFlags(rawCommand)

	var itemClass string
	flags := map[string]int{
		"min":  0,
		"max":  99,
		"big":  1,
		"show": 7,
	}
	// @todo(nick-ng): base the defaults on the item class
	levels := []int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80}

	for _, flag := range rawFlags {
		switch flag.Name {
		case "class":
			{
				itemClass = flag.Value
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

	filterStrings := []string{}

	for i, dropLevel := range levels {
		upperDropLevel := 86
		if i < (len(levels) - 1) {
			upperDropLevel = levels[i+1]
		}

		if dropLevel > flags["max"] {
			continue
		}

		if dropLevel < flags["min"] {
			continue
		}

		filterLines := []string{
			// big highlight
			"Show",
			fmt.Sprintf("\tAreaLevel <= %d", dropLevel+flags["big"]),
			"\tRarity <= Rare",
			fmt.Sprintf("\tClass == \"%s\"", itemClass),
			fmt.Sprintf("\tDropLevel >= \"%d\"", dropLevel),
			fmt.Sprintf("\tDropLevel < \"%d\"", upperDropLevel),
			// "\tCorrupted False",
			"\tSetFontSize 45",
			"\t#!LevelingBorder!# 230",
			"\tMinimapIcon 2 Pink Cross",
			// "\tCustomAlertSound \"sounds/ben-finegold-this-is-serious.mp3\"",
			// small highlight
			"Show",
			fmt.Sprintf("\tAreaLevel <= %d", dropLevel+flags["show"]),
			"\tRarity <= Rare",
			fmt.Sprintf("\tClass == \"%s\"", itemClass),
			fmt.Sprintf("\tDropLevel >= \"%d\"", dropLevel),
			fmt.Sprintf("\tDropLevel < \"%d\"", upperDropLevel),
			// "\tCorrupted False",
			"\tSetFontSize 25",
			"\t#!LevelingBorder!# 230",
			"\tMinimapIcon 1 Pink Cross",
		}

		filterStrings = append(filterStrings, strings.Join(filterLines, "\n"))
	}

	return strings.Join(filterStrings, "\n\n"), nil
}
