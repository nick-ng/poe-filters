package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	colourTokens map[string]string
	armourTokens map[string]map[string][][]string
)

func init() {
	colourJsonBytes, err := os.ReadFile(filepath.Join("utils", "colour-tokens.json"))

	if err != nil {
		fmt.Println("error loading colour-tokens.json", err)
		os.Exit(1)
	}

	err = json.Unmarshal(colourJsonBytes, &colourTokens)

	if err != nil {
		fmt.Println("invalid colour-tokens.json file", err)
		os.Exit(1)
	}

	armourJsonBytes, err := os.ReadFile(filepath.Join("utils", "armour-tokens.json"))

	if err != nil {
		fmt.Println("error loading armour-tokens.json", err)
		os.Exit(1)
	}

	err = json.Unmarshal(armourJsonBytes, &armourTokens)

	if err != nil {
		fmt.Println("invalid armour-tokens.json file", err)
		os.Exit(1)
	}
}

func ApplyAllTokens(rawFilter string) string {
	processedFilter := applyColourTokens(rawFilter)
	processedFilter = applyArmourTokens(processedFilter)

	remainingTokensRe := regexp.MustCompile(`#!.+!#`)

	processedFilter = remainingTokensRe.ReplaceAllString(processedFilter, "")

	doubleSpaceRe := regexp.MustCompile(` +`)

	processedFilter = doubleSpaceRe.ReplaceAllString(processedFilter, " ")

	return processedFilter
}

func applyColourTokens(rawFilter string) string {
	processedFilter := rawFilter

	for key, value := range colourTokens {
		if strings.IndexRune(key, '#') == 0 {
			continue
		}

		token := fmt.Sprintf("#!%s!#", key)
		processedFilter = strings.ReplaceAll(processedFilter, token, value)
	}

	return processedFilter
}

// e.g.
// #!strbody0!# = "Glorious Plate"
// #!strbody1!# = "Glorious Plate" "Gladiator Plate" "Astral Plate"
// #!strbody2!# = "Glorious Plate" "Gladiator Plate" "Astral Plate"
// #!str0!# = "Royal Plate" "Royal Burgonet" "Titan Greaves" "Titan Gauntlets"
// #!helm0!# = "Royal Burgonet" "Lion Pelt" "Hubris Circlet" etc.
func applyArmourTokens(rawFilter string) string {
	processedFilter := rawFilter

	tiers := 5

	// groups
	byAttribute := map[string][]string{
		"str":    make([]string, tiers),
		"dex":    make([]string, tiers),
		"int":    make([]string, tiers),
		"strdex": make([]string, tiers),
		"strint": make([]string, tiers),
		"dexint": make([]string, tiers),
	}

	bySlot := map[string][]string{
		"body":   make([]string, tiers),
		"helm":   make([]string, tiers),
		"boots":  make([]string, tiers),
		"gloves": make([]string, tiers),
		"square": make([]string, tiers),
	}

	for n := 0; n < tiers; n++ {
		for attribute, value := range armourTokens {
			for slot, value2 := range value {
				token := fmt.Sprintf("#!%s%s%d!#", attribute, slot, n)

				bases := ""
				for m := 0; m <= n; m++ {
					// each token should include all items from lower tiers
					if len(value2) >= (m+1) && len(value2[m]) > 0 {
						tempBases := strings.Join(value2[m], "\" \"")

						bases = fmt.Sprintf("%s \"%s\"", bases, tempBases)
					}
				}

				processedFilter = strings.ReplaceAll(processedFilter, token, bases)

				byAttribute[attribute][n] = fmt.Sprintf("%s%s", byAttribute[attribute][n], bases)

				bySlot[slot][n] = fmt.Sprintf("%s%s", bySlot[slot][n], bases)
				if slot != "body" {
					bySlot["square"][n] = fmt.Sprintf("%s%s", bySlot["square"][n], bases)
				}
			}
		}

		for attribute, bases := range byAttribute {
			token := fmt.Sprintf("#!%s%d!#", attribute, n)

			processedFilter = strings.ReplaceAll(processedFilter, token, bases[n])
		}

		for slot, bases := range bySlot {
			token := fmt.Sprintf("#!%s%d!#", slot, n)

			processedFilter = strings.ReplaceAll(processedFilter, token, bases[n])
		}
	}

	return processedFilter
}
