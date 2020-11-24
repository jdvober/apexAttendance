package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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
func main() {
	// Add goroutines for speed later!

	/* enverr := godotenv.Load("config.env")
	 * if enverr != nil {
	 *     log.Println("Error loading .env file")
	 * } */
	mon := os.Getenv("MON")
	tue := os.Getenv("TUE")
	wed := os.Getenv("WED")
	thu := os.Getenv("THU")
	fri := os.Getenv("FRI")
	var days []string = []string{mon, tue, wed, thu, fri}
	SpreadsheetIDRoster := os.Getenv("SSID_ROSTER")
	SpreadsheetIDAttendance := os.Getenv("SSID_ATTENDANCE")

	// Get Google Client
	client := auth.Authorize()

	// Get the student IDs from the Class Roster 2.0 Spreadsheet Sunguard
	studentIDs := ssVals.Get(client, SpreadsheetIDRoster, "Master!I2:I")

	// Get attendance data
	attendanceVals := ssVals.Get(client, SpreadsheetIDAttendance, "All Classes!E2:I")

	// Open the template request as a byte array
	src, err := ioutil.ReadFile("request.json")
	if err != nil {
		log.Println("Error reading file.")
		return
	}

	// Seperate original request into the data for each student
	data := strings.SplitAfterN(string(src), "\"Alerts\":{\"rowspan\":1}}},", -1)
	/* fmt.Println("Length of data: ", len(data)) */

	// Get the mod number
	r := regexp.MustCompile(`"AttendancePeriods":"[0-9][0-9]`)
	m := r.FindString(data[0])
	mod := strings.SplitAfter(m, ":\"")

	// Loop over each day
	for d := 0; d < 5; d++ {
		// Regex replace the date
		r = regexp.MustCompile(`[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]`)
		data[0] = r.ReplaceAllString(data[0], days[d])

		// ReplaceAll attendance data based on Absent or Present
		for i, requestSection := range data {
			// Check to see if that section contains the student ID
			for j, studentID := range studentIDs {
				// Check for blank values of StudentID from Class Roster
				/* if len(studentID[0].(string)) > 0 { */
				/* fmt.Printf("studentID[0].(string): %s\n\n", studentID[0].(string)) */
				if strings.Contains(requestSection, studentID[0].(string)) {
					attendanceStatus := attendanceVals[j][d].(string)
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

		f, err := os.Create(filename)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Created file ./%s\n", filename)

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

	/* Post all days to Sunguard */
	for i := range days {
		var filename string = days[i] + "_Mod_" + mod[1] + ".txt"
		postFile, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println("Error reading file when attempting to post to Sunguard.")
			return
		}
		timeout := time.Duration(20 * time.Second)
		client2 := http.Client{
			Timeout: timeout,
		}

		// Load TACCookie environment variable from config.env
		// Look into Viper for future projects?
		TACCookie := os.Getenv("TAC_COOKIE")

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

		fmt.Printf("Attempting POST for %s :: Mod %s", days[i], mod[1])

		resp, err := client2.Do(request)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf(" << %s\n", resp.Status)
	}
}
