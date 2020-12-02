package testNumDaysAbsent

import (
	"fmt"
)

func calcNumDaysAbsent() {
	// get attendance AllClasses!O2:O
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

	var cutoffs []int = []int{20, 40, 60, 80, 100}
	/* var totalHours int = 150 */
	fmt.Scanf("totalHours = ", &totalHours)
	numDaysPresent := 0

	for _, cutoff := range cutoffs {
		if totalHours < cutoff {
			/* numDaysPresent = a */
			break
		} else {
			numDaysPresent++
		}
	}
	fmt.Printf("\n\n\n\nHypothetical calculation of number of days to mark absent for %d hours\n\n", totalHours)
	fmt.Printf("Number of total hours:%d\nNumber of days to mark absent:%d\n", totalHours, numDaysPresent)
}
