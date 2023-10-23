package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const defaultURL = "https://api.onizmx.com/lambda/tower_stream"

type TowerData struct {
	totalRSSI float64
	count     int
}

func main() {
	// Check the farm ID is supplied as argument
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Invalid arguments, please supply the farm ID as argument, example:\n./main.exe [farm_id]")
		os.Exit(-1)
	}
	urlArray := getURLs(defaultURL)
	farmID := strings.TrimSpace(args[1])

	// Read CSV Data for each URL
	allTowerData := make(map[string]*TowerData)
	errorCnt := 0
	fmt.Println(len(urlArray), "CSV objects to retrieve.")
	for _, url := range urlArray {
		var towerData, error = getCSVData(url, farmID)
		if error != nil {
			fmt.Println("Error reading CSV" + error.Error())
			errorCnt++
			continue
		}
		// Merge maps
		allTowerData = mergeMaps(allTowerData, towerData)
	}
	fmt.Println((len(urlArray) - errorCnt), "out of", len(urlArray), "CSV files were successfully retrieved.\n=========================================================")
	if len(allTowerData) == 0 {
		fmt.Println("No farm with ID", farmID, "found")
		os.Exit(0)
	}

	bestTowerID := "none"
	bestTowerRSSI := float64(-1000)
	fmt.Println("Towers in farm", farmID, ":\n")
	for farm_id, farmdata := range allTowerData {
		averageRSSI := farmdata.totalRSSI / float64(farmdata.count)
		fmt.Println("Tower ID", farm_id)
		fmt.Println("Average RSSI", averageRSSI, "\n")
		if bestTowerRSSI < averageRSSI {
			bestTowerRSSI = averageRSSI
			bestTowerID = farm_id
		}
	}
	// Announce best tower in farm
	fmt.Println("=========================================================\nBest tower:")
	fmt.Println("Tower ID", bestTowerID)
	fmt.Println("Average RSSI", bestTowerRSSI)
}

func getURLs(url string) []string {
	// Receive the GET response
	response, error := http.Get(url)
	if error != nil {
		fmt.Println("Error occured in retrieving array:" + error.Error())
		os.Exit(-1)
	}

	// Store all of the response body
	responseBody, error := io.ReadAll(response.Body)
	if error != nil {
		fmt.Println("Error occured in reading array:" + error.Error())
		os.Exit(-1)
	}

	// Deserialize JSON data to array
	var urlArray []string
	if json.Unmarshal(responseBody, &urlArray) != nil {
		fmt.Println("Error occured in unmarshalling json: " + error.Error())
		os.Exit(-1)
	}

	// Clean up memory
	defer response.Body.Close()
	return urlArray
}

// Returns a nested map of farm_id : tower_id : TowerData
func getCSVData(url string, farmID string) (map[string]*TowerData, error) {
	farmData := make(map[string]*TowerData)
	// Get response
	request, error := http.NewRequest("GET", url, nil)
	if error != nil {
		fmt.Println(error)
	}

	//request.Header.Set("X-Amz-Security-Token", "test")

	client := &http.Client{}
	resp, error := client.Do(request)
	//fmt.Println(resp.Header)

	// Read CSV
	reader := csv.NewReader(resp.Body)
	// Prune the titles
	reader.Read()
	for {
		record, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			return farmData, error
		}

		// Check record
		if len(record) != 3 {
			return farmData, fmt.Errorf("")
		}

		// Save record
		thisFarmID := record[0]
		towerID := record[1]
		RSSI, _ := strconv.ParseFloat(record[2], 64)

		if thisFarmID != farmID {
			continue
		}

		if _, ok := farmData[towerID]; !ok {
			farmData[towerID] = &TowerData{}
		}
		farmData[towerID].totalRSSI += RSSI
		farmData[towerID].count++
	}

	resp.Body.Close()
	fmt.Println("Successfully processed CSV")
	return farmData, nil
}

func mergeMaps(maps ...map[string]*TowerData) map[string]*TowerData {
	result := make(map[string]*TowerData)

	for _, m := range maps {
		for towerID, towerData := range m {
			if _, ok := result[towerID]; !ok {
				result[towerID] = &TowerData{}
			}
			result[towerID].totalRSSI += towerData.totalRSSI
			result[towerID].count += towerData.count
		}
	}
	return result
}
