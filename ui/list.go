package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

const (
	logFmt = "[aqua::]%s[-:-:-]:%s\n"
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

	switch which {
	case "events":
		if entity.service == nil {
			contentString += "[red::]No valid contents[-:-:-]"
		}
		for _, e := range entity.events {
			createdAt := e.CreatedAt.In(currentTz)
			contentString += fmt.Sprintf(logFmt, createdAt.Format(time.RFC3339), *e.Message)
		}
	case "logs":
		var logs []types.OutputLogEvent
		var err error

		if entity.service != nil {
			logs, err = v.app.Store.GetLogs(entity.service.TaskDefinition)
		} else if entity.task != nil {
			logs, err = v.app.Store.GetLogs(entity.task.TaskDefinitionArn)
		}

		if err != nil {
			contentString += "[red::]No valid contents[-:-:-]"
		}

		for _, log := range logs {
			m := log.Message
			contentString += fmt.Sprintf(logFmt, time.Unix(0, *log.Timestamp*int64(time.Millisecond)).Format(time.RFC3339), *m)
		}
	}

	return contentString
}
