package api

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/utils"
)

// Equivalent to
// aws ecs list-tasks --cluster ${cluster} --service ${service}
// OR
// aws ecs list-tasks --cluster ${cluster}
// aws ecs describe-tasks --cluster ${cluster} --tasks ${taskID}
// OR
// aws ecs list-tasks --cluster ${cluster} --desired-status STOPPED
// `aws ecs list-tasks --cluster ${CLUSTER} --service-name ${SERVICE} --desired-status STOPPED` return nothing
// `aws ecs list-tasks --cluster ${CLUSTER} --desired-status STOPPED` return all stopped tasks in cluster
func (store *Store) ListTasks(clusterName, serviceName *string, status types.DesiredStatus) ([]types.Task, bool, error) {
	limit := int32(100)
	resultTasks := []types.Task{}
	describeTasksInclude := []types.TaskField{
		types.TaskFieldTags,
	}
	listTaskServiceName := serviceName
	noRunningShowStopped := false

	// true when show desiredStatus:stopped tasks
	if status == types.DesiredStatusStopped {
		listTaskServiceName = nil
		noRunningShowStopped = true
	}

	listTasksOutput, _ := store.ecs.ListTasks(context.Background(), &ecs.ListTasksInput{
		Cluster:       clusterName,
		ServiceName:   listTaskServiceName,
		DesiredStatus: status,
		MaxResults:    &limit,
	})

	if status == types.DesiredStatusStopped && len(listTasksOutput.TaskArns) == 0 {
		return nil, noRunningShowStopped, nil
	}

	if status == types.DesiredStatusRunning && len(listTasksOutput.TaskArns) == 0 {
		listTasksOutput, _ := store.ecs.ListTasks(context.Background(), &ecs.ListTasksInput{
			Cluster:       clusterName,
			DesiredStatus: types.DesiredStatusStopped,
			MaxResults:    &limit,
		})
		if len(listTasksOutput.TaskArns) == 0 {
			return nil, noRunningShowStopped, nil
		}
		noRunningShowStopped = true
	}

	describeTasksOutput, err := store.ecs.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
		Cluster: clusterName,
		Tasks:   listTasksOutput.TaskArns,
		Include: describeTasksInclude,
	})

	if err != nil {
		slog.Warn("failed to run aws api to describe tasks", "error", err)
		return []types.Task{}, noRunningShowStopped, err
	}

	if len(describeTasksOutput.Tasks) > 0 {
		if !noRunningShowStopped {
			resultTasks = append(resultTasks, describeTasksOutput.Tasks...)
		} else {
			for _, t := range describeTasksOutput.Tasks {
				if *serviceName == utils.GetServiceByTaskGroup(t.Group) {
					resultTasks = append(resultTasks, t)
				}
			}
		}
	}

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

	return resultTasks, noRunningShowStopped, nil
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

// aws ecs stop-task --cluster ${cluster} --task ${taskId}
func (store *Store) StopTask(input *ecs.StopTaskInput) error {
	_, err := store.ecs.StopTask(context.Background(), input)
	if err != nil {
		return err
	}
	return nil
}

// aws ecs describe-container-instances --cluster ${cluster} --container-instances ${instanceId}
func (store *Store) GetTaskInstanceId(cluster, containerInstance *string) (string, error) {
	describeOutput, err := store.ecs.DescribeContainerInstances(context.Background(), &ecs.DescribeContainerInstancesInput{
		Cluster:            cluster,
		ContainerInstances: []string{*containerInstance},
	})

	if err != nil {
		slog.Warn("failed to run aws api to describe container instances", "error", err)
		return "", err
	}

	if len(describeOutput.ContainerInstances) != 1 {
		return "", fmt.Errorf("expect 1 container instance, got %d", len(describeOutput.ContainerInstances))
	}

	return *describeOutput.ContainerInstances[0].Ec2InstanceId, nil
}
