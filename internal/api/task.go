package api

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// aws ecs list-tasks --cluster ${cluster} --service ${service}
// OR
// aws ecs list-tasks --cluster ${cluster}
// aws ecs describe-tasks --cluster ${cluster} --tasks ${taskID}
// OR
// aws ecs list-tasks --cluster ${cluster} --desired-status STOPPED
func (store *Store) ListTasks(clusterName, serviceName *string, status types.DesiredStatus) ([]types.Task, error) {
	limit := int32(100)
	listTasksOutput, err := store.ecs.ListTasks(context.Background(), &ecs.ListTasksInput{
		Cluster:       clusterName,
		ServiceName:   serviceName,
		DesiredStatus: status,
		MaxResults:    &limit,
	})
	if err != nil {
		logger.Warnf("Failed to run aws api to list tasks, err: %v", err)
		return []types.Task{}, err
	}
	if len(listTasksOutput.TaskArns) == 0 {
		return nil, nil
	}

	include := []types.TaskField{
		types.TaskFieldTags,
	}

	resultTasks := []types.Task{}

	describeTasksOutput, err := store.ecs.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
		Cluster: clusterName,
		Tasks:   listTasksOutput.TaskArns,
		Include: include,
	})

	if err != nil {
		logger.Warnf("Failed to run aws api to describe tasks, error: %v", err)
		return []types.Task{}, err
	}

	resultTasks = append(resultTasks, describeTasksOutput.Tasks...)

	// sort tasks by task name
	sort.Slice(resultTasks, func(i, j int) bool {
		return *resultTasks[i].TaskArn > *resultTasks[j].TaskArn
	})

	// sort containers by health status
	for _, t := range resultTasks {
		sort.Slice(t.Containers, func(i, j int) bool {
			return t.Containers[i].HealthStatus < t.Containers[j].HealthStatus
		})
	}

	return resultTasks, nil
}

// aws ecs register-task-definition --family ${{family}} --...
// return registered task definition revision
func (store *Store) RegisterTaskDefinition(input *ecs.RegisterTaskDefinitionInput) (string, int32, error) {
	registeredTdOutput, err := store.ecs.RegisterTaskDefinition(context.Background(), input)
	if err != nil {
		return "", 0, err
	}
	family := *registeredTdOutput.TaskDefinition.Family
	revision := registeredTdOutput.TaskDefinition.Revision
	return family, revision, nil
}
