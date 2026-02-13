package view

import (
	"testing"
	"time"
)

func TestCompareCellValues_ClusterTasksColumn(t *testing.T) {
	// [#458588]0 Pending[-] | [#98971a]18 Running
	a := "[#458588]0 Pending[-] | [#98971a]18 Running"
	b := "[#458588]0 Pending[-] | [#98971a]5 Running"

	t.Run("asc order - smaller running count first", func(t *testing.T) {
		// a=18 running, b=5 running; asc wants 5 before 18, so a should not be before b
		got := CompareCellValues(a, b, "Tasks", true, "asc")
		if got {
			t.Errorf("CompareCellValues(18, 5, asc) = %v, want false (5 before 18)", got)
		}
	})

	t.Run("desc order - larger running count first", func(t *testing.T) {
		// a=18 running, b=5 running; desc wants 18 before 5, so a should be before b
		got := CompareCellValues(a, b, "Tasks", true, "desc")
		if !got {
			t.Errorf("CompareCellValues(18, 5, desc) = %v, want true (18 before 5)", got)
		}
	})

	t.Run("equal running counts", func(t *testing.T) {
		same := "[#458588]0 Pending[-] | [#98971a]10 Running"
		got := CompareCellValues(same, same, "Tasks", true, "asc")
		if got {
			t.Errorf("CompareCellValues(10, 10, asc) = %v, want false", got)
		}
	})
}

func TestCompareCellValues_Ages(t *testing.T) {
	// Age: asc = newest first (smaller duration), desc = oldest first (larger duration)
	newer := "5m ago"
	older := "2h ago"

	t.Run("asc - newest first", func(t *testing.T) {
		got := CompareCellValues(newer, older, "Created", false, "asc")
		if !got {
			t.Errorf("CompareCellValues(5m, 2h, asc) = %v, want true (5m before 2h)", got)
		}
	})
	t.Run("desc - oldest first", func(t *testing.T) {
		got := CompareCellValues(newer, older, "Created", false, "desc")
		if got {
			t.Errorf("CompareCellValues(5m, 2h, desc) = %v, want false (2h before 5m)", got)
		}
	})
	t.Run("0s sorts first in asc", func(t *testing.T) {
		got := CompareCellValues("0s", "1m ago", "Created", false, "asc")
		if !got {
			t.Error("expected 0s before 1m ago in asc")
		}
	})
}

func TestCompareCellValues_Dates(t *testing.T) {
	early := "2024-01-01T00:00:00Z"
	late := "2024-12-31T23:59:59Z"

	t.Run("asc - earlier first", func(t *testing.T) {
		got := CompareCellValues(early, late, "Created", false, "asc")
		if !got {
			t.Errorf("CompareCellValues(early, late, asc) = %v, want true", got)
		}
	})

	t.Run("desc - later first", func(t *testing.T) {
		got := CompareCellValues(early, late, "Created", false, "desc")
		if got {
			t.Errorf("CompareCellValues(early, late, desc) = %v, want false", got)
		}
	})
}

func TestCompareCellValues_Numbers(t *testing.T) {
	t.Run("asc", func(t *testing.T) {
		if !CompareCellValues("10", "20", "Count", false, "asc") {
			t.Error("expected 10 before 20 in asc")
		}
	})
	t.Run("desc", func(t *testing.T) {
		if CompareCellValues("10", "20", "Count", false, "desc") {
			t.Error("expected 20 before 10 in desc")
		}
	})
}

func TestCompareCellValues_Strings(t *testing.T) {
	t.Run("asc", func(t *testing.T) {
		if !CompareCellValues("apple", "banana", "Name", false, "asc") {
			t.Error("expected apple before banana in asc")
		}
	})
	t.Run("desc", func(t *testing.T) {
		if CompareCellValues("apple", "banana", "Name", false, "desc") {
			t.Error("expected banana before apple in desc")
		}
	})
}

func TestCompareCellValues_ClusterTasksColumn_NonTasksHeader(t *testing.T) {
	// When isClusterTasks is true but header doesn't contain "tasks", we still use tasks logic
	// because we check header in CompareCellValues. So isClusterTasksColumn true + header "Tasks" = tasks column.
	// If header is "Name", we don't use tasks comparison. So we need isClusterTasksColumn to mean
	// "we are in cluster view" and the header check is for "tasks". So the param is "is cluster view and this column is tasks".
	// Looking at my implementation: we use tasks logic when `isClusterTasksColumn && strings.Contains(columnHeader, "tasks")`.
	// So when header is "Name", we won't use tasks logic. Good.
	a := "[#458588]0 Pending[-] | [#98971a]18 Running"
	b := "something else"
	// Column "Name" -> not tasks column, so falls through to string compare
	got := CompareCellValues(a, b, "Name", true, "asc")
	// String compare: "[" < "s" so a < b in asc -> true
	if !got {
		t.Errorf("CompareCellValues with Name header = %v, want true (string fallback)", got)
	}
}

func TestCompareCellValues_InvalidOrShortTasksFormat(t *testing.T) {
	// When there are fewer than 2 matches, we compare 0 vs 0 or 0 vs extracted
	a := "no brackets"
	b := "[#98971a]3 Running" // only one ](\d+)
	got := CompareCellValues(a, b, "Tasks", true, "asc")
	// a gives 0, b gives 3 (one match only, so len(aMatches) >= 2 is false -> 0). Actually b has one match: ]3. So len(bMatches) == 1. So bInt = 0. So 0 < 0 = false.
	if got {
		t.Errorf("CompareCellValues with short matches = %v, want false", got)
	}
}

func TestCompareCellValues_RealWorldTasksFormat(t *testing.T) {
	// Ensure real cluster task strings sort by running count
	rows := []string{
		"[#458588]0 Pending[-] | [#98971a]0 Running",
		"[#458588]0 Pending[-] | [#98971a]1 Running",
		"[#458588]2 Pending[-] | [#98971a]10 Running",
		"[#458588]0 Pending[-] | [#98971a]5 Running",
	}
	ascOrder := []int{0, 1, 3, 2} // by running count: 0, 1, 5, 10
	for i := 0; i < len(ascOrder)-1; i++ {
		before := rows[ascOrder[i]]
		after := rows[ascOrder[i+1]]
		if !CompareCellValues(before, after, "Tasks", true, "asc") {
			t.Errorf("expected %q before %q in asc", before, after)
		}
	}
}

func TestCompareCellValues_TimeFormat(t *testing.T) {
	// Only RFC3339 is treated as date; mixed with non-date string still uses date comparison for a
	valid := time.Now().UTC().Format(time.RFC3339)
	invalid := "2024-01-01 12:00:00"
	_ = CompareCellValues(valid, invalid, "Created", false, "asc")
}
