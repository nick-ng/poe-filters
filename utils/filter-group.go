package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type FilterGroup struct {
	// if there are any comments before the first show, the type will be comments
	Type    string
	MaxArea int // this becomes AreaLevel >= MaxArea
	MinArea int // this becomes AreaLevel <= MinArea
	// lines don't include show/hide but does include all comments
	Lines         []string
	OriginalLines []string
}

func FindFilterGroups(filter string) ([]FilterGroup, error) {
	var filterGroups []FilterGroup
	currentFilterGroup := FilterGroup{Type: "comments", MinArea: 0, MaxArea: 999}

	rawLines := strings.Split(filter, "\n")

	for _, rawLine := range rawLines {
		currentFilterGroup.OriginalLines = append(currentFilterGroup.OriginalLines, rawLine)
		trimmedLine := strings.TrimSpace(rawLine)

		// if it's show or hide, put any existing stuff into the filter groups
		// then clear the group
		if strings.HasPrefix(trimmedLine, "Show") || strings.HasPrefix(trimmedLine, "Hide") {
			if len(currentFilterGroup.Lines) > 0 {
				filterGroups = append(filterGroups, currentFilterGroup)
			}

			currentFilterGroup = FilterGroup{Type: trimmedLine, MinArea: 0, MaxArea: 999}
			continue
		}

		if strings.HasPrefix(trimmedLine, "AreaLevel") {
			greaterThan := regexp.MustCompile(`>`).MatchString(trimmedLine)
			lessThan := regexp.MustCompile(`<`).MatchString(trimmedLine)
			equalTo := regexp.MustCompile(`=`).MatchString(trimmedLine)
			levelString := regexp.MustCompile(`\d+`).FindString(trimmedLine)

			if len(levelString) == 0 {
				fmt.Println("couldn't find level for AreaLevel")
				return nil, errors.New("couldn't find level for AreaLevel")
			}

			level, err := strconv.ParseInt(levelString, 10, 0)

			if err != nil {
				return nil, err
			}

			if greaterThan {
				if equalTo {
					currentFilterGroup.MinArea = int(level)
				} else {
					currentFilterGroup.MinArea = int(level + 1)
				}
			} else if lessThan {
				if equalTo {
					currentFilterGroup.MaxArea = int(level)
				} else {
					currentFilterGroup.MaxArea = int(level - 1)
				}
			} else if equalTo {
				currentFilterGroup.MinArea = int(level)
				currentFilterGroup.MaxArea = int(level)
			}
			continue
		}

		// otherwise add it to lines
		currentFilterGroup.Lines = append(currentFilterGroup.Lines, rawLine)
	}

	filterGroups = append(filterGroups, currentFilterGroup)

	return filterGroups, nil
}

func LimitMaxAreaLevel(filter string, maxAreaLevel int) (string, error) {
	filterGroups, err := FindFilterGroups(filter)

	if err != nil {
		return "", err
	}

	var newFilterLines []string

	for _, filterGroup := range filterGroups {
		if filterGroup.Type == "comments" {
			newFilterLines = append(newFilterLines, filterGroup.Lines...)
			continue
		}

		newFilterLines = append(newFilterLines, filterGroup.Type)

		if filterGroup.MinArea > 1 {
			newFilterLines = append(newFilterLines, fmt.Sprintf("AreaLevel >= %d", filterGroup.MinArea))
		}

		if filterGroup.MaxArea > maxAreaLevel {
			newFilterLines = append(newFilterLines, fmt.Sprintf("AreaLevel <= %d", maxAreaLevel))
		} else {
			newFilterLines = append(newFilterLines, fmt.Sprintf("AreaLevel <= %d", filterGroup.MaxArea))
		}

		newFilterLines = append(newFilterLines, filterGroup.Lines...)
	}

	newFilterLines = append(newFilterLines, "") // add final newline

	return strings.Join(newFilterLines, "\n"), nil
}
