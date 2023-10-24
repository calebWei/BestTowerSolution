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

// The URL to the list of CSV URLs
const defaultURL = "https://api.onizmx.com/lambda/tower_stream"

// Keeps track of total RSSI and count for each tower to calculate the average, saves memory
type TowerData struct {
	totalRSSI float64
	count     int
}

func main() {
	// Check farm ID is supplied as argument
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Invalid arguments, please supply the farm ID as argument, example:\n./main.exe [farm_id]")
		os.Exit(-1)
	}
	farmID := strings.TrimSpace(args[1])

	// Obtain CSV URLs
	urlArray := getURLs(defaultURL)

	// Read CSV data for each URL
	allTowerData := getAllCSVData(urlArray, farmID)
	fmt.Print("=========================================================\n\n")

	// Find the tower with the best (highest) average RSSI
	bestTowerID, bestTowerRSSI := findBestTower(allTowerData, farmID)

	// Announce best tower in farm
	fmt.Print("=========================================================\n\n")
	fmt.Println("Best tower:")
	fmt.Println("Tower ID", bestTowerID)
	fmt.Println("Average RSSI", bestTowerRSSI)
}

// Returns an array of string URLs, given an URL to the array
func getURLs(url string) []string {
	// Collect the GET response
	response, error := http.Get(url)
	checkError(error)

	// Read response body to byte stream
	responseBody, error := io.ReadAll(response.Body)
	checkError(error)

	// Deserialize byte stream to array
	var urlArray []string
	error = json.Unmarshal(responseBody, &urlArray)
	checkError(error)

	defer response.Body.Close()
	return urlArray
}

// Collect and Merge all CSV files to a map of tower_id : TowerData
// Only consider towers with the specified farm_id
// Ignore unsuccssful retrieval of CSVs
func getAllCSVData(urlArray []string, farmID string) map[string]*TowerData {
	// Retrieve CSV for each URL, ignore failed requests
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
		// Merge tower data into one map
		allTowerData = mergeMaps(allTowerData, towerData)
	}
	fmt.Print((len(urlArray) - errorCnt), " out of ", len(urlArray), " CSV files were successfully retrieved.\n\n")

	// Check there is at least 1 farm ID match
	if len(allTowerData) == 0 {
		fmt.Println("No farm with ID", farmID, "found, exiting program.")
		os.Exit(0)
	}
	return allTowerData
}

// Returns a nested map of tower_id : TowerData for a particular CSV file
// Only consider towers with the specified farm_id
func getCSVData(url string, farmID string) (map[string]*TowerData, error) {
	farmData := make(map[string]*TowerData)
	// Get the CSV file
	request, error := http.NewRequest("GET", url, nil)
	checkError(error)
	response, error := http.DefaultClient.Do(request)

	// Initialize csv reader
	reader := csv.NewReader(response.Body)

	// Prune the titles first, then read CSV record by record
	reader.Read()
	for {
		// Handle errors on Read
		record, error := reader.Read()
		if error == io.EOF {
			break
		}
		checkError(error)

		// Validify record
		if len(record) != 3 {
			return farmData, fmt.Errorf("")
		}

		// Read record data
		thisFarmID := record[0]
		towerID := record[1]
		RSSI, _ := strconv.ParseFloat(record[2], 64)

		// Skip if farm ID does not match
		if thisFarmID != farmID {
			continue
		}

		// If key for tower doesn't exist, create it, then update its data
		if _, ok := farmData[towerID]; !ok {
			farmData[towerID] = &TowerData{}
		}
		farmData[towerID].totalRSSI += RSSI
		farmData[towerID].count++
	}

	defer response.Body.Close()
	fmt.Println("Successfully processed CSV")
	return farmData, nil
}

// Merge and return two or more tower_id : TowerData maps
func mergeMaps(maps ...map[string]*TowerData) map[string]*TowerData {
	result := make(map[string]*TowerData)
	for _, m := range maps {
		for towerID, towerData := range m {
			// If key for tower doesn't exist, create it, then update its data
			if _, ok := result[towerID]; !ok {
				result[towerID] = &TowerData{}
			}
			result[towerID].totalRSSI += towerData.totalRSSI
			result[towerID].count += towerData.count
		}
	}
	return result
}

// Given a map of tower_id : TowerData, return tower_id and average RSSI of the best tower
func findBestTower(allTowerData map[string]*TowerData, farmID string) (string, float64) {
	bestTowerID := "None"
	bestTowerRSSI := float64(-1000)
	fmt.Print("Towers in farm ", farmID, ":\n\n")
	for farm_id, farmdata := range allTowerData {
		averageRSSI := farmdata.totalRSSI / float64(farmdata.count)
		fmt.Println("Tower ID", farm_id)
		fmt.Print("Average RSSI ", averageRSSI, "\n\n")
		if bestTowerRSSI < averageRSSI {
			bestTowerRSSI = averageRSSI
			bestTowerID = farm_id
		}
	}
	return bestTowerID, bestTowerRSSI
}

// Raise error if present
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
