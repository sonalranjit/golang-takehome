package validator

import (
	"encoding/json"
	"strconv"
	"time"

	"../activity"
)

var (
	weeklyActivity = make(map[string]map[string]activity.ActivityByWeek)
	dailyActivity  = make(map[string]map[string]activity.DailyTransactionCount)
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func generateResponse(id string, customerID string, accepted bool) string {
	_response := activity.ReturnResponse{
		ID:         id,
		CustomerID: customerID,
		Accepted:   accepted}
	response, _ := json.Marshal(_response)
	return string(response)
}

func getYearWeekNumber(dateTimeString string) string {
	activityDate, err := time.Parse("2006-01-02T15:04:05Z07:00", dateTimeString)
	check(err)
	_, weekNumber := activityDate.ISOWeek()
	yearWeeknumber := string(activityDate.Year()) + string(weekNumber)
	return yearWeeknumber
}

func checkWeekTotal(weekData activity.ActivityByWeek) bool {
	if weekData.WeekTotal > 20000 {
		return false
	}
	return true
}

func checkDailyCount(customerID string, dayString string, loadAmount float64) bool {

	dailyData := dailyActivity[customerID][dayString]

	//If current running daily total is > $5000 subtract the latest amount and decrease the count
	if dailyData.DailyTotal > 5000 {
		_dailyData := activity.DailyTransactionCount{}
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

func handleDailyActivity(activityData activity.Activity, dayString string, loadAmount float64) bool {
	if _, ok := dailyActivity[activityData.CustomerID]; ok {
		if dailyData, ok := dailyActivity[activityData.CustomerID][dayString]; ok {
			dailyData.TransactionCount++
			dailyData.DailyTotal += loadAmount
			dailyActivity[activityData.CustomerID][dayString] = dailyData
		} else {
			dailyData.TransactionDay = dayString
			dailyData.TransactionCount = 1
			dailyData.DailyTotal = loadAmount
			dailyActivity[activityData.CustomerID][dayString] = dailyData
		}
	} else {
		dailyActivity[activityData.CustomerID] = make(map[string]activity.DailyTransactionCount)
		dailyData := activity.DailyTransactionCount{}
		dailyData.ID = activityData.ID
		dailyData.TransactionDay = dayString
		dailyData.TransactionCount = 1
		dailyData.DailyTotal = loadAmount
		dailyActivity[activityData.CustomerID][dayString] = dailyData
	}
	return checkDailyCount(activityData.CustomerID, dayString, loadAmount)
}

func handleweeklyActivity(activityData activity.Activity, yearWeekNumber string, loadAmount float64) bool {
	if _, ok := weeklyActivity[activityData.CustomerID]; ok {
		if weekData, ok := weeklyActivity[activityData.CustomerID][yearWeekNumber]; ok {
			weekData.WeekTotal += loadAmount
			weeklyActivity[activityData.CustomerID][yearWeekNumber] = weekData
		} else {
			weekData.WeekTotal = loadAmount
			weeklyActivity[activityData.CustomerID][yearWeekNumber] = weekData
		}
	} else {
		weeklyActivity[activityData.CustomerID] = make(map[string]activity.ActivityByWeek)
		weekData := activity.ActivityByWeek{}
		weekData.WeekTotal = loadAmount
		weeklyActivity[activityData.CustomerID][yearWeekNumber] = weekData
	}

	return checkWeekTotal(weeklyActivity[activityData.CustomerID][yearWeekNumber])
}

func ValidateEvent(activityData activity.Activity) string {

	//Parse the string representing the currency
	loadAmount, parseErr := strconv.ParseFloat(activityData.LoadAmount[1:], 32)
	check(parseErr)
	if loadAmount > 5000 {
		return generateResponse(activityData.ID, activityData.CustomerID, false)
	} else {
		//Get yearWeekNumber used for maps
		yearWeekNumber := getYearWeekNumber(activityData.Time)

		// Day string to group Activity by Day
		dayString := activityData.Time[:10]

		// Handle the data put it into their respective structs
		isDailyValid := handleDailyActivity(activityData, dayString, loadAmount)
		isWeeklyValid := handleweeklyActivity(activityData, yearWeekNumber, loadAmount)

		if isDailyValid && isWeeklyValid {
			return generateResponse(activityData.ID, activityData.CustomerID, true)
		} else {
			return generateResponse(activityData.ID, activityData.CustomerID, false)
		}

	}

}
