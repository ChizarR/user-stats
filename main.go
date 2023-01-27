package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	getUserStatsURL = "http://78.24.216.23:5050/users_stats/views/today"
)

func main() {
	file, err := os.Open("ids.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	ids := []int{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		idRaw := strings.TrimFunc(scanner.Text(), func(r rune) bool { return !unicode.IsNumber(r) })
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}

	client := http.Client{Timeout: 15 * time.Second}
	analyzer := NewAnalyzer(&client)

	activeUsers := 0
	inactiveUsers := 0
	totalViews := map[string]int{}
	for _, id := range ids {
		data, err := analyzer.getUserStats(id)
		if err != nil {
			panic(err)
		}
		if data.Ok {
			views := data.Result.Views
			if len(views) == 0 {
				inactiveUsers += 1
				continue
			}
			activeUsers += 1

			for k, v := range views {
				_, ok := totalViews[k]
				if ok {
					totalViews[k] += v
					continue
				}
				totalViews[k] = v
			}
		}
	}

	activityReport := fmt.Sprintf("Users Report:\n - Active new users: %d\n - Inactive new users: %d\n", activeUsers, inactiveUsers)
	viewsReport := "\nViews Report:\n"
	for category, num := range totalViews {
		categoryData := fmt.Sprintf(" - %s: %d\n", category, num)
		viewsReport += categoryData
	}
	report := activityReport + viewsReport

	fmt.Print(report)
}

type Response struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		TgId  int            `json:"tg_id"`
		Views map[string]int `json:"views"`
	} `json:"result"`
}

type Analyzer struct {
	client *http.Client
}

func NewAnalyzer(client *http.Client) *Analyzer {
	return &Analyzer{client: client}
}

func (a *Analyzer) getUserStats(tgId int) (Response, error) {
	bodyBytes, err := json.Marshal(map[string]int{"tg_id": tgId})
	if err != nil {
		return Response{}, err
	}
	bodyReader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest(http.MethodGet, getUserStatsURL, bodyReader)
	resp, err := a.client.Do(req)
	if err != nil {
		return Response{}, err
	}

	bodyBytes, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return Response{}, err
	}

	var data Response
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return Response{}, err
	}

	return data, nil
}
