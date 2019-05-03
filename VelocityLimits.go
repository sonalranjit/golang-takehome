package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Activity main datastruct that is being decoded from the text file as json
type Activity struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	LoadAmount string `json:"load_amount"`
	Time       string `json:"time"`
}

// Struct to hold the count of the running weekly total
type ActivityByWeek struct {
	WeekTotal float64
}

// Holds the data for the runnning daily total and transactions count grouped by day
type DailyTransactionCount struct {
	ID               string
	TransactionDay   string
	TransactionCount int
	DailyTotal       float64
}

// Struct for return the response after parsing the data from the input file
type ReturnResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Accepted   bool   `json:"accepted"`
}

// Declaring 2 global variables, Activity grouped by week and activity grouped by day
var (
	weeklyActivity = make(map[string]map[string]ActivityByWeek)
	dailyActivity  = make(map[string]map[string]DailyTransactionCount)
)

/*Reading files requires checkign most calls for errors. This helper
will streamline our error checks below.*/
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Converts a datetime into a Year-ISOWeek string using to map and group
func getYearWeekNumber(dateTimeString string) string {
	activityDate, err := time.Parse("2006-01-02T15:04:05Z07:00", dateTimeString)
	check(err)
	_, weekNumber := activityDate.ISOWeek()
	yearWeekNumber := string(activityDate.Year()) + string(weekNumber)
	return yearWeekNumber
}

// Validate the Weekly limit of $20,000
func checkWeekTotal(weekData ActivityByWeek) bool {
	if weekData.WeekTotal > 20000 {
		return false
	} else {
		return true
	}
}

// Validate the Daily activity, of $5,000 a day and 3 loading transactions
func checkDailyCount(customerID string, dayString string, loadAmount float64) bool {

	dailyData := dailyActivity[customerID][dayString]

	//If current running daily total is > $5,000 subtract the latest amount decreased the count
	if dailyData.DailyTotal > 5000 {
		_dailyData := DailyTransactionCount{}
		_dailyData.TransactionCount = dailyData.TransactionCount - 1
		_dailyData.DailyTotal = dailyData.DailyTotal - loadAmount
		dailyActivity[customerID][dayString] = _dailyData
		return false
	} else if dailyData.TransactionCount > 3 {
		return false
	} else {
		return true
	}
}

//Returns the response struct in a json string format
func returnResponse(id string, customerID string, accepted bool) string {
	_response := &ReturnResponse{
		ID:         id,
		CustomerID: customerID,
		Accepted:   accepted}
	response, _ := json.Marshal(_response)
	return string(response)
}

// Parse the incoming data from the text file and maps it to its corresponding struct grouped by Customer ID and daystring
func handleDailyActivity(data Activity, dayString string, loadAmount float64) {
	if _, ok := dailyActivity[data.CustomerID]; ok {
		if dailyData, ok := dailyActivity[data.CustomerID][dayString]; ok {
			dailyData.TransactionCount++
			dailyData.DailyTotal += loadAmount
			dailyActivity[data.CustomerID][dayString] = dailyData
		} else {
			dailyData.TransactionDay = dayString
			dailyData.TransactionCount = 1
			dailyData.DailyTotal = loadAmount
			dailyActivity[data.CustomerID][dayString] = dailyData
		}
	} else {
		dailyActivity[data.CustomerID] = make(map[string]DailyTransactionCount)
		dailyData := DailyTransactionCount{}
		dailyData.ID = data.ID
		dailyData.TransactionDay = dayString
		dailyData.TransactionCount = 1
		dailyData.DailyTotal = loadAmount
		dailyActivity[data.CustomerID][dayString] = dailyData
	}
}

// Parse the incoming data from the text file and maps it tothe its corresponding struct grouped by Customer ID and yearWeekNumber
func handleWeeklyActivity(data Activity, yearWeekNumber string, loadAmount float64) {
	if _, ok := weeklyActivity[data.CustomerID]; ok {
		if weekData, ok := weeklyActivity[data.CustomerID][yearWeekNumber]; ok {
			weekData.WeekTotal += loadAmount
			weeklyActivity[data.CustomerID][yearWeekNumber] = weekData
		} else {
			weekData.WeekTotal = loadAmount
			weeklyActivity[data.CustomerID][yearWeekNumber] = weekData
		}
	} else {
		weeklyActivity[data.CustomerID] = make(map[string]ActivityByWeek)
		weekData := ActivityByWeek{}
		weekData.WeekTotal = loadAmount
		weeklyActivity[data.CustomerID][yearWeekNumber] = weekData
	}
}

// Test to check generated file from reference file
func checkOutput(generatedFile string, referenceFile string) bool {
	genF, err := os.Open("./generated-output.txt")
	check(err)

	refF, err := os.Open("./output.txt")
	check(err)

	var genFileLines []string
	genScanner := bufio.NewScanner(genF)

	for genScanner.Scan() {
		genFileLines = append(genFileLines, genScanner.Text())
	}

	var refFileLines []string
	refScanner := bufio.NewScanner(refF)

	for refScanner.Scan() {
		refFileLines = append(refFileLines, refScanner.Text())
	}

	passCount := 0
	totalCount := 0
	if len(genFileLines) == len(refFileLines) {
		for i := 0; i < len(genFileLines); i++ {
			if genFileLines[i] == refFileLines[i] {
				passCount++
			} else {
				fmt.Println(genFileLines, " failed")
			}
			fmt.Printf("%d/%d cases passed\n", passCount, i+1)
			totalCount++
		}
		if passCount == totalCount {
			return true
		} else {
			return false
		}
	} else {
		fmt.Println("Length Mismatch for 2 files.")
		return false
	}
}

func main() {

	//Open file for reading
	f, err := os.Open("./input.txt")
	check(err)

	//Open file for writing the output
	outFile, err := os.Create("./generated-output.txt")
	w := bufio.NewWriter(outFile)

	// Initialize the bufio scanner
	scanner := bufio.NewScanner(f)

	// Map to check there are no duplicate transaction ID for same User
	idCount := make(map[string]map[string]int)

	for scanner.Scan() {

		data := Activity{}

		// Decode and cast the json data to the Activity type struct
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			panic(err)
		}

		//Parse the string representing the currency
		loadAmount, parseErr := strconv.ParseFloat(data.LoadAmount[1:], 32)
		check(parseErr)

		//handle duplicate ID for same customer ID
		if _, ok := idCount[data.CustomerID]; ok {
			if _, ok := idCount[data.CustomerID][data.ID]; ok {
				continue
			} else {
				idCount[data.CustomerID][data.ID] = 1
			}
		} else {
			idCount[data.CustomerID] = make(map[string]int)
			idCount[data.CustomerID][data.ID] = 1
		}

		// If daily amount is greater than 5,000 return false
		if loadAmount > 5000 {
			outString := returnResponse(data.ID, data.CustomerID, false) + "\n"
			_, err := w.WriteString(outString)
			check(err)
		} else {

			//Get yearWeekNumber used for map later
			yearWeekNumber := getYearWeekNumber(data.Time)

			// Day string to group Activity by Day
			dayString := data.Time[:10]

			//Handle the data by it's grouping
			handleDailyActivity(data, dayString, loadAmount)
			handleWeeklyActivity(data, yearWeekNumber, loadAmount)

			//Validate the transactions
			isDailyValid := checkDailyCount(data.CustomerID, dayString, loadAmount)
			isWeeklyValid := checkWeekTotal(weeklyActivity[data.CustomerID][yearWeekNumber])

			//If both transcations are valid, return a true response
			if isDailyValid && isWeeklyValid {
				outString := returnResponse(data.ID, data.CustomerID, true) + "\n"
				_, err := w.WriteString(outString)
				check(err)
			} else {
				outString := returnResponse(data.ID, data.CustomerID, false) + "\n"
				_, err := w.WriteString(outString)
				check(err)
			}
		}
	}
	if scanner.Err() != nil {
		panic(scanner.Err)
	}
	w.Flush()
	f.Close()

	fmt.Println("ALL test cases passed ", checkOutput("./generated-output.txt", "./output.txt"))
}
