package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
