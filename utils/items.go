package utils

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

type CurrencyBreakpoint struct {
	Comment    string
	Styles     []string
	ChaosValue float64
	HasMapIcon bool
}

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

// var wisdomStyle = []string{
// 	"SetFontSize 35",
// 	"SetTextColor 210 178 135 220",
// 	"SetBackgroundColor 0 0 0 120",
// 	"SetBorderColor 130 130 255 255",
// }
// var portalStyle = []string{
// 	"SetFontSize 35",
// 	"SetTextColor 50 240 240 220",
// 	"SetBackgroundColor 0 0 0 120",
// 	"SetBorderColor 130 130 255 255",
// }

// @todo(nick-ng): make wisdom and portal scrolls use the correct style
func GetStackableCurrencyFilter(currencyPrices CurrencyPrices, minAreaLevel int, minChaos float64) string {
	needMapIcon := map[string]bool{
		"Orb of Fusing":                true,
		"Orb of Regret":                true,
		"Eldritch Chaos Orb":           true,
		"Eldritch Exalted Orb":         true,
		"Eldritch Orb of Annulment":    true,
		"Lesser Eldritch Ember":        true,
		"Greater Eldritch Ember":       true,
		"Grand Eldritch Ember":         true,
		"Exceptional Eldritch Ember":   true,
		"Lesser Eldritch Ichor":        true,
		"Greater Eldritch Ichor":       true,
		"Grand Eldritch Ichor":         true,
		"Exceptional Eldritch Ichor":   true,
		"Foulborn Exalted Orb":         true,
		"Foulborn Regal Orb":           true,
		"Foulborn Orb of Augmentation": true,
	}
	needShow := map[string]bool{
		"Jeweller's Orb":  true,
		"Orb of Chance":   true,
		"Orb of Alchemy":  true,
		"Orb of Binding":  true,
		"Orb of Scouring": true,
	}
	breakPoints := []CurrencyBreakpoint{
		{
			Comment:    "1+ divine",
			ChaosValue: 1 / currencyPrices.DivinePerChaos,
			Styles: []string{
				"SetFontSize 45",
				"SetTextColor 255 0 0 255",
				"SetBackgroundColor 255 255 255",
				"SetBorderColor 130 130 255 255",
				"MinimapIcon 0 White Diamond",
				"PlayAlertSound 6 300",
				"PlayEffect Red",
			},
			HasMapIcon: true,
		},
		{
			Comment:    "0.5+ divine",
			ChaosValue: 0.5 / currencyPrices.DivinePerChaos,
			Styles: []string{
				"SetFontSize 43",
				"SetTextColor 255 0 0 255",
				"SetBackgroundColor 0 0 0 120",
				"SetBorderColor 130 130 255 255",
				"MinimapIcon 0 Pink Circle",
				"CustomAlertSound \"sounds/thps-special-trick-1.mp3\" 300",
				"PlayEffect Green",
			},
			HasMapIcon: true,
		},
		{
			Comment:    "1+ chaos",
			ChaosValue: 1,
			Styles: []string{
				"SetFontSize 41",
				"SetTextColor 255 150 0 255",
				"SetBackgroundColor 0 0 0 120",
				"SetBorderColor 130 130 255 255",
				"MinimapIcon 1 Orange Circle",
				"PlayAlertSound 11 250",
				"PlayEffect Green",
			},
			HasMapIcon: true,
		},
		{
			Comment:    "0.5+ chaos",
			ChaosValue: 0.5,
			Styles: []string{
				"SetFontSize 39",
				"SetTextColor 255 255 0 255",
				"SetBackgroundColor 0 0 0 120",
				"SetBorderColor 130 130 255 255",
				"MinimapIcon 1 Yellow Circle",
				"PlayAlertSound 9 250",
			},
			HasMapIcon: true,
		},
		{
			Comment:    "0.1+ chaos",
			ChaosValue: 0.1,
			Styles: []string{
				"SetFontSize 37",
				"SetTextColor 0 255 0 255",
				"SetBackgroundColor 0 0 0 120",
				"SetBorderColor 130 130 255 255",
			},
			HasMapIcon: false,
		},
	}

	filterString := fmt.Sprintf(`# Auto Currency Filter
Hide
	AreaLevel >= %d
	Class == "Stackable Currency"
	BaseType == "Scroll of Wisdom" "Portal Scroll"
`, minAreaLevel)
	for i, breakPoint := range breakPoints {
		if breakPoint.ChaosValue < minChaos {
			continue
		}

		if !breakPoint.HasMapIcon {
			mapIconBaseTypes := []string{}
			for baseType, b := range needMapIcon {
				if b {
					mapIconBaseTypes = append(mapIconBaseTypes, baseType)
				}
			}

			if len(mapIconBaseTypes) > 0 {
				baseTypes := strings.Join(mapIconBaseTypes, "\" \"")
				temp := fmt.Sprintf(`Show
	BaseTypes == "%s"
	MinimapIcon 2 Green Circle
	Continue
`, baseTypes)

				filterString = fmt.Sprintf("%s\n%s\n", filterString, temp)
			}
		}

		baseTypesInBreakPoint := []string{}
		for _, curr := range currencyPrices.Prices {
			if i > 0 && curr.ChaosValue >= breakPoints[i-1].ChaosValue {
				continue
			}

			if curr.ChaosValue >= breakPoint.ChaosValue {
				baseTypesInBreakPoint = append(baseTypesInBreakPoint, curr.BaseType)
				// since the currency is shown in this group, we don't need to show anymore
				needShow[curr.BaseType] = false
				if breakPoint.HasMapIcon {
					needMapIcon[curr.BaseType] = false
				}
			}
		}

		if len(baseTypesInBreakPoint) > 0 {
			baseTypes := strings.Join(baseTypesInBreakPoint, "\" \"")
			styles := strings.Join(breakPoint.Styles, "\n\t")
			thisFilterGroup := fmt.Sprintf(`# %s
Show
	AreaLevel >= %d
	Class == "Stackable Currency"
	BaseTypes == "%s"
	%s
`, breakPoint.Comment, minAreaLevel, baseTypes, styles)

			filterString = fmt.Sprintf("%s\n%s\n", filterString, thisFilterGroup)
		}
	}

	mustShowBaseTypes := []string{}
	for baseType, b := range needShow {
		if b {
			mustShowBaseTypes = append(mustShowBaseTypes, baseType)
		}
	}
	for baseType, b := range needMapIcon {
		if b {
			mustShowBaseTypes = append(mustShowBaseTypes, baseType)
		}
	}

	baseTypes := strings.Join(mustShowBaseTypes, "\" \"")
	mustShowFilter := fmt.Sprintf(`# must show
Show
	AreaLevel >= %d
	Class == "Stackable Currency"
	BaseTypes == "%s"
	SetFontSize 35,
	SetTextColor 0 255 150 255
	SetBackgroundColor 0 0 0 120
	SetBorderColor 130 130 255 255
`, minAreaLevel, baseTypes)

	filterString = fmt.Sprintf("%s\n%s\n", filterString, mustShowFilter)

	// @todo(nick-ng): omit items that were in earlier groups
	minStackSize := map[int][]string{}
	chaosThreshold := math.Max(0.1, minChaos)
	for _, curr := range currencyPrices.Prices {
		if curr.ChaosValue < chaosThreshold {
			requiredStackSize := int(math.Ceil(chaosThreshold / curr.ChaosValue))

			_, ok := minStackSize[requiredStackSize]
			if !ok {
				minStackSize[requiredStackSize] = []string{}
			}
			minStackSize[requiredStackSize] = append(minStackSize[requiredStackSize], curr.BaseType)
		}
	}

	for stackSize, bb := range minStackSize {
		baseTypes := strings.Join(bb, "\" \"")
		stackSizeFilter := fmt.Sprintf(`# stack size
Show
	AreaLevel >= %d
	StackSize >= %d
	Class == "Stackable Currency"
	BaseTypes == "%s"
	SetFontSize 35,
	SetTextColor 0 255 150 255
	SetBackgroundColor 0 0 0 120
	SetBorderColor 130 130 255 255
`, minAreaLevel, stackSize, baseTypes)

		filterString = fmt.Sprintf("%s\n%s\n", filterString, stackSizeFilter)
	}

	return filterString
}
