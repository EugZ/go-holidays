package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EugZ/go-holidays/ask"
	"github.com/EugZ/go-holidays/employees"
)

func main() {
	toStart := ask.DefinePart(1, "Holiday API")

	if !toStart {
		return
	}

	resp, err := http.Get("https://date.nager.at/api/v2/publicholidays/2020/UA")

	if err != nil {
		fmt.Println("ERROR:", err)
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	var jsonEncodedHolidays string
	for scanner.Scan() {
		jsonEncodedHolidays += scanner.Text()
	}

	decodedHolidays := decodeHolidays(jsonEncodedHolidays)

	currentDate := getCurrentDate()

	checkHolidays(decodedHolidays, currentDate)

	toContinue := ask.DefinePart(2, "MySQL employees database")

	if !toContinue {
		return
	}

	employees.ExamineEmployees()
}

type holiday struct {
	Date        string
	LocalName   string
	Name        string
	CountryCode string
	Fixed       bool
	Global      bool
	Counties    interface{}
	LaunchYear  interface{}
	Type        string
}

func getCurrentDate() time.Time {
	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.UTC().Location())
	return currentDate
}

func decodeHolidays(holidays string) []holiday {
	bs := []byte(holidays)

	var holidaysDecoded []holiday

	err := json.Unmarshal(bs, &holidaysDecoded)

	if err != nil {
		fmt.Println(err)
	}

	return holidaysDecoded
}

func checkHolidays(holidays []holiday, currentDate time.Time) {
	const layoutISO = "2006-01-02"
	var isToday bool
	var closestHoliday holiday
	var holidayLength float64
	var dateRange string

	for _, holiday := range holidays {
		holidayDate, err := time.Parse(layoutISO, holiday.Date)

		if err != nil {
			fmt.Println(err)
		}
		if currentDate.Equal(holidayDate) {
			isToday = true
		}
		if currentDate.Equal(holidayDate) || currentDate.Before(holidayDate) {
			closestHoliday = holiday
			holidayLength, dateRange = defineHolidayLength(holidayDate)
			break
		}
	}
	conclusion(closestHoliday, holidayLength, dateRange, isToday)
}

func defineHolidayLength(holidayDate time.Time) (float64, string) {
	var holidayStart time.Time
	var holidayEnd time.Time
	var holidayDays float64
	var dateRange string

	holidayWeekday := holidayDate.Weekday()

	switch holidayWeekday.String() {
	case "Monday":
		holidayStart = holidayDate.AddDate(0, 0, -2)
		holidayEnd = holidayDate.AddDate(0, 0, 1)
		dateRange, holidayDays = calcDuration(holidayStart, holidayEnd)
	case "Friday":
		holidayStart = holidayDate
		holidayEnd = holidayDate.AddDate(0, 0, 3)
		dateRange, holidayDays = calcDuration(holidayStart, holidayEnd)
	case "Saturday":
		holidayStart = holidayDate
		holidayEnd = holidayDate.AddDate(0, 0, 3)
		dateRange, holidayDays = calcDuration(holidayStart, holidayEnd)
	case "Sunday":
		holidayStart = holidayDate.AddDate(0, 0, -1)
		holidayEnd = holidayDate.AddDate(0, 0, 2)
		dateRange, holidayDays = calcDuration(holidayStart, holidayEnd)
	default:
		holidayStart = holidayDate
		holidayEnd = holidayDate.AddDate(0, 0, 1)
		dateRange, holidayDays = calcDuration(holidayStart, holidayEnd)
	}
	return holidayDays, dateRange
}

func calcDuration(start, end time.Time) (string, float64) {
	const hoursInADay = 24

	dateRange := fmt.Sprintf("From %v %v to %v %v", start.Day(), start.Month(), end.Day(), end.Month())
	holidayDays := end.Sub(start).Hours() / hoursInADay

	return dateRange, holidayDays
}

func conclusion(holiday holiday, lengthInDays float64, dateRange string, isToday bool) {
	todayOrNot := "The closest holiday"
	dayOrDays := "days"

	if isToday {
		todayOrNot = "Today"
	}
	if lengthInDays == 1 {
		dayOrDays = "day"
	}

	output := fmt.Sprintf("%v is a %v on %v. It will last %v %v: %v.", todayOrNot, holiday.Name, holiday.Date, lengthInDays, dayOrDays, dateRange)

	fmt.Println(output)
}
