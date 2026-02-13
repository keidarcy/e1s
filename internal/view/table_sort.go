package view

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/keidarcy/e1s/internal/utils"
)

// CompareCellValues compares two cell values for table sorting.
// It is a pure function of its arguments for easy unit testing.
//
// Parameters:
//   - a, b: the two cell string values to compare
//   - columnHeader: the header text of the column (used to detect "tasks" column)
//   - isClusterTasksColumn: true when this is the cluster view's tasks column
//   - sortOrder: "asc" or "desc"
//
// Returns true when a should sort before b.
func CompareCellValues(a, b string, columnHeader string, isClusterTasksColumn bool, sortOrder string) bool {
	if isClusterTasksColumn && strings.Contains(strings.ToLower(columnHeader), "tasks") {
		return compareTasksColumn(a, b, sortOrder)
	}

	if utils.IsAge(a) {
		return compareAges(a, b, sortOrder)
	}

	if _, err := time.Parse(time.RFC3339, a); err == nil {
		return compareDates(a, b, sortOrder)
	}

	if _, err := strconv.Atoi(a); err == nil {
		return compareNumbers(a, b, sortOrder)
	}

	return compareStrings(a, b, sortOrder)
}

func compareTasksColumn(a, b string, sortOrder string) bool {
	re := regexp.MustCompile(`\](\d+)`)
	aMatches := re.FindAllStringSubmatch(a, -1)
	bMatches := re.FindAllStringSubmatch(b, -1)
	var aInt, bInt int
	if len(aMatches) >= 2 {
		aInt, _ = strconv.Atoi(aMatches[1][1])
	}
	if len(bMatches) >= 2 {
		bInt, _ = strconv.Atoi(bMatches[1][1])
	}
	if sortOrder == "asc" {
		return aInt < bInt
	}
	return aInt > bInt
}

// compareAges compares age strings (e.g. "5m ago", "0s") by duration.
// Asc = newest first (smaller duration), desc = oldest first (larger duration).
func compareAges(a, b string, sortOrder string) bool {
	aDur, okA := utils.ParseAge(a)
	bDur, okB := utils.ParseAge(b)
	if !okA {
		aDur = 0
	}
	if !okB {
		bDur = 0
	}
	if sortOrder == "asc" {
		return aDur < bDur
	}
	return aDur > bDur
}

func compareDates(a, b string, sortOrder string) bool {
	aTime, _ := time.Parse(time.RFC3339, a)
	bTime, _ := time.Parse(time.RFC3339, b)
	if sortOrder == "asc" {
		return aTime.Before(bTime)
	}
	return aTime.After(bTime)
}

func compareNumbers(a, b string, sortOrder string) bool {
	aInt, _ := strconv.Atoi(a)
	bInt, _ := strconv.Atoi(b)
	if sortOrder == "asc" {
		return aInt < bInt
	}
	return aInt > bInt
}

func compareStrings(a, b string, sortOrder string) bool {
	if sortOrder == "asc" {
		return a < b
	}
	return a > b
}
