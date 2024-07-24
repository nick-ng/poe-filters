package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"poe-filters/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const MY_FILTERS_PATH string = "my-filters"
const BASE_FILTERS_PATH string = "base-filters"
const THIRD_PARTY_FILTERS_PATH string = "third-party-filters"
const OUTPUT_FILTERS_PATH string = "output-filters"

func main() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Println("\nCouldn't get home directory", err)
		os.Exit(1)
	}

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

		fmt.Printf("%d: %s... ", i+1, filterName)

		path := filepath.Join(MY_FILTERS_PATH, filterName)

		filter, errList := processFilter(path, false)

		outputFilterPath := filepath.Join(OUTPUT_FILTERS_PATH, filterName)
		gameFilterPath := filepath.Join(homeDir, "Documents", "My Games", "Path of Exile", filterName)
		if len(filter) == 0 {
			err := os.Remove(outputFilterPath)
			err2 := os.Remove(gameFilterPath)

			if err != nil && !os.IsNotExist(err) {
				fmt.Print(err)
			}

			if err2 != nil && !os.IsNotExist(err2) {
				fmt.Print(err2)
			}

			fmt.Println("skipped")

			continue
		}

		filterData := []byte(filter)

		err := os.WriteFile(outputFilterPath, filterData, 0666)

		if err != nil {
			fmt.Println("\nError writing filter to output-filters", filterName, err)
			continue
		}

		if filterName != "example.filter" {
			err = os.WriteFile(gameFilterPath, filterData, 0666)

			if err != nil {
				fmt.Println("\nError writing filter to PoE Directory", filterName, err)
				continue
			}
		}

		if len(errList) > 1 {
			fmt.Printf("done with %d errors\n", len(errList))
			continue
		} else if len(errList) == 1 {
			fmt.Println("done with 1 error")
			continue
		} else {
			fmt.Println("done")
		}
	}

	utils.PrintSoundStats()
}

// @todo(nick-ng): move some functions to separate files
func processFilter(filterPath string, isImported bool) (string, []error) {
	var errorList []error

	filterData, err := os.ReadFile(filterPath)

	if err != nil {
		return "", append(errorList, err)
	}

	rawLines := strings.Split(string(filterData), "\n")

	var processedLines []string
	if !isImported {
		processedLines = append(processedLines, `Show
	#!DefaultBackground!#
	Continue`)
	}

	var currentCommand string
	var importAfterTimestamp int64
	now := time.Now().Unix()
	options := make(map[string][]string)

	for _, rawLine := range rawLines {
		trimmedLine := strings.TrimSpace(rawLine)
		switch currentCommand {
		case "import":
			{
				subCommandArguments := getCommands(trimmedLine)
				switch subCommandArguments[0] {
				case "del":
					fallthrough
				case "delete":
					{
						regexpString := strings.Join(subCommandArguments[1:], " ")
						options["delete"] = append(options["delete"], regexpString)
					}
				case "maxarea":
					{
						options["maxarea"] = []string{subCommandArguments[1]}
					}
				case "import":
					fallthrough
				case "noop":
					{
						if !strings.HasPrefix(trimmedLine, "#") || subCommandArguments[0] == "import" {
							// it's a non-comment line so import the filter here then write the line
							_, present := options["file"]
							if present {
								_, fullPath, err := importBaseFilter(options["file"][0])

								if err != nil {
									errorList = append(errorList, err)
									processedLines = append(processedLines, fmt.Sprintf("#? error: couldn't import %s", options["file"]))
									processedLines = append(processedLines, fmt.Sprintf("#?  %s\n", err))
								}

								tempFilter, errs := processFilter(fullPath, true)

								if len(errs) != 0 {
									errorList = append(errorList, err)
									processedLines = append(processedLines, fmt.Sprintf("#? error: couldn't import %s", options["file"]))
									processedLines = append(processedLines, fmt.Sprintf("#?  %s\n", err))
								} else if now < int64(importAfterTimestamp) {
									processedLines = append(processedLines, fmt.Sprintf("#?  skipped because generated at %d", now))
								} else {
									for _, deleteRegexpString := range options["delete"] {
										deleteRegexp, err := regexp.Compile(deleteRegexpString)

										if err != nil {
											errorString := fmt.Sprintf("couldn't compile regexp: %s", deleteRegexpString)
											errorList = append(errorList, errors.New(errorString))
											processedLines = append(processedLines, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter = deleteRegexp.ReplaceAllString(tempFilter, "")
										}
									}

									if options["maxarea"] != nil && len(options["maxarea"]) > 0 {
										maxarea, err := strconv.ParseInt(options["maxarea"][0], 10, 0)

										if err != nil {
											errorString := fmt.Sprintf("maxarea argument not an integer: %s", options["maxarea"][0])
											errorList = append(errorList, errors.New(errorString))
											processedLines = append(processedLines, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter2, err := utils.LimitMaxAreaLevel(tempFilter, int(maxarea))

											if err != nil {
												errorString := fmt.Sprintf("couldn't lower filter's level: %s", err)
												errorList = append(errorList, errors.New(errorString))
												processedLines = append(processedLines, fmt.Sprintf("#? error: %s", errorString))
											} else {
												tempFilter = tempFilter2
											}
										}
									}

									processedLines = append(processedLines, tempFilter)
									processedLines = append(processedLines, fmt.Sprintf("#? end of %s\n", options["file"]))
								}

							} else {
								processedLines = append(processedLines, "#? error: No file specified.\n")
							}

							currentCommand = ""
							clear(options)
						}

						if subCommandArguments[0] == "import" {
							currentCommand = "import"
							options["file"] = []string{subCommandArguments[1]}

							tempLine := utils.CleanUpCommand(rawLine)
							processedLines = append(processedLines, tempLine)
						} else {
							processedLines = append(processedLines, rawLine)
						}
					}
				default:
					{
						processedLines = append(processedLines, trimmedLine)

						warning := fmt.Sprintf("#? warning: Unknown sub-command %s", subCommandArguments[0])
						processedLines = append(processedLines, warning)
					}
				}
			}
		default:
			{
				commandArguments := getCommands(trimmedLine)
				switch commandArguments[0] {
				case "import":
					{
						currentCommand = "import"
						options["file"] = []string{commandArguments[1]}
						if len(commandArguments) > 2 {
							importAfterTimestamp, err = strconv.ParseInt(commandArguments[2], 10, 0)

							if err != nil {
								errorString := fmt.Sprintf("couldn't parse import after timestamp: %s", commandArguments[2])
								errorList = append(errorList, errors.New(errorString))
							}
						}

						tempLine := utils.CleanUpCommand(rawLine)
						processedLines = append(processedLines, tempLine)
					}
				case "del":
					fallthrough
				case "delete":
					{
						processedLines = append(processedLines, rawLine)
						warning := fmt.Sprintf("#? warning: %s only allowed during an import", commandArguments[0])
						processedLines = append(processedLines, warning)
					}
				case "skip":
					{
						return "", []error{}
					}
				case "noop":
					fallthrough
				default:
					{
						processedLines = append(processedLines, rawLine)
						// Handle in the next loop so we don't get double warnings
						// noop is also here
					}
				}
			}
		}
	}

	rawLines = strings.Split(strings.Join(processedLines, "\n"), "\n")
	var processedLines2 []string
	for _, rawLine := range rawLines {
		trimmedLine := strings.TrimSpace(rawLine)

		if strings.HasPrefix(trimmedLine, "#! ") {
			commandArguments := getCommands(trimmedLine)
			switch commandArguments[0] {
			case "import":
				fallthrough
			case "del":
				fallthrough
			case "delete":
				fallthrough
			case "noop":
				{
					processedLines2 = append(processedLines2, rawLine)
					// noop
				}
			case "links":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					processedLines2 = append(processedLines2, tempLine)

					if len(commandArguments) < 2 {
						processedLines2 = append(processedLines2, "#? warning: need at least coloured links (SocketGroup) to work")
						continue
					}

					filterBlock, err := utils.GetSocketGroupFilter(commandArguments[1], commandArguments[2:]...)

					if err != nil {
						e := fmt.Sprintf("#? error: couldn't generate links\n#?  %s", err)
						processedLines2 = append(processedLines2, e)
					} else {
						processedLines2 = append(processedLines2, filterBlock)
					}

				}
			case "linksa":
				fallthrough
			case "linksarmor":
				fallthrough
			case "linksarmour":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					processedLines2 = append(processedLines2, tempLine)

					if len(commandArguments) < 2 {
						processedLines2 = append(processedLines2, "#? warning: need at least coloured links (SocketGroup) to work")
						continue
					}

					newLine := utils.GetArmourSocketGroupFilter(commandArguments[1], commandArguments[2:]...)
					processedLines2 = append(processedLines2, newLine)
				}
			case "tts":
				{
					newLine := utils.MakeTts(trimmedLine)
					processedLines2 = append(processedLines2, newLine)
				}
			default:
				{
					warning := fmt.Sprintf("#? warning: Unknown command %s", commandArguments[0])
					processedLines2 = append(processedLines2, warning)
				}
			}
		} else if regexp.MustCompile(
			`^[^#]*CustomAlertSound(Optional)? +"[^:"]+"`,
		).MatchString(trimmedLine) {
			processedLines2 = append(processedLines2, utils.FixSoundPath(trimmedLine))
		} else {
			processedLines2 = append(processedLines2, rawLine)
		}
	}

	filter := strings.Join(processedLines2, "\n")
	filter = utils.ApplyAllTokens(filter)
	filter = utils.CleanUpFilter(filter)

	return filter, errorList
}

func importBaseFilter(filterName string) (string, string, error) {
	path := filepath.Join(BASE_FILTERS_PATH, filterName)

	filterData, err := os.ReadFile(path)

	if err != nil {
		path = filepath.Join(THIRD_PARTY_FILTERS_PATH, filterName)
		filterData, err = os.ReadFile(path)
	}

	return string(filterData), path, err
}

func getCommands(rawCommand string) []string {
	commands := strings.Split(rawCommand, " ")
	if commands[0] != "#!" {
		return []string{"noop"}
	}

	return commands[1:]
}
