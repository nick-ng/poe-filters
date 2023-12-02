package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var blockPrefixes = []string{
	"GemQualityType",
}

var commandCommentRe = regexp.MustCompile(`(^[^#]*)#!`)

func CleanUpFilter(filter string) string {
	groups := []string{}
	currentGroup := []string{}

	rawLines := strings.Split(filter, "\n")

	for _, rawLine := range rawLines {
		trimmedLine := strings.TrimSpace(rawLine)
		lowercaseLine := strings.ToLower(trimmedLine)

		if strings.HasPrefix(lowercaseLine, "show") {
			if isGroupGood(currentGroup) {
				groups = append(groups, strings.Join(currentGroup, "\n"))
			}

			currentGroup = []string{"Show"}
		} else if strings.HasPrefix(lowercaseLine, "hide") {
			if isGroupGood(currentGroup) {
				groups = append(groups, strings.Join(currentGroup, "\n"))
			}

			currentGroup = []string{"Hide"}
		} else if strings.HasPrefix(trimmedLine, "#") {
			currentGroup = append(currentGroup, rawLine)
		} else if len(trimmedLine) > 0 {
			currentGroup = append(currentGroup, fmt.Sprintf("\t%s", trimmedLine))
		}
	}

	return strings.Join(groups, "\n")
}

func isGroupGood(group []string) bool {
	for _, line := range group {
		trimmedLine := strings.TrimSpace(line)

		for _, blockPrefix := range blockPrefixes {
			if strings.HasPrefix(trimmedLine, blockPrefix) {
				return false
			}
		}
	}

	return true
}

func CleanUpCommand(originalLine string) string {
	trimmedLine := strings.TrimSpace(originalLine)
	if strings.HasPrefix(trimmedLine, "#!") {
		tempLine := commandCommentRe.ReplaceAllString(originalLine, "$1#?")

		return tempLine
	}

	return originalLine
}
