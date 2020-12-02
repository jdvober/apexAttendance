package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	auth "github.com/jdvober/goGoogleAuth"
	ssVals "github.com/jdvober/goSheets/values"
	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("Error loading .env")
	}
}

/* // SpreadsheetIDRoster comes from the config file
 * var SpreadsheetIDRoster string = os.Getenv("SSID_ROSTER")
 *
 * // SpreadsheetIDAttendance comes from the config file
 * var SpreadsheetIDAttendance string = os.Getenv("SSID_ATTENDANCE") */

func main() {
	// Add goroutines for speed later!

	nDays, err := strconv.Atoi(os.Getenv("NUM_OF_DAYS"))
	if err != nil {
		fmt.Println(err)
	}
	nPeriods, err := strconv.Atoi(os.Getenv("NUM_OF_PERIODS"))
	if err != nil {
		fmt.Println(err)
	}
	/* mon := os.Getenv("MON")
	 * tue := os.Getenv("TUE")
	 * wed := os.Getenv("WED")
	 * thu := os.Getenv("THU")
	 * fri := os.Getenv("FRI") */
	var allDays []string = []string{os.Getenv("MON"), os.Getenv("TUE"), os.Getenv("WED"), os.Getenv("THU"), os.Getenv("FRI")}
	var days []string = make([]string, nDays)
	for d := 0; d < nDays; d++ {
		days[d] = allDays[d]
	}

	// Get a list of all Mods
	var allPeriods []string = []string{os.Getenv("PERIOD1"), os.Getenv("PERIOD2"), os.Getenv("PERIOD3"), os.Getenv("PERIOD4"), os.Getenv("PERIOD5"), os.Getenv("PERIOD6"), os.Getenv("PERIOD7"), os.Getenv("PERIOD8"), os.Getenv("PERIOD9")}
	var periods []string = make([]string, nPeriods)
	for p := 0; p < nPeriods; p++ {
		periods[p] = allPeriods[p]
	}

	SpreadsheetIDRoster := os.Getenv("SSID_ROSTER")
	SpreadsheetIDAttendance := os.Getenv("SSID_ATTENDANCE")

	// Get Google Client
	client := auth.Authorize()

	// Get the student IDs from the Class Roster 2.0 Spreadsheet Sunguard
	fmt.Println("Getting studentIDs")
	studentIDs := ssVals.Get(client, SpreadsheetIDRoster, "Master!I2:I")

	// Get attendance data
	fmt.Println("Getting attendanceVals")
	attendanceVals := ssVals.Get(client, SpreadsheetIDAttendance, "All Classes!E2:I")

	// Get the total hours each student worked last week
	fmt.Println("Getting totalMins")
	totalMins := ssVals.Get(client, SpreadsheetIDAttendance, "All Classes!D2:D")

	attendanceSheetData := [][][]interface{}{studentIDs, attendanceVals, totalMins}

	// Get filenames of each file in ./postTemplates.  Should be named mod03.json etc
	files, err := ioutil.ReadDir("./postTemplates")
	if err != nil {
		log.Fatal(err)
	}
	var openList []string
	for _, file := range files {
		openList = append(openList, file.Name())
	}
	fmt.Printf("Opening files %s\n", openList)
	for _, f := range files {
		// Loop over each day
		for d := 0; d < nDays; d++ {
			// Open the template request as a byte array
			src, err := ioutil.ReadFile("./postTemplates/" + f.Name())
			if err != nil {
				log.Printf("Error reading file ./postTemplates/%s.\n", f.Name())
				return
			}

			// Seperate original request into the data for each student
			data := strings.SplitAfterN(string(src), "\"Alerts\":{\"rowspan\":1}}},", -1)
			/* fmt.Println("Length of data: ", len(data)) */

			// Get the mod number
			r := regexp.MustCompile(`"AttendancePeriods":"[0-9][0-9]`)
			m := r.FindString(data[0])
			mod := strings.SplitAfter(m, ":\"")

			//
			makeFile(client, d, days, data, mod, attendanceSheetData, SpreadsheetIDAttendance)

			// Post to sunguard if necessary
			switch os.Getenv("POST_TO_SUNGUARD") {
			case "true", "True", "TRUE":
				postToSunguard(days[d], mod)
			case "false", "False", "FALSE":
				fmt.Println("Not posting to Sunguard")
			}
		}
	}
}

func makeFile(client *http.Client, d int, days []string, data []string, mod []string, attendanceSheetData [][][]interface{}, SpreadsheetIDAttendance string) {
	studentIDs := attendanceSheetData[0]
	/* attendanceVals := attendanceSheetData[1] */
	totalMins := attendanceSheetData[2]

	// Regex replace the date
	r := regexp.MustCompile(`[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]`)
	data[0] = r.ReplaceAllString(data[0], days[d])

	/* fmt.Printf("totalMins:\n%v\n", totalMins) */
	attendanceVals := calcAttendance(client, totalMins)
	/* for _, attendanceVal := range attendanceVals {
	 *     fmt.Printf("Test attendance slice: %v\n", attendanceVal)
	 * } */

	// ReplaceAll attendance data based on Absent or Present
	for i, requestSection := range data {
		// Check to see if that section contains the student ID
		for j, studentID := range studentIDs {
			// Check for blank values of StudentID from Class Roster
			/* if len(studentID[0].(string)) > 0 { */
			/* fmt.Printf("studentID[0].(string): %s\n\n", studentID[0].(string)) */

			if strings.Contains(requestSection, studentID[0].(string)) {
				/* attendanceStatus := attendanceVals[j][d].(string) */
				attendanceStatus := attendanceVals[j][d]
				/* if i == 97 { */
				/* fmt.Printf("Matched Student ID %s => Attendance: %s\n", studentID, attendanceStatus) */
				/* fmt.Printf("Before:\n") */
				/* fmt.Println(data[i]) */
				/* } */

				// Replace their attendance
				switch attendanceStatus {
				case "Absent":
					// Do absent stuff
					// if i == 97 {
					/* fmt.Println("(If) Absent:false --> Absent:true") */
					/* fmt.Println("(If) Present:true --> Present:false") */
					// }
					data[i] = strings.ReplaceAll(data[i], "\"Absent\":false", "\"Absent\":true")
					data[i] = strings.ReplaceAll(data[i], "\"Present\":true", "\"Present\":false")
				case "Present":
					// Do present stuff
					// if i == 97 {
					/* fmt.Println("(If) Absent:true --> Absent:false") */
					/* fmt.Println("(If) Present:false --> Present:true") */
					// }
					data[i] = strings.ReplaceAll(data[i], "\"Absent\":true", "\"Absent\":false")
					data[i] = strings.ReplaceAll(data[i], "\"Present\":false", "\"Present\":true")
				}
				/* if i == 97 { */
				/* fmt.Printf("\n\nAfter:\n") */
				/* fmt.Println(data[i]) */
				/* } */
			}
			/* } */
		}
	}
	// Rejoin all elements to a single string
	dataFinal := strings.Join(data, "")
	/* fmt.Printf("\n\n\ndataFinal:\n\n%s\n\n", dataFinal) */

	var filename string = days[d] + "_Mod_" + mod[1] + ".txt"

	f, err := os.Create("./txt/" + filename)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("\nCreated file ./%s\n", filename)

	l, err := f.WriteString(dataFinal)
	if err != nil {
		fmt.Printf("Problem writing %v bytes", l)
		log.Println(err)
		f.Close()
		return
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
		return
	}
}

func postToSunguard(day string, mod []string) {
	/* Post all days to Sunguard */

	var filename string = day + "_Mod_" + mod[1] + ".txt"
	postFile, err := ioutil.ReadFile("./txt/" + filename)
	if err != nil {
		log.Printf("Error reading file ./txt/%s when attempting to post to Sunguard.\n", filename)
		return
	}
	timeout := time.Duration(20 * time.Second)
	client2 := http.Client{
		Timeout: timeout,
	}

	// Load TACCookie environment variable from config.env
	// Look into Viper for future projects?
	TACCookie := os.Getenv("TAC_COOKIE")

	// Check for copy/paster error.  When copying from Firefox, it includes the "Cookie: " part of the cookie, and we want just the value!
	if strings.HasPrefix(TACCookie, "TACDistrict") {

		request, err := http.NewRequest("POST", "https://tac.sparcc.org/TAC/Attendance/SaveAttendance", bytes.NewBuffer(postFile))
		request.Header.Set("Host", "tac.sparcc.org")
		request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:83.0) Gecko/20100101 Firefox/83.0")
		request.Header.Add("Accept", "*/*")
		request.Header.Add("Accept-Language", "en-US,en;q=0.5")
		request.Header.Add("Accept-Encoding", "gzip, deflate, br")
		request.Header.Add("Content-Type", "application/json;charset=UTF-8")
		request.Header.Add("X-Requested-With", "XMLHttpRequest")
		request.Header.Add("Content-Length", "295321")
		request.Header.Add("Origin", "https://tac.sparcc.org")
		request.Header.Add("DNT", "1")
		request.Header.Add("Connection", "keep-alive")
		request.Header.Add("Referer", "https://tac.sparcc.org/TAC/Attendance/IndexList")
		request.Header.Add("Cookie", TACCookie)

		fmt.Printf("Attempting POST for %s :: Mod %s", day, mod[1])

		resp, err := client2.Do(request)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf(" << %s\n", resp.Status)
	} else {
		fmt.Println("Check your copy/paste for the value of TAC_COOKIE in the config.env file")
	}
}

func calcAttendance(client *http.Client, totalMins [][]interface{}) [][]string {
	// get attendance AllClasses!O2:O
	/* fmt.Printf("\nCalculating attendance\n") */

	// If course is AP Physics or Physics, set absent value to N/a
	// Divide number of hours from Column O by number of days fron config.env
	/*
		Total Hours  Number of days present
							each loop, increment i if i < nDays
		0-19                   0
		20-39                  1
		40-59                  2
		60-79                  3
		80-99                  4
		100+                   5
	*/

	var cutoffs []float64 = []float64{20, 40, 60, 80, 100}
	/* var attendance [][]string */
	attendance := make([][]string, len(totalMins), len(totalMins))

	// Calculate how many days they were present
	for t := range totalMins {

		/* fmt.Printf("totalMins[%v] = %v\n", t, totalMins[t]) */
		numDaysPresent := 0

		for _, cutoff := range cutoffs {
			// Convert interface {} string to a float to be compared for hours
			if s, err := strconv.ParseFloat(totalMins[t][0].(string), 32); err == nil {
				if s < cutoff {
					/* numDaysPresent = a */
					break
				} else {
					numDaysPresent++
				}
			}
		}
		/* fmt.Printf("Number of total hours:%f\nNumber of days to mark present:%d\n", totalMin, numDaysPresent) */

		// Fill in the first days of the array with present, as many days as calculated
		for i := 0; i < numDaysPresent; i++ {
			attendance[t] = append(attendance[t], "Present")
		}
		// Fill in remainder of days as absent
		for i := numDaysPresent - 1; i < 5; i++ {
			attendance[t] = append(attendance[t], "Absent")
		}
	}
	return attendance
}
