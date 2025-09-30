package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"poe-filters/utils"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ProcessedFilterFlags struct {
	Game string // poe1, poe2, or unset
}

const MY_FILTERS_PATH string = "my-filters"
const MY_POE2_FILTERS_PATH string = "my-poe2-filters"
const BASE_FILTERS_PATH string = "base-filters"
const BUILD_FILTERS_PATH string = "build-filters"
const THIRD_PARTY_FILTERS_PATH string = "third-party-filters"
const OUTPUT_FILTERS_PATH string = "output-filters"
const CACHE_PATH string = "cache"

func main() {
	utils.MakeDivinationCardsFilterPoeNinja()

	utils.MkDirIfNotExist(MY_FILTERS_PATH)
	utils.MkDirIfNotExist(MY_POE2_FILTERS_PATH)
	utils.MkDirIfNotExist(OUTPUT_FILTERS_PATH)
	utils.MkDirIfNotExist(BASE_FILTERS_PATH)
	utils.MkDirIfNotExist(THIRD_PARTY_FILTERS_PATH)
	utils.MkDirIfNotExist(CACHE_PATH)

	if runtime.GOOS != "windows" {
		poe1TtsDir := utils.GetPoe1Path("tts/")
		utils.MkDirIfNotExist(poe1TtsDir)

		poe2TtsDir := utils.GetPoe2Path("tts/")
		utils.MkDirIfNotExist(poe2TtsDir)
	}

	path1 := filepath.Join(MY_FILTERS_PATH)
	dat1, err := os.ReadDir(path1)
	if err != nil {
		fmt.Println("Error reading file", err)
	}

	// parse poe2 filters first
	var poe1Filters []string
	var poe2Filters []string
	for _, dir := range dat1 {
		filterName := dir.Name()
		filterPath := filepath.Join(MY_FILTERS_PATH, filterName)
		if strings.Contains(filterName, "poe2") {
			poe2Filters = append(poe2Filters, filterPath)
		} else {
			poe1Filters = append(poe1Filters, filterPath)
		}
	}

	path2 := filepath.Join(MY_POE2_FILTERS_PATH)
	dat2, err := os.ReadDir(path2)
	if err != nil {
		fmt.Println("Error reading file", err)
	}

	for _, dir := range dat2 {
		filterName := dir.Name()
		filterPath := filepath.Join(MY_POE2_FILTERS_PATH, filterName)
		poe2Filters = append(poe2Filters, filterPath)
	}

	allFilters := append(poe2Filters, poe1Filters...)

	fmt.Printf("Found %d filters\n", len(allFilters))

	for i, filterPath := range allFilters {
		temp := strings.Split(filterPath, string(filepath.Separator))
		filterName := temp[len(temp)-1]
		// if err != nil {
		// 	continue
		// }

		// filterName := dir.Name()

		fmt.Printf("%d: %s... ", i+1, filterName)

		filter, flags, errList := processFilter(filterPath, false)

		outputFilterPath := filepath.Join(OUTPUT_FILTERS_PATH, filterName)
		gameFilterPath := utils.GetPoe1Path(filterName)

		if flags.Game == "poe2" {
			fmt.Print("PoE 2 ")
			poe2FilterName := fmt.Sprintf("poe2-%s", filterName)
			outputFilterPath = filepath.Join(OUTPUT_FILTERS_PATH, poe2FilterName)
			gameFilterPath = utils.GetPoe2Path(filterName)
		}

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

		if filterName != "example.filter" && filterName != "example2.filter" {
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

	if runtime.GOOS != "windows" {
		temp, err := filepath.Abs(filepath.Join("sounds"))
		if err != nil {
			fmt.Println("error getting sounds directory", err)
			os.Exit(1)
		}
		soundsPath := fmt.Sprintf("%s/", temp)

		fmt.Println("soundsPath", soundsPath)

		temp = utils.GetPoe1Path("")
		gameDir1 := fmt.Sprintf("%s/", temp)
		utils.MkDirIfNotExist(gameDir1)
		cmd1 := exec.Command(
			"cp",
			"--update=none",
			"-r",
			soundsPath,
			gameDir1,
		)
		var outb, errb bytes.Buffer
		cmd1.Stdout = &outb
		cmd1.Stderr = &errb
		err = cmd1.Run()

		if err != nil {
			fmt.Println("error copying PoE 1 sound files", err, outb.String(), errb.String())
		}

		temp = utils.GetPoe2Path("")
		gameDir2 := fmt.Sprintf("%s/", temp)
		utils.MkDirIfNotExist(gameDir2)
		cmd2 := exec.Command(
			"cp",
			"--update=none",
			"-r",
			soundsPath,
			gameDir2,
		)
		err = cmd2.Run()
		if err != nil {
			fmt.Println("error copying PoE 2 sound files", err)
		}
	}
}

// @todo(nick-ng): move some functions to separate files
func processFilter(filterPath string, isImported bool) (string, ProcessedFilterFlags, []error) {
	var errorList []error

	flags := ProcessedFilterFlags{Game: "poe1"}

	filterData, err := os.ReadFile(filterPath)
	filterString := string(filterData)

	filterString = utils.PatchThirdPartyFilter(filterString)

	if err != nil {
		return "", flags, append(errorList, err)
	}

	rawLines := strings.Split(filterString, "\n")

	var filterChunks []string
	if !isImported {
		filterChunks = append(filterChunks, `Show
	#!DefaultBackground!# 120
	Continue`)
	}

	var currentCommand string
	var importAfterTimestamp int64
	now := time.Now().Unix()
	options := make(map[string][]string)
	// handle import command (multi-line)
	for _, rawLine := range rawLines {
		trimmedLine := strings.TrimSpace(rawLine)
		switch currentCommand {
		case "import":
			{
				subCommandArguments := getCommands(trimmedLine)
				switch subCommandArguments[0] {
				case "exclude":
					fallthrough
				case "del":
					fallthrough
				case "delete":
					{
						regexpString := strings.Join(subCommandArguments[1:], " ")
						options["delete"] = append(options["delete"], regexpString)
					}
				case "include":
					{
						regexpString := strings.Join(subCommandArguments[1:], " ")
						options["include"] = append(options["delete"], regexpString)
					}
				case "maxarea":
					{
						options["maxarea"] = []string{subCommandArguments[1]}
					}
				case "after":
					{
						importAfterTimestamp, err = strconv.ParseInt(subCommandArguments[1], 10, 0)
						if err != nil {
							errorString := fmt.Sprintf("couldn't parse import after timestamp: %s", subCommandArguments[1])
							errorList = append(errorList, errors.New(errorString))
						}
					}
				case "import":
					fallthrough
				case "noop": // the line didn't start with #!
					{
						// why is this "if" statement necessary? doesn't it match both cases?
						if !strings.HasPrefix(trimmedLine, "#") || subCommandArguments[0] == "import" {
							// it's a non-comment line so import the filter here then write the line
							_, present := options["file"]
							if present {
								_, fullPath, err := importBaseFilter(options["file"][0])

								if err != nil {
									errorList = append(errorList, err)
									filterChunks = append(filterChunks, fmt.Sprintf("#? error: couldn't import %s", options["file"]))
									filterChunks = append(filterChunks, fmt.Sprintf("#?  %s\n", err))
								}

								tempFilter, _, errs := processFilter(fullPath, true)

								if len(errs) != 0 {
									errorList = append(errorList, err)
									filterChunks = append(filterChunks, fmt.Sprintf("#? error: couldn't import %s", options["file"]))
									filterChunks = append(filterChunks, fmt.Sprintf("#?  %s\n", err))
								} else if now < int64(importAfterTimestamp) {
									filterChunks = append(filterChunks, fmt.Sprintf("#?  skipped because generated at %d", now))
								} else { // nothing wrong. we can modify the imported filter
									for _, deleteRegexpString := range options["delete"] {
										deleteRegexp, err := regexp.Compile(deleteRegexpString)

										if err != nil {
											errorString := fmt.Sprintf("couldn't compile regexp: %s", deleteRegexpString)
											errorList = append(errorList, errors.New(errorString))
											filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter = deleteRegexp.ReplaceAllString(tempFilter, "")
										}
									}

									for _, includeRegexpString := range options["include"] {
										includeRegexp, err := regexp.Compile(includeRegexpString)

										if err != nil {
											errorString := fmt.Sprintf("couldn't compile regexp: %s", includeRegexpString)
											errorList = append(errorList, errors.New(errorString))
											filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter = includeRegexp.ReplaceAllString(tempFilter, "")
										}
									}

									if len(options["maxarea"]) > 0 {
										maxarea, err := strconv.ParseInt(options["maxarea"][0], 10, 0)

										if err != nil {
											errorString := fmt.Sprintf("maxarea argument not an integer: %s", options["maxarea"][0])
											errorList = append(errorList, errors.New(errorString))
											filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
										} else {
											tempFilter2, err := utils.LimitMaxAreaLevel(tempFilter, int(maxarea))

											if err != nil {
												errorString := fmt.Sprintf("couldn't lower filter's level: %s", err)
												errorList = append(errorList, errors.New(errorString))
												filterChunks = append(filterChunks, fmt.Sprintf("#? error: %s", errorString))
											} else {
												tempFilter = tempFilter2
											}
										}
									}

									filterChunks = append(filterChunks, tempFilter)
									filterChunks = append(filterChunks, fmt.Sprintf("#? end of %s\n", options["file"]))
								}

							} else {
								filterChunks = append(filterChunks, "#? error: No file specified.\n")
							}

							currentCommand = ""
							clear(options)
						}

						if subCommandArguments[0] == "import" {
							currentCommand = "import"
							options["file"] = []string{subCommandArguments[1]}

							tempLine := utils.CleanUpCommand(rawLine)
							filterChunks = append(filterChunks, tempLine)
						} else {
							filterChunks = append(filterChunks, rawLine)
						}
					}
				default:
					{
						filterChunks = append(filterChunks, trimmedLine)

						warning := fmt.Sprintf("#? warning: Unknown sub-command %s", subCommandArguments[0])
						filterChunks = append(filterChunks, warning)
					}
				}
			}
		// there is no current command or the current line isn't a sub command of the current command
		default:
			{
				commandArguments := getCommands(trimmedLine)
				switch commandArguments[0] {
				case "poe2":
					{
						// when a filter is imported, the game flag is ignored
						flags.Game = "poe2"
						tempLine := utils.CleanUpCommand(rawLine)
						filterChunks = append(filterChunks, tempLine)
					}
				case "import":
					{
						currentCommand = "import"
						options["file"] = []string{strings.Join(commandArguments[1:], " ")}

						tempLine := utils.CleanUpCommand(rawLine)
						filterChunks = append(filterChunks, tempLine)
					}
				case "del":
					fallthrough
				case "delete":
					{
						filterChunks = append(filterChunks, rawLine)
						warning := fmt.Sprintf("#? warning: %s only allowed during an import", commandArguments[0])
						filterChunks = append(filterChunks, warning)
					}
				case "skip":
					{
						return "", flags, []error{}
					}
				case "noop":
					fallthrough
				default:
					{
						filterChunks = append(filterChunks, rawLine)
						// Handle in the next loop so we don't get double warnings
						// noop is also here
					}
				}
			}
		}
	}

	tempFilter := strings.Join(filterChunks, "\n")
	tempFilterLines := strings.Split(tempFilter, "\n")
	// handle import command (multi-line)
	var tempFilterChunks []string
	var rawCommand string
	var isCommandStarted bool
	clear(options)
	for _, line := range tempFilterLines {
		trimmedLine := strings.TrimSpace(line)
		commandArgs := getCommands(trimmedLine)
		command := strings.ToLower(commandArgs[0])
		if isCommandStarted && command != "custom" && command != "custombig" {
			isCommandStarted = false
			dropLevelChunk, err := utils.GetDropLevelFilter(rawCommand, options["custom"], options["custombig"])
			if err != nil {
				tempFilterChunks = append(tempFilterChunks, fmt.Sprintf("#? error: couldn't make droplevel filter from \"%s\" because %s", rawCommand, err))
			} else {
				tempFilterChunks = append(tempFilterChunks, dropLevelChunk)
			}
		}

		completedCommand := utils.CleanUpCommand(line)
		switch command {
		case "droplevel":
			{
				isCommandStarted = true
				rawCommand = line
				clear(options)
				tempFilterChunks = append(tempFilterChunks, completedCommand)
			}
		case "custom":
			fallthrough
		case "custombig":
			{
				customStyle := strings.Join(commandArgs[1:], " ")
				options[command] = append(options[command], customStyle)
				tempFilterChunks = append(tempFilterChunks, completedCommand)
			}
		default:
			{
				tempFilterChunks = append(tempFilterChunks, line)
			}
		}
	}

	filterChunks = tempFilterChunks

	// @todo(nick-ng): move these to a method so we can process all commands (multi-line or otherwise) in a single loop?
	// separate loop for all non-multi-line commands. if we encounter a
	// different command during a multi-line command, we need to handle the
	// multi-line command and the new command.
	tempFilter = strings.Join(filterChunks, "\n")
	tempFilterLines = strings.Split(tempFilter, "\n")
	clear(tempFilterChunks)
	for _, rawLine := range tempFilterLines {
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
					tempFilterChunks = append(tempFilterChunks, rawLine)
					// noop
				}
			case "links":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					tempFilterChunks = append(tempFilterChunks, tempLine)

					if len(commandArguments) < 2 {
						tempFilterChunks = append(tempFilterChunks, "#? warning: need at least coloured links (SocketGroup) to work")
						continue
					}

					filterBlock, err := utils.GetSocketGroupFilter(commandArguments[1], commandArguments[2:]...)

					if err != nil {
						e := fmt.Sprintf("#? error: couldn't generate links\n#?  %s", err)
						tempFilterChunks = append(tempFilterChunks, e)
					} else {
						tempFilterChunks = append(tempFilterChunks, filterBlock)
					}

				}
			case "linksa":
				fallthrough
			case "linksarmor":
				fallthrough
			case "linksarmour":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					tempFilterChunks = append(tempFilterChunks, tempLine)

					if len(commandArguments) < 2 {
						tempFilterChunks = append(tempFilterChunks, "#? warning: need at least coloured links (SocketGroup) to work")
						continue
					}

					newLine := utils.GetArmourSocketGroupFilter(commandArguments[1], commandArguments[2:]...)
					tempFilterChunks = append(tempFilterChunks, newLine)
				}
			case "weapon":
				fallthrough
			case "weapons":
				{
					tempLine := utils.CleanUpCommand(rawLine)
					tempFilterChunks = append(tempFilterChunks, tempLine)
					maxLevel := 99
					minLevel := 0
					seriousAdjustment := 1
					showAdjustment := 7

					if len(commandArguments) < 2 {
						tempFilterChunks = append(tempFilterChunks, "#? warning: need at least the weapon group (slow-wands, crit-bows, etc.) to work")
						continue
					}

					if len(commandArguments) >= 3 {
						maxLevel64, err := strconv.ParseInt(commandArguments[2], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse max level")
						} else {
							maxLevel = int(maxLevel64)
						}
					}

					if len(commandArguments) >= 4 {
						minLevel64, err := strconv.ParseInt(commandArguments[3], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse min level")
						} else {
							minLevel = int(minLevel64)
						}
					}

					if len(commandArguments) >= 5 {
						seriousAdjustment64, err := strconv.ParseInt(commandArguments[4], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse serious adjustment")
						} else {
							seriousAdjustment = int(seriousAdjustment64)
						}
					}

					if len(commandArguments) >= 6 {
						showAdjustment64, err := strconv.ParseInt(commandArguments[5], 10, 64)

						if err != nil {
							tempFilterChunks = append(tempFilterChunks, "#? couldn't parse show adjustment")
						} else {
							showAdjustment = int(showAdjustment64)
						}
					}

					newLine, err := utils.GetWeaponGroupFilter(commandArguments[1], maxLevel, minLevel, seriousAdjustment, showAdjustment)

					if err != nil {
						tempFilterChunks = append(tempFilterChunks, "#? warning: couldn't get weapon group")
						continue
					}

					tempFilterChunks = append(tempFilterChunks, newLine)
				}
			// // @todo(nick-ng): move this to its own loop
			// case "droplevel":
			// 	{
			// 		// makes item filter based on item drop level
			// 		_ = utils.ParseFlags(rawLine)
			// 		newLine, err := utils.GetDropLevelFilter(rawLine)

			// 		if err != nil {
			// 			tempFilterChunks = append(tempFilterChunks, "#? warning: couldn't get drop level filter")
			// 			continue
			// 		}

			// 		tempFilterChunks = append(tempFilterChunks, newLine)
			// 	}
			case "tts":
				{
					newLine := utils.MakeTts(trimmedLine, flags.Game)
					tempFilterChunks = append(tempFilterChunks, newLine)
				}
			default:
				{
					warning := fmt.Sprintf("#? warning: Unknown command %s", commandArguments[0])
					tempFilterChunks = append(tempFilterChunks, warning)
				}
			}
		} else {
			tempFilterChunks = append(tempFilterChunks, rawLine)
		}
	}

	filterChunks = tempFilterChunks
	filter := strings.Join(filterChunks, "\n")

	if !isImported {
		filter = utils.ApplyAllTokens(filter)
		filter = utils.CleanUpFilter(filter)
	}

	if strings.HasSuffix(filterPath, ".ruthlessfilter") {
		filter = strings.ReplaceAll(filter, "Hide", "Show")
	}

	return filter, flags, errorList
}

func importBaseFilter(filterName string) (string, string, error) {
	var err error

	for _, directoryName := range []string{BASE_FILTERS_PATH, BUILD_FILTERS_PATH, THIRD_PARTY_FILTERS_PATH} {
		path := filepath.Join(directoryName, filterName)

		filterData, err := os.ReadFile(path)

		if err == nil {
			filterString := string(filterData)
			return filterString, path, err
		}
	}

	return "", "", err
}

func getCommands(rawCommand string) []string {
	commands := strings.Split(rawCommand, " ")
	if commands[0] != "#!" {
		return []string{"noop"}
	}

	return commands[1:]
}
