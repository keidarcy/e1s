package api

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// aws ecs list-tasks --cluster ${cluster} --service ${service}
// aws ecs describe-tasks --cluster ${cluster} --tasks ${taskID}
func (store *Store) ListTasks(clusterName, serviceName *string) ([]types.Task, error) {
	listTasksOutput, err := store.ecs.ListTasks(context.Background(), &ecs.ListTasksInput{
		Cluster:     clusterName,
		ServiceName: serviceName,
	})
	if err != nil {
		logger.Printf("aws failed to list tasks, err: %v\n", err)
		return []types.Task{}, err
	}
	if len(listTasksOutput.TaskArns) == 0 {
		return nil, nil
	}

	include := []types.TaskField{
		types.TaskFieldTags,
	}

	describeTasksOutput, err := store.ecs.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
		Cluster: clusterName,
		Tasks:   listTasksOutput.TaskArns,
		Include: include,
	})
	if err != nil {
		logger.Printf("aws failed to describe tasks, error: %v\n", err)
		return []types.Task{}, err
	}

	// sort tasks by task name
	sort.Slice(describeTasksOutput.Tasks, func(i, j int) bool {
		return *describeTasksOutput.Tasks[i].TaskArn > *describeTasksOutput.Tasks[j].TaskArn
	})

	// sort containers by health status
	for _, t := range describeTasksOutput.Tasks {
		sort.Slice(t.Containers, func(i, j int) bool {
			return t.Containers[i].HealthStatus < t.Containers[j].HealthStatus
		})
	}

	return describeTasksOutput.Tasks, nil
}
