package api

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogsTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// latest_log_stream=$(aws logs describe-log-streams \
//   --log-group-name "$log_group" \
//   --limit 1 \
//   --order-by "LastEventTime" \
//   --descending \
//   --query "logStreams[0].logStreamName" \
//   --output text)

// # Get the latest 10 log events from the log stream
//
//	aws logs get-log-events \
//	  --log-group-name "$log_group" \
//	  --log-stream-name "$latest_log_stream" \
//	  --limit 10 \
//	  --query "events[*].[timestamp,message]" \
//	  --output table
func (store *Store) GetLogs(tdArn *string) ([]cloudwatchlogsTypes.OutputLogEvent, error) {
	store.getCloudwatchlogsClient()
	td, err := store.DescribeTaskDefinition(tdArn)

	if err != nil {
		logger.Warnf("Failed to run aws api to describe task definition, err: %v", err)
		return nil, err
	}

	logs := []cloudwatchlogsTypes.OutputLogEvent{}
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
			logger.Warnf("Failed to run aws api to describe log stream, err: %v", err)
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
			logger.Warnf("Failed to run aws api to get log events, err: %v", err)
			continue
		}
		logs = append(logs, getLogEventsOutput.Events...)
	}
	return logs, nil
}
