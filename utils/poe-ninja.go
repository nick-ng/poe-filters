package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type PoeNinjaData struct {
	Leagues        []PoeLeague
	CurrencyPrices map[string]CurrencyPrices
}

type CurrencyPrices struct {
	FetchedAt      time.Time
	DivinePerChaos float64
	Prices         []CurrencyPrice
}

type CurrencyPrice struct {
	PoeNinjaId string
	BaseType   string
	ChaosValue float64
}

type PoeNinjaIndexStateItem struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	DisplayName string `json:"displayName"`
}

type PoeNinjaDivinationCardItem struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Icon         string  `json:"icon"`
	BaseType     string  `json:"baseType"`
	StackSize    int     `json:"stackSize"`
	ChaosValue   float64 `json:"chaosValue"`
	ExaltedValue float64 `json:"exaltedValue"`
	DivineValue  float64 `json:"divineValue"`
}

type PoeNinjaDivinationCardsResponse struct {
	Lines     []PoeNinjaDivinationCardItem `json:"lines"`
	Timestamp int64                        `json:"timestamp"`
}

type PoeNinjaCurrencyResponse struct {
	Core  PoeNinjaCurrencyResponseCore `json:"core"`
	Lines []PoeNinjaCurrencyLine       `json:"lines"`
	Items []PoeNinjaCurrencyItem       `json:"items"`
}

type PoeNinjaCurrencyResponseCore struct {
	Rates map[string]float64 `json:"rates"`
}

type PoeNinjaCurrencyLine struct {
	Id           string  `json:"id"`
	PrimaryValue float64 `json:"primaryValue"`
}

type PoeNinjaCurrencyItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	DetailsId string `json:"detailsId"`
}

type PoeLeague struct {
	Name        string
	DisplayName string
	Url         string
	IsHardcore  bool
	IsEternal   bool
}

var eternalLeagues = []string{"standard", "hardcore"}
var divinationCardsFilterPath = filepath.Join("base-filters", "full-stack-divination-cards.filter")

const MAX_CACHE_AGE = 0.9 * 60 * 60 // just under 1 hour in seconds

func CreatePoeNinjaData() PoeNinjaData {
	poeNinjaData := PoeNinjaData{
		CurrencyPrices: map[string]CurrencyPrices{},
	}
	poeNinjaData.UpdateLeagues()

	return poeNinjaData
}

func (pnd *PoeNinjaData) UpdateLeagues() {
	leagues, err := GetTradeChallengeLeagues()
	if err != nil {
		slog.Error("error getting challenge leagues", "err", err)
		return
	}

	pnd.Leagues = leagues
}

func (pnd *PoeNinjaData) GetLeague(hardcore bool, eternal bool) (PoeLeague, error) {
	for _, league := range pnd.Leagues {
		if league.IsEternal == eternal && league.IsHardcore == hardcore {
			return league, nil
		}
	}

	newError := errors.New("no matching league found")

	return PoeLeague{}, newError
}

func (pnd *PoeNinjaData) GetCurrencyPrices(hardcore bool, eternal bool) (CurrencyPrices, error) {
	league, err := pnd.GetLeague(hardcore, eternal)
	if err != nil {
		return CurrencyPrices{}, err
	}
	temp, ok := pnd.CurrencyPrices[league.Name]

	cutoffTime := time.Now().Add(-time.Second * MAX_CACHE_AGE)
	if !ok || temp.FetchedAt.Before(cutoffTime) {
		pnd.UpdateCurrencyPrices(hardcore, eternal)
	}

	temp, ok = pnd.CurrencyPrices[league.Name]
	if !ok {
		return CurrencyPrices{}, errors.New("no prices even after update")
	}

	return temp, nil
}

func (pnd *PoeNinjaData) UpdateCurrencyPrices(hardcore bool, eternal bool) error {
	league, err := pnd.GetLeague(hardcore, eternal)
	if err != nil {
		return err
	}

	currencyUrl := fmt.Sprintf("https://poe.ninja/poe1/api/economy/exchange/current/overview?league=%s&type=Currency", league.Name)

	slog.Debug("currencyUrl", "url", currencyUrl)

	resp, err := http.Get(currencyUrl)

	if err != nil {
		slog.Error("error: couldn't fetch currency", "league", league, "currencyUrl", currencyUrl, "error", err)
		return err
	}

	resBody, err := io.ReadAll(resp.Body)

	if err != nil {
		slog.Error("error: couldn't get currency body", "league", league, "currencyUrl", currencyUrl, "error", err)
		return err
	}

	var poeNinjaCurrencyResponse PoeNinjaCurrencyResponse
	err = json.Unmarshal(resBody, &poeNinjaCurrencyResponse)

	if err != nil {
		slog.Error("error: couldn't decode currency body", "league", league, "currencyUrl", currencyUrl, "error", err)
		return err
	}

	newPrices := []CurrencyPrice{}
	for _, item := range poeNinjaCurrencyResponse.Items {
		for _, line := range poeNinjaCurrencyResponse.Lines {
			if item.Id == line.Id {
				newPrice := CurrencyPrice{
					PoeNinjaId: item.Id,
					BaseType:   item.Name,
					ChaosValue: line.PrimaryValue,
				}

				newPrices = append(newPrices, newPrice)

				break
			}
		}
	}

	divinePerChaos, ok := poeNinjaCurrencyResponse.Core.Rates["divine"]
	if !ok {
		divinePerChaos = 1.0 / 150.0
	}

	pnd.CurrencyPrices[league.Name] = CurrencyPrices{
		FetchedAt:      time.Now(),
		DivinePerChaos: divinePerChaos,
		Prices:         newPrices,
	}

	return nil
}

func GetIndexState() (map[string][]PoeNinjaIndexStateItem, error) {
	resp, err := http.Get("https://poe.ninja/poe1/api/data/index-state")

	if err != nil {
		fmt.Println("couldn't fetch poe.ninja stats", err)
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("couldn't get body", err)
		return nil, err
	}

	poeNinjaStateIndex := map[string][]PoeNinjaIndexStateItem{}

	json.Unmarshal(resBody, &poeNinjaStateIndex)

	return poeNinjaStateIndex, nil
}

func GetTradeChallengeLeagues() ([]PoeLeague, error) {
	poeLeagues := []PoeLeague{}

	poeNinjaStateIndex, err := GetIndexState()

	if err != nil {
		fmt.Println("couldn't get league names")
		return nil, err
	}

	for _, v := range poeNinjaStateIndex["economyLeagues"] {
		isEternalLeague := slices.Contains(eternalLeagues, v.Url)

		poeLeagues = append(poeLeagues, PoeLeague{
			Name:        v.Name,
			DisplayName: v.DisplayName,
			Url:         v.Url,
			IsHardcore:  v.Url == "hardcore" || strings.HasSuffix(v.Url, "hc"),
			IsEternal:   isEternalLeague,
		})
	}

	return poeLeagues, nil
}

// @todo(nick-ng): the poe.ninja endpoint this calls doesn't return stack size anymore. That is in a different endpoint. you have to update this before you can use it again
func MakeDivinationCardsFilterPoeNinja() error {
	leagues, err := GetTradeChallengeLeagues()
	if err != nil {
		slog.Error("error getting challenge leagues", "err", err)
		return err
	}

	var softcoreLeague PoeLeague

	for _, league := range leagues {
		if !league.IsHardcore && !league.IsEternal {
			softcoreLeague = league
			break
		}
	}

	if softcoreLeague.Name == "" {
		slog.Error("error getting challenge leagues", "err", err)
		return errors.New("no softcore trade league found")
	}

	cachePath := filepath.Join("cache", strings.ToLower(fmt.Sprintf("poeninja-divination-cards-%s.json", softcoreLeague.Name)))

	cacheData, err := os.ReadFile(cachePath)

	poeNinjaDivinationCardsResponse := PoeNinjaDivinationCardsResponse{}

	now := time.Now()

	needRefetch := true

	if err == nil {
		json.Unmarshal(cacheData, &poeNinjaDivinationCardsResponse)

		if (poeNinjaDivinationCardsResponse.Timestamp - now.Unix()) < MAX_CACHE_AGE {
			needRefetch = false
		}
	}

	if needRefetch {
		url := fmt.Sprintf("https://poe.ninja/poe1/api/economy/exchange/current/overview?league=%s&type=DivinationCard", softcoreLeague.Name)

		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("couldn't fetch divination cards")
			return errors.New("couldn't fetch divination cards")
		}

		resBody, err := io.ReadAll(resp.Body)

		if err != nil {
			slog.Error("error: couldn't get body", "err", err)
			return errors.New("couldn't get body")
		}

		err = json.Unmarshal(resBody, &poeNinjaDivinationCardsResponse)

		if err != nil {
			slog.Error("error decoding divination cards response", "err", err, "url", url)
			return err
		}

		poeNinjaDivinationCardsResponse.Timestamp = now.Unix()

		cacheData, err := json.Marshal(poeNinjaDivinationCardsResponse)

		if err == nil {
			err = os.WriteFile(cachePath, []byte(cacheData), 0666)

			if err != nil {
				slog.Error("error encoding divination cards to JSON", "err", err)
				return err
			}
		}
	}

	divinationCardsByStackSize := make(map[int][]string)

	maxStackSize := 0

	for _, divinationCard := range poeNinjaDivinationCardsResponse.Lines {
		stackSize := divinationCard.StackSize

		divinationCardsByStackSize[stackSize] = append(divinationCardsByStackSize[stackSize], divinationCard.BaseType)

		if stackSize > maxStackSize {
			maxStackSize = stackSize
		}
	}

	var filterLines []string

	for i := 0; i <= maxStackSize; i++ {
		divinationCardNames := divinationCardsByStackSize[i]

		slices.Sort(divinationCardNames)

		if len(divinationCardNames) > 0 {
			filterLines =
				append(filterLines,
					"Show",
					fmt.Sprintf("\tBaseType == \"%s\"", strings.Join(divinationCardNames, "\" \"")),
					"\tClass \"Divination\"",
					"\tSetFontSize 45",
					"\t#!BrightBackground!# 255",
					"\t#!CurrencyBorder!# 230",
					"\tMinimapIcon 0 Pink UpsideDownHouse",
					"\tCustomAlertSound \"sounds/intel-thanks-steve.mp3\"")

			if i > 1 {
				filterLines = append(filterLines, fmt.Sprintf("\tStackSize >= %d", i))
			}

			filterLines = append(filterLines, "")
		}
	}

	filter := strings.Join(filterLines, "\n")

	os.WriteFile(divinationCardsFilterPath, []byte(filter), 0666)

	fmt.Printf("%d divination cards found\n", len(poeNinjaDivinationCardsResponse.Lines))

	return nil
}
