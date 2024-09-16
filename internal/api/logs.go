package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogsTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

const (
	logFmt = "[aqua::]%s[-:-:-]:%s\n"
)

// Equivalent to
// latest_log_stream=$(aws logs describe-log-streams \
//   --log-group-name "$log_group" \
//   --limit 1 \
//   --order-by "LastEventTime" \
//   --descending \
//   --query "logStreams[0].logStreamName" \
//   --output text)

// # Get the latest 100 log events from the log stream
//
//	aws logs get-log-events \
//	  --log-group-name "$log_group" \
//	  --log-stream-name "$latest_log_stream" \
//	  --limit 100 \
//	  --query "events[*].[timestamp,message]" \
//	  --output table
func (store *Store) GetServiceLogs(tdArn *string) ([]string, error) {
	store.initCloudwatchlogsClient()
	td, err := store.DescribeTaskDefinition(tdArn)

	if err != nil {
		slog.Warn("failed to run aws api to describe task definition", "error", err)
		return nil, err
	}

	logs := []string{}
	logGroupNames := make(map[string]bool)
	for _, c := range td.ContainerDefinitions {
		if c.LogConfiguration == nil {
			continue
		}
		if c.LogConfiguration.LogDriver != types.LogDriverAwslogs {
			continue
		}
		groupName := c.LogConfiguration.Options["awslogs-group"]
		if groupName == "" {
			continue
		}

		// avoid the same log group for different containers
		if _, ok := logGroupNames[groupName]; ok {
			continue
		}
		logGroupNames[groupName] = true

		describeLogStreamsInput := &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: &groupName,
			Limit:        aws.Int32(1),
			OrderBy:      cloudwatchlogsTypes.OrderByLastEventTime,
			Descending:   aws.Bool(true),
		}
		describeLogStreamsOutput, err := store.cloudwatchlogs.DescribeLogStreams(context.Background(), describeLogStreamsInput)
		if err != nil {
			slog.Warn("failed to run aws api to describe log stream", "error", err)
			continue
		}
		streamName := describeLogStreamsOutput.LogStreams[0].LogStreamName

		getLogEventsInput := &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  &groupName,
			LogStreamName: streamName,
			Limit:         aws.Int32(100),
		}
		getLogEventsOutput, err := store.cloudwatchlogs.GetLogEvents(context.Background(), getLogEventsInput)
		if err != nil {
			slog.Warn("failed to run aws api to get log events", "error", err)
			continue
		}
		if len(getLogEventsOutput.Events) == 0 {
			continue
		}
		logs = append(logs, fmt.Sprintf("Log stream name: %s\n", *streamName))
		for _, e := range getLogEventsOutput.Events {
			logs = append(logs, fmt.Sprintf(logFmt, time.Unix(0, *e.Timestamp*int64(time.Millisecond)).Format(time.RFC3339), *e.Message))
		}
	}
	return logs, nil
}

// Get the 50 log events for each container
// Equivalent to
//
//	aws logs get-log-events \
//	  --log-group-name "$log_group" \
//	  --log-stream-name "$stream_prefix/$container_name/$task_id" \
//	  --limit 50 \
//	  --query "events[*].[timestamp,message]" \
//	  --output table
func (store *Store) GetLogStreamLogs(tdArn *string, taskId string, containerName string) ([]string, error) {
	store.initCloudwatchlogsClient()
	td, err := store.DescribeTaskDefinition(tdArn)

	if err != nil {
		slog.Warn("failed to run aws api to describe task definition", "error", err)
		return nil, err
	}

	logs := []string{}
	for _, c := range td.ContainerDefinitions {
		if *c.Name != containerName && containerName != "" {
			continue
		}
		if c.LogConfiguration == nil {
			continue
		}
		if c.LogConfiguration.LogDriver != types.LogDriverAwslogs {
			continue
		}
		groupName := c.LogConfiguration.Options["awslogs-group"]
		if groupName == "" {
			continue
		}

		streamPrefix := *c.Name

		if _, ok := c.LogConfiguration.Options["awslogs-stream-prefix"]; ok {
			streamPrefix = c.LogConfiguration.Options["awslogs-stream-prefix"]
		}

		streamName := fmt.Sprintf("%s/%s/%s", streamPrefix, *c.Name, taskId)

		getLogEventsInput := &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  &groupName,
			LogStreamName: &streamName,
			Limit:         aws.Int32(50),
		}
		getLogEventsOutput, err := store.cloudwatchlogs.GetLogEvents(context.Background(), getLogEventsInput)
		if err != nil {
			slog.Warn("failed to run aws api to get log events", "error", err)
			continue
		}

		if len(getLogEventsOutput.Events) == 0 {
			continue
		}
		logs = append(logs, fmt.Sprintf("Log stream name: %s\n", streamName))
		for _, e := range getLogEventsOutput.Events {
			logs = append(logs, fmt.Sprintf(logFmt, time.Unix(0, *e.Timestamp*int64(time.Millisecond)).Format(time.RFC3339), *e.Message))
		}
	}
	return logs, nil
}
