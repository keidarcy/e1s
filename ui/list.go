package ui

import (
	"os"
	"time"
)

// Show new page from LIST(like logs, events) content in table area and handle done event to go back
func (v *View) showListPages(entity Entity, which string) {
	contentString := v.getListString(entity, which)
	v.handleContentPageSwitch(entity, which, contentString)
}

// Based on current entity return list string as content
func (v *View) getListString(entity Entity, which string) string {
	contentString := ""
	tz := os.Getenv("TZ")
	currentTz, _ := time.LoadLocation(tz)

	switch {
	case which == "events":
		for _, e := range entity.events {
			createdAt := e.CreatedAt.In(currentTz)
			contentString += createdAt.Format(time.RFC3339) + ": " + *e.Message + "\n"
		}
		// add logs
		// case which == "logs"
	}

	return contentString
}
