package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"poe-filters/utils"
	"regexp"
	"strings"
)

const MY_FILTERS_PATH string = "my-filters"
const BASE_FILTERS_PATH string = "base-filters"
const THIRD_PARTY_FILTERS_PATH string = "third-party-filters"
const OUTPUT_FILTERS_PATH string = "output-filters"

func main() {
	utils.MkDirIfNotExist(MY_FILTERS_PATH)
	utils.MkDirIfNotExist(OUTPUT_FILTERS_PATH)
	utils.MkDirIfNotExist(BASE_FILTERS_PATH)
	utils.MkDirIfNotExist(THIRD_PARTY_FILTERS_PATH)

	path1 := filepath.Join(MY_FILTERS_PATH)

	dat1, err := os.ReadDir(path1)

	if err != nil {
		fmt.Println("Error reading file", err)
	}

	fmt.Printf("Found %d filters\n", len(dat1))

	for i, dir := range dat1 {
		if err != nil {
			continue
		}

		filterName := dir.Name()

		fmt.Printf("%d: %s...", i+1, filterName)

		path := filepath.Join(MY_FILTERS_PATH, filterName)

		filter, errList := processFilter(path)

		filterData := []byte(filter)

		path = filepath.Join(OUTPUT_FILTERS_PATH, filterName)
		err := os.WriteFile(path, filterData, 0666)

		if err != nil {
			fmt.Println("\nError writing filter", filterName, err)
			continue
		}

		if len(errList) > 0 {
			fmt.Printf(" done with %d errors\n", len(errList))
			continue
		} else {
			fmt.Println(" done")
		}
	}
}

// @todo(nick-ng): move some functions to separate files
func processFilter(filterPath string) (string, []error) {
	var errorList []error

	filterData, err := os.ReadFile(filterPath)

	if err != nil {
		return "", append(errorList, err)
	}

	rawLines := strings.Split(string(filterData), "\n")

	processedLines := []string{}

	var currentCommand string
	options := make(map[string][]string)

	for _, rawLine := range rawLines {
		trimmedLine := strings.TrimSpace(rawLine)
		switch currentCommand {
		case "import":
			{
				if strings.HasPrefix(trimmedLine, "#! ") {
					processedLines = append(processedLines, trimmedLine)
					subCommandArguments := getCommands(trimmedLine)
					switch subCommandArguments[1] {
					case "del":
						fallthrough
					case "delete":
						{
							regexpString := strings.Join(subCommandArguments[2:], " ")
							options["delete"] = append(options["delete"], regexpString)
						}
					default:
						{
							processedLines = append(processedLines, fmt.Sprintf("# Warning: Unknown sub-command %s", subCommandArguments[1]))
						}
					}
				} else if strings.HasPrefix(trimmedLine, "#") {
					processedLines = append(processedLines, trimmedLine)
				} else {
					// it's a non-comment line so import the filter here then write the line
					_, present := options["file"]
					if present {
						tempFilter, err := importBaseFilter(options["file"][0])
						if err != nil {
							errorList = append(errorList, err)
							processedLines = append(processedLines, fmt.Sprintf("# Error: couldn't import %s", options["file"]))
							processedLines = append(processedLines, fmt.Sprintf("#  %s\n", err))
						} else {
							for _, deleteRegexpString := range options["delete"] {
								deleteRegexp, err := regexp.Compile(deleteRegexpString)

								if err != nil {
									errorString := fmt.Sprintf("couldn't compile regexp: %s", deleteRegexpString)
									errorList = append(errorList, errors.New(errorString))
									processedLines = append(processedLines, fmt.Sprintf("# Error: %s", errorString))
								} else {
									tempFilter = deleteRegexp.ReplaceAllString(tempFilter, "")
								}
							}

							processedLines = append(processedLines, tempFilter)
							processedLines = append(processedLines, fmt.Sprintf("# End of %s\n", options["file"]))
						}

					} else {
						processedLines = append(processedLines, "# Error: No file specified.\n")
					}
					processedLines = append(processedLines, trimmedLine)

					currentCommand = ""
					clear(options)
				}
				break
			}
		default:
			{
				processedLines = append(processedLines, trimmedLine)
				if strings.HasPrefix(trimmedLine, "#! ") {
					// #! <command> <argument1> <argument2> <etc>
					// 0  1         2           3
					commandArguments := getCommands(trimmedLine)
					switch commandArguments[1] {
					case "import":
						{
							currentCommand = "import"
							options["file"] = []string{commandArguments[2]}
							break
						}
					case "links":
						fallthrough
					case "linksa":
						{
							processedLines = append(processedLines, utils.GetArmourSocketGroupFilter(commandArguments[2], commandArguments[3:]...))
						}
					default:
						{
							processedLines = append(processedLines, fmt.Sprintf("# Warning: Unknown command %s", commandArguments[1]))
						}
					}
				}
			}
		}
	}

	joinedFilter := strings.Join(processedLines, "\n")
	return joinedFilter, errorList
	// @todo(nick-ng): replace tokens and remove all unknown tokens
	// return regexp.MustCompile(`#!.+!#`).ReplaceAllString(joinedFilter, ""), errorList
}

func importBaseFilter(filterName string) (string, error) {
	path := filepath.Join(BASE_FILTERS_PATH, filterName)

	filterData, err := os.ReadFile(path)

	if err != nil {
		path = filepath.Join(THIRD_PARTY_FILTERS_PATH, filterName)
		filterData, err = os.ReadFile(path)
	}

	return string(filterData), err
}

func getCommands(rawCommand string) []string {
	return strings.Split(rawCommand, " ")
}
