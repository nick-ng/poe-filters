package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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

	maxTiers := 10

	// groups
	byAttribute := map[string][]string{
		"str":         make([]string, maxTiers),
		"dex":         make([]string, maxTiers),
		"int":         make([]string, maxTiers),
		"strdex":      make([]string, maxTiers),
		"strint":      make([]string, maxTiers),
		"dexint":      make([]string, maxTiers),
		"suppression": make([]string, maxTiers),
	}

	bySlot := map[string][]string{
		"body":   make([]string, maxTiers),
		"helm":   make([]string, maxTiers),
		"boots":  make([]string, maxTiers),
		"gloves": make([]string, maxTiers),
		"square": make([]string, maxTiers),
	}

	byBoth := map[string][]string{
		"strbody":           make([]string, maxTiers),
		"suppressionbody":   make([]string, maxTiers),
		"strsquare":         make([]string, maxTiers),
		"suppressionsquare": make([]string, maxTiers),
	}

	for tierLimit := 0; tierLimit < maxTiers; tierLimit++ {
		for attribute, armourTokensForAttribute := range armourTokens {
			for slot, armourBases := range armourTokensForAttribute {
				token := fmt.Sprintf("#!%s%s%d!#", attribute, slot, tierLimit)

				bases := []string{}
				for tier := 0; tier <= tierLimit; tier++ {
					// each token should include all items from lower tiers
					if len(armourBases) >= (tier+1) && len(armourBases[tier]) > 0 {
						bases = append(bases, armourBases[tier]...)
					}
				}

				slices.Sort(bases)

				basesString := fmt.Sprintf("\"%s\" ", strings.Join(bases, "\" \""))

				processedFilter = strings.ReplaceAll(processedFilter, token, basesString)

				// @todo(nick-ng): convert slice to string just before replacement so you can sort all bases
				byAttribute[attribute][tierLimit] = fmt.Sprintf("%s%s", byAttribute[attribute][tierLimit], basesString)

				bySlot[slot][tierLimit] = fmt.Sprintf("%s%s", bySlot[slot][tierLimit], basesString)
				if slot != "body" {
					bySlot["square"][tierLimit] = fmt.Sprintf("%s%s", bySlot["square"][tierLimit], basesString)
				}

				if strings.Contains(attribute, "dex") {
					byAttribute["suppression"][tierLimit] = fmt.Sprintf("%s%s", byAttribute["suppression"][tierLimit], basesString)

					if slot != "body" {
						byBoth["suppressionsquare"][tierLimit] = fmt.Sprintf("%s%s", byBoth["suppressionsquare"][tierLimit], basesString)
					} else {
						byBoth["suppressionbody"][tierLimit] = fmt.Sprintf("%s%s", byBoth["suppressionbody"][tierLimit], basesString)
					}
				}

				if attribute == "str" {
					if slot != "body" {
						byBoth["strsquare"][tierLimit] = fmt.Sprintf("%s%s", byBoth["strsquare"][tierLimit], basesString)
					} else {
						byBoth["strbody"][tierLimit] = fmt.Sprintf("%s%s", byBoth["strbody"][tierLimit], basesString)
					}
				}
			}
		}

		// @todo(nick-ng): remove duplicates - i.e. two-toned boots
		for attribute, basesString := range byAttribute {
			token := fmt.Sprintf("#!%s%d!#", attribute, tierLimit)

			processedFilter = strings.ReplaceAll(processedFilter, token, basesString[tierLimit])
		}

		for slot, basesString := range bySlot {
			token := fmt.Sprintf("#!%s%d!#", slot, tierLimit)

			processedFilter = strings.ReplaceAll(processedFilter, token, basesString[tierLimit])
		}

		for both, basesString := range byBoth {
			token := fmt.Sprintf("#!%s%d!#", both, tierLimit)

			processedFilter = strings.ReplaceAll(processedFilter, token, basesString[tierLimit])
		}
	}

	return processedFilter
}
