package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DivinationCard struct {
	Name      string
	StackSize int
	Tier      int
}

type Flag struct {
	Name  string
	Value string
}

func MkDirIfNotExist(dirPath string) {
	err := os.Mkdir(dirPath, 0755)

	if err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Println(err)
		os.Exit(1)
	}
}

func MakeDivinationCardsFilter() {
	divinationCardsTxtPath := filepath.Join("notes", "divination-cards.txt")
	divinationCardsFilterPath := filepath.Join("base-filters", "full-stack-divination-cards.filter")

	rawDivinationCards, err := os.ReadFile(divinationCardsTxtPath)

	if err != nil {
		fmt.Println("no divination cards file")
		return
	}

	var divinationCards []DivinationCard

	var previousLine string

	rawLines := strings.Split(string(rawDivinationCards), "\n")
	maxStackSize := 0

	for _, divinationCardLine := range rawLines {
		if strings.HasPrefix(divinationCardLine, "Stack Size:") {
			temp := strings.Split(divinationCardLine, "/")

			if len(temp) < 2 {
				previousLine = divinationCardLine
				continue
			}

			stackSizeString := strings.Trim(temp[1], " ")
			stackSize, err := strconv.Atoi(stackSizeString)

			if err != nil {
				fmt.Println("error when parsing divination card line", divinationCardLine, previousLine)
				os.Exit(1)
			}

			divinationCards = append(divinationCards, DivinationCard{
				Name:      previousLine,
				StackSize: stackSize,
				Tier:      5,
			})

			if stackSize > maxStackSize {
				maxStackSize = stackSize
			}

			previousLine = ""
		} else {
			previousLine = divinationCardLine
		}
	}

	var filterGroups []string

	for i := 1; i <= maxStackSize; i++ {
		var divinationCardNames []string

		for _, divinationCard := range divinationCards {
			if divinationCard.StackSize == i {
				divinationCardNames = append(divinationCardNames, divinationCard.Name)
			}
		}

		if len(divinationCardNames) == 0 {
			continue
		}

		var filterGroup []string

		baseTypesLine := fmt.Sprintf("\tBaseType == \"%s\"", strings.Join(divinationCardNames, "\" \""))
		stackSizeLine := fmt.Sprintf("\tStackSize >= %d", i)

		filterGroup = append(filterGroup, "Show", baseTypesLine, stackSizeLine, `	Class "Divination"
	SetFontSize 45
	#!BrightBackground!# 255
	#!CurrencyBorder!# 230
	MinimapIcon 0 Pink UpsideDownHouse
	PlayEffect Green
	CustomAlertSound "sounds/intel-thanks-steve.mp3"`)

		filterGroups = append(filterGroups, strings.Join(filterGroup, "\n"))
	}

	filter := strings.Join(filterGroups, "\n\n")

	os.WriteFile(divinationCardsFilterPath, []byte(filter), 0666)

	fmt.Printf("%d divination cards found\n", len(divinationCards))
}

// given a string, returns a map with all flags. includes the name if present. otherwise, the name will be nil.
// the flag's name, if present, must not contain an equals sign
func ParseFlags(rawCommand string) []Flag {
	fullCommand := strings.TrimSpace(rawCommand)
	var flags []Flag

	var word []rune
	inQuote := false

	for _, runeValue := range fullCommand {
		if !inQuote && runeValue == ' ' {
			flagString := string(word)
			word = nil

			if flagString == "#!" {
				continue
			}

			hasEquals := strings.Contains(flagString, "=")
			if !hasEquals {
				flags = append(flags, Flag{
					Value: flagString,
				})

			} else {
				flagParts := strings.Split(flagString, "=")
				flagName := flagParts[0]
				flagValue := strings.Join(flagParts[1:], "=")
				flagValue = strings.TrimSpace(flagValue)
				flags = append(flags, Flag{
					Name:  flagName,
					Value: flagValue,
				})
			}

			continue
		}

		// @todo(nick-ng): this has weird behaviour if you "open" and "close" quotes multiple times
		if !inQuote && runeValue == '"' {
			inQuote = true
			continue
		} else if inQuote && runeValue == '"' {
			inQuote = false
			continue
		}

		word = append(word, runeValue)
	}

	// handle the last flag
	if word != nil {
		flagString := string(word)

		if flagString == "#!" {
			return flags
		}

		hasEquals := strings.Contains(flagString, "=")
		if !hasEquals {
			flags = append(flags, Flag{
				Value: flagString,
			})

		} else {
			flagParts := strings.Split(flagString, "=")
			flagName := flagParts[0]
			flagValue := strings.Join(flagParts[1:], "=")
			flagValue = strings.TrimSpace(flagValue)
			flags = append(flags, Flag{
				Name:  flagName,
				Value: flagValue,
			})
		}
	}

	return flags
}
