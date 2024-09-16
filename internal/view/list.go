package view

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/utils"
)

const (
	logFmt = "[aqua::]%s[-:-:-]:%s\n"
)

// Show new page from LIST(like logs, events) content in table area and handle done event to go back
func (v *view) showListPages(entity Entity) {
	contentString := v.getListString(entity)
	v.handleSecondaryPageSwitch(entity, contentString, []byte{})
	v.handleHeaderPageSwitch(entity)
}

// Based on current entity return list string as content
func (v *view) getListString(entity Entity) string {
	contentString := ""
	tz := os.Getenv("TZ")
	currentTz, _ := time.LoadLocation(tz)

	switch v.app.secondaryKind {
	case ServiceEventsKind:
		if entity.service == nil {
			contentString += "[red::]No valid contents[-:-:-]"
		}
		for _, e := range entity.events {
			createdAt := e.CreatedAt.In(currentTz)
			contentString += fmt.Sprintf(logFmt, createdAt.Format(time.RFC3339), *e.Message)
		}
	case LogKind:
		var logs []string
		var err error

		switch v.app.kind {
		case ServiceKind:
			logs, err = v.app.Store.GetServiceLogs(entity.service.TaskDefinition)
		case TaskKind:
			taskId := utils.ArnToName(entity.task.TaskArn)
			logs, err = v.app.Store.GetLogStreamLogs(entity.task.TaskDefinitionArn, taskId, "")
		case ContainerKind:
			taskId := utils.ArnToName(v.app.task.TaskArn)
			containerName := entity.container.Name
			logs, err = v.app.Store.GetLogStreamLogs(v.app.task.TaskDefinitionArn, taskId, *containerName)
		}

		if err != nil {
			contentString += "[red::]No valid contents[-:-:-]"
			v.app.Notice.Warnf("failed to getListString")
		}

		if len(logs) == 0 || len(logs) == 1 {
			contentString += "[orange::]Empty logs[-:-:-]"
		} else {
			for _, log := range logs {
				contentString += log
			}
		}
	}

	return contentString
}

// Switch to selected service events JSON page
func (v *view) switchToServiceEventsList() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warnf("failed to switchToServiceEventsList")
		return
	}
	if v.app.kind != ServiceKind {
		return
	}
	v.showListPages(selected)
}

// Switch to selected service events JSON page
func (v *view) switchToLogsList() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warnf("failed to switchToLogsList")
		return
	}
	v.showListPages(selected)
}

func (v *view) realtimeAwsLog(entity Entity) {
	var tdArn *string
	var logGroup string
	var logStreamNames []string
	var canRealtime bool
	if entity.service != nil {
		tdArn = entity.service.TaskDefinition
	} else if entity.task != nil {
		tdArn = entity.task.TaskDefinitionArn
	} else if entity.container != nil {
		tdArn = v.app.service.TaskDefinition
	}
	if tdArn == nil {
		return
	}
	td, err := v.app.Store.DescribeTaskDefinition(tdArn)
	if err != nil {
		v.app.Notice.Warnf("failed to switchToLogsList")
		return
	}
	for _, c := range td.ContainerDefinitions {
		// if current container kind is not target container skip
		if v.app.kind == ContainerKind {
			if *entity.container.Name != *c.Name {
				continue
			}
		}

		// if current container has no log driver
		if c.LogConfiguration.LogDriver != ecsTypes.LogDriverAwslogs {
			continue
		}

		groupName := c.LogConfiguration.Options["awslogs-group"]
		// if current container log configuration has no awslogs-group
		if groupName == "" {
			continue
		}

		// if logGroup is empty, assign it, can realtime logs
		// or if groupName is the same with previous
		if logGroup == "" || logGroup == groupName {
			logGroup = groupName
			canRealtime = true

			// get log stream name
			streamPrefix := *c.Name
			if _, ok := c.LogConfiguration.Options["awslogs-stream-prefix"]; ok {
				streamPrefix = c.LogConfiguration.Options["awslogs-stream-prefix"]
			}
			taskId := utils.ArnToName(v.app.task.TaskArn)
			streamName := fmt.Sprintf("%s/%s/%s", streamPrefix, *c.Name, taskId)
			logStreamNames = append(logStreamNames, streamName)
		} else {
			canRealtime = false
		}
	}

	if canRealtime {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		bin, err := exec.LookPath(awsCli)
		if err != nil {
			v.app.Notice.Warnf("failed to find aws cli binary, error: %v", err)
			v.app.back()
		}
		args := []string{
			"logs",
			"tail",
			"--follow",
			logGroup,
		}

		if v.app.kind == TaskKind || v.app.kind == ContainerKind {
			if len(args) == 4 && len(logStreamNames) > 0 {
				args = append(args, "--log-stream-names")
				args = append(args, logStreamNames...)
			}
		}

		slog.Info("exec", "command", bin+" "+strings.Join(args, " "))

		v.app.Suspend(func() {
			v.app.isSuspended = true
			cmd := exec.Command(bin, args...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

			_, err = cmd.Stdout.Write([]byte(fmt.Sprintf(realtimeLogFmt, *v.app.cluster.ClusterName, *v.app.service.ServiceName, logGroup, utils.ShowArray(logStreamNames))))
			err = cmd.Run()

			// return signal
			signal.Stop(interrupt)
			close(interrupt)
			v.app.isSuspended = false
		})
	} else {
		v.app.Notice.Warn("invalid realtime logs")
	}
}
