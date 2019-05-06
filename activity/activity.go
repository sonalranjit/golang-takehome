package activity

// Activity type struct main datastruct that is being decoded from the text file as json
type Activity struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	LoadAmount string `json:"load_amount"`
	Time       string `json:"time"`
}

// ActivityByWeek holds the coun of the running weekly total
type ActivityByWeek struct {
	WeekTotal float64
}

// DailyTransaction holds the data for running daily total and transactions count grouped by day
type DailyTransactionCount struct {
	ID               string
	TransactionDay   string
	TransactionCount int
	DailyTotal       float64
}

//Return response is struct for the output message
type ReturnResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Accepted   bool   `json:"accepted"`
}
