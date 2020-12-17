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

	"github.com/jdvober/gauth"
	"github.com/jdvober/gsheets"
	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("Error loading .env")
	}
}

type makeFileArgs struct {
	client                  *http.Client
	d                       int
	days                    []string
	studentTemplate         string
	courseTemplate          string
	mod                     string
	googleSheetsData        [][][]interface{}
	spreadsheetIDAttendance string
	course                  string
}

func main() {
	// Add goroutines for speed later!

	// Load the template for a student that will later be copied and find/replaced.
	st, err := ioutil.ReadFile("./studentTemplate/student-template.json")
	if err != nil {
		fmt.Println(err)
	}
	StudentTemplate := string(st)

	// Assign environment vars from file
	nDays, err := strconv.Atoi(os.Getenv("NUM_OF_DAYS"))
	if err != nil {
		fmt.Println(err)
	}

	var allDays []string = []string{os.Getenv("MON"), os.Getenv("TUE"), os.Getenv("WED"), os.Getenv("THU"), os.Getenv("FRI")}
	var days []string = make([]string, nDays)
	for d := 0; d < nDays; d++ {
		days[d] = allDays[d]
	}

	// Get a list of all courses by CourseID
	var allCourseIDs []string = []string{os.Getenv("COURSEID1"), os.Getenv("COURSEID2"), os.Getenv("COURSEID3"), os.Getenv("COURSEID4"), os.Getenv("COURSEID5"), os.Getenv("COURSEID6"), os.Getenv("COURSEID7"), os.Getenv("COURSEID8"), os.Getenv("COURSEID9"), os.Getenv("COURSEID10")}

	// Get a list of all courses by CourseID
	var allShouldPost []string = []string{os.Getenv("POST_COURSE_1"), os.Getenv("POST_COURSE_2"), os.Getenv("POST_COURSE_3"), os.Getenv("POST_COURSE_4"), os.Getenv("POST_COURSE_5"), os.Getenv("POST_COURSE_6"), os.Getenv("POST_COURSE_7"), os.Getenv("POST_COURSE_8"), os.Getenv("POST_COURSE_9"), os.Getenv("POST_COURSE_10")}

	SpreadsheetIDRoster := os.Getenv("SSID_ROSTER")
	SpreadsheetIDAttendance := os.Getenv("SSID_ATTENDANCE")

	// Get Google Client for Oauth for Google Sheets
	client := gauth.Authorize()

	// Get the student IDs from the Class Roster 2.0 Spreadsheet Sunguard
	fmt.Println("Getting studentIDs")
	studentIDs := gsheets.GetValues(client, SpreadsheetIDRoster, "Master!I2:I")

	// Get the student's courses from the Class Roster 2.0 Spreadsheet Sunguard
	fmt.Println("Getting courses")
	courses := gsheets.GetValues(client, SpreadsheetIDRoster, "Master!T2:T")

	// Get attendance data
	fmt.Println("Getting attendanceVals")
	attendanceVals := gsheets.GetValues(client, SpreadsheetIDAttendance, "All Classes!E2:I")

	// Get the total minutes each student worked last week
	fmt.Println("Getting totalMins")
	totalMins := gsheets.GetValues(client, SpreadsheetIDAttendance, "All Classes!D2:D")

	// Get the full name of the students
	fmt.Println("Getting Last, First Middle")
	lastCommaFirstMiddles := gsheets.GetValues(client, SpreadsheetIDRoster, "Master!U2:U")

	// Get the mod of the students
	fmt.Println("Getting Mods")
	mods := gsheets.GetValues(client, SpreadsheetIDRoster, "Master!F2:F")

	// Get the grades of the students
	fmt.Println("Getting Grade Levels")
	gradeLevels := gsheets.GetValues(client, SpreadsheetIDRoster, "Master!E2:E")

	googleSheetsData := [][][]interface{}{studentIDs, courses, attendanceVals, totalMins, lastCommaFirstMiddles, mods, gradeLevels}

	// Get filenames of each file in ./courseTemplates.  Should be named COURSEID-template.json etc
	files, err := ioutil.ReadDir("./courseTemplates")
	if err != nil {
		fmt.Println(err)
	}
	var openList []string
	for _, file := range files {
		openList = append(openList, file.Name())
	}
	fmt.Printf("Opening files %s\n", openList)
	for f := range files {

		// Skips courses that are not on APEX
		if allShouldPost[f] == "true" {
			// Loop over each day
			for d := 0; d < nDays; d++ {
				// Open the template request as a byte array
				ct, err := ioutil.ReadFile("./courseTemplates/" + allCourseIDs[f] + ".json")
				if err != nil {
					fmt.Println(err)
				}
				courseTemplate := string(ct)

				// Get the mod number
				r := regexp.MustCompile(`"AttendancePeriods":"[0-9]{2}`)
				m := r.FindString(courseTemplate)
				mod := strings.SplitAfter(m, ":\"")

				args := makeFileArgs{
					client:                  client,
					d:                       d,
					days:                    days,
					studentTemplate:         StudentTemplate,
					courseTemplate:          courseTemplate,
					mod:                     mod[1],
					googleSheetsData:        googleSheetsData,
					spreadsheetIDAttendance: SpreadsheetIDAttendance,
					course:                  allCourseIDs[f],
				}

				makeFile(args)

				// Post to sunguard if necessary
				switch os.Getenv("POST_TO_SUNGUARD") {
				case "true", "True", "TRUE":
					postToSunguard(days[d], mod, args.course)
				case "false", "False", "FALSE":
					fmt.Printf("Not posting to Sunguard.\n\n\n")
				}
			}
		}
	}
}

func makeFile(args makeFileArgs) {
	studentIDs := args.googleSheetsData[0]
	courses := args.googleSheetsData[1]
	totalMins := args.googleSheetsData[3]
	attendanceVals := calcAttendance(args.client, totalMins)
	lastCommaFirstMiddles := args.googleSheetsData[4]
	gradeLevels := args.googleSheetsData[6]
	p := 0
	q := 1

	allStudents := []string{}

	// Make string for each student
	updatedStudentTemplate := args.studentTemplate
	r := regexp.MustCompile(`"SectionKey":[0-9]{5}`)
	sk := r.FindString(args.courseTemplate)
	sectionKey := strings.SplitAfter(sk, ":")

	for j, studentID := range studentIDs {
		// if student is in course based on matching Course from ss to course on list
		// This is to only add students that are actually in the course we are generating a file for, and not students from other classes
		if courses[j][0].(string) == args.course {
			// Fill in information to studentData

			// Replace Section Key
			r := regexp.MustCompile(`[\$]SECTIONKEY[\$]`)
			studentData := r.ReplaceAllString(updatedStudentTemplate, sectionKey[1])

			// Replace Student Id
			r = regexp.MustCompile(`[\$]STUDENTID[\$]`)
			studentData = r.ReplaceAllString(studentData, studentID[0].(string))

			// Replace Full Name
			lastCommaFirstMiddle := lastCommaFirstMiddles[j][0].(string)
			r = regexp.MustCompile(`[\$]LASTCOMMAFIRSTMIDDLE[\$]`)
			studentData = r.ReplaceAllString(studentData, lastCommaFirstMiddle)

			// Replace Course
			r = regexp.MustCompile(`[\$]COURSE[\$]`)
			studentData = r.ReplaceAllString(studentData, args.course)

			// Replace Grade
			r = regexp.MustCompile(`[\$]GRADELEVEL[\$]`)
			studentData = r.ReplaceAllString(studentData, gradeLevels[j][0].(string))

			// Replace Grade
			r = regexp.MustCompile(`[\$]GRADELEVEL[\$]`)
			studentData = r.ReplaceAllString(studentData, gradeLevels[j][0].(string))

			// Replace P
			r = regexp.MustCompile(`[\$]P[\$]`)
			studentData = r.ReplaceAllString(studentData, strconv.Itoa(p))

			// Replace Q
			r = regexp.MustCompile(`[\$]Q[\$]`)
			studentData = r.ReplaceAllString(studentData, strconv.Itoa(q))

			// Replace Student Name For Sort
			// Example: Vober, Joseph Daniel
			studentNameForSort := lastCommaFirstMiddle
			studentNameForSort = strings.ReplaceAll(studentNameForSort, ",", "#") // --> "Vober# Joseph Daniel"
			studentNameForSort = strings.ReplaceAll(studentNameForSort, " ", "+") // --> "Vober#+Joseph+Daniel"
			studentNameForSort = strings.ToUpper(studentNameForSort)              // --> "VOBER#+JOSEPH+DANIEL"
			r = regexp.MustCompile(`[\$]STUDENTNAMEFORSORT[\$]`)
			studentData = r.ReplaceAllString(studentData, studentNameForSort)

			// Replace Attendance based on return of CalcAttendance() (which is stored in attendanceVals)
			var absentBool string = "false"
			var presentBool string = "true"
			if attendanceVals[j][args.d] == "Absent" {
				absentBool = "true"
				presentBool = "false"
			}
			fmt.Printf("Attendance was listed as %q for %s\n", attendanceVals[j][args.d], studentNameForSort)
			fmt.Printf("Setting absentBool to %q \tand presentBool to %q\tfor %s.\n\n", absentBool, presentBool, args.days[args.d])

			r = regexp.MustCompile(`[\$]ABSENTBOOL[\$]`)
			studentData = r.ReplaceAllString(studentData, absentBool)
			r = regexp.MustCompile(`[\$]PRESENTBOOL[\$]`)
			studentData = r.ReplaceAllString(studentData, presentBool)

			allStudents = append(allStudents, studentData)
			p++
			q++

		}
	}
	// Replace $STUDENTS$ with all of the student information

	allStudentsFinal := strings.Join(allStudents, ",")
	r = regexp.MustCompile(`[\$]STUDENTS[\$]`)
	dataFinal := r.ReplaceAllString(args.courseTemplate, allStudentsFinal)
	r = regexp.MustCompile(`[\$]DATE[\$]`)
	dataFinal = r.ReplaceAllString(dataFinal, args.days[args.d])

	var filename string = args.days[args.d] + "_" + args.course + ".txt"

	if _, err := os.Stat("./txt/" + args.course + "/" + args.days[args.d]); os.IsNotExist(err) {
		fmt.Println("./txt/" + args.course + "/" + args.days[args.d] + " does not exist. MkdirAll will create ./txt/" + args.course + "/" + args.days[args.d] + "/")
		os.MkdirAll("./txt/"+args.course+"/"+args.days[args.d], 0777)
	}
	file, err := os.Create("./txt/" + args.course + "/" + args.days[args.d] + "/" + filename)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Created file ./txt/%s/%s/%s\n", args.course, args.days[args.d], filename)

	l, err := file.WriteString(dataFinal)
	if err != nil {
		fmt.Printf("Problem writing %v bytes", l)
		log.Println(err)
		file.Close()
		return
	}
	err = file.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func postToSunguard(day string, mod []string, course string) {
	/* Post all days to Sunguard */

	var filename string = day + "_" + course + ".txt"
	postFile, err := ioutil.ReadFile("./txt/" + course + "/" + day + "/" + filename)
	if err != nil {
		log.Printf("Error reading file ./txt/%s/%s/%s when attempting to post to Sunguard.\n", course, day, filename)
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

		// Fill in the first days of the array with present, as many days as calculated
		for i := 0; i < numDaysPresent; i++ {
			attendance[t] = append(attendance[t], "Present")
		}
		// Fill in remainder of days as absent
		for i := numDaysPresent - 1; i < 5; i++ {
			attendance[t] = append(attendance[t], "Absent")
		}
	}
	// Returns a 2D array of ex:
	// [
	//	["Present", "Present", "Absent", "Absent", "Absent"], // Student 1 M T W R F
	// ]
	return attendance
}
