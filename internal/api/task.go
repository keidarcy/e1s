package api

import (
	"context"
	"fmt"
	"log/slog"

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
// The bool is true only when the running list was empty and results are stopped tasks
// from a cluster-wide list (so the UI can warn). It is false when the user asked for stopped tasks directly.
func (store *Store) ListTasks(clusterName, serviceName *string, status types.DesiredStatus) ([]types.Task, bool, error) {
	limit := int32(100)
	resultTasks := []types.Task{}
	describeTasksInclude := []types.TaskField{
		types.TaskFieldTags,
	}
	listTaskServiceName := serviceName

	// When listing stopped tasks, ECS ignores service on ListTasks; we filter after DescribeTasks.
	filterDescribedByService := status == types.DesiredStatusStopped
	warnFallbackFromRunningToStopped := false

	if status == types.DesiredStatusStopped {
		listTaskServiceName = nil
	}

	listTasksOutput, err := store.ecs.ListTasks(context.Background(), &ecs.ListTasksInput{
		Cluster:       clusterName,
		ServiceName:   listTaskServiceName,
		DesiredStatus: status,
		MaxResults:    &limit,
	})

	if err != nil {
		slog.Warn("failed to run aws api to list tasks", "error", err)
		return []types.Task{}, false, err
	}

	if status == types.DesiredStatusStopped && len(listTasksOutput.TaskArns) == 0 {
		return nil, false, nil
	}

	if status == types.DesiredStatusRunning && len(listTasksOutput.TaskArns) == 0 {
		listTasksOutput, err = store.ecs.ListTasks(context.Background(), &ecs.ListTasksInput{
			Cluster:       clusterName,
			DesiredStatus: types.DesiredStatusStopped,
			MaxResults:    &limit,
		})
		if err != nil {
			slog.Warn("failed to run aws api to list tasks", "error", err)
			return nil, false, err
		}
		if len(listTasksOutput.TaskArns) == 0 {
			return nil, false, nil
		}
		filterDescribedByService = true
		warnFallbackFromRunningToStopped = true
	}

	describeTasksOutput, err := store.ecs.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
		Cluster: clusterName,
		Tasks:   listTasksOutput.TaskArns,
		Include: describeTasksInclude,
	})

	if err != nil {
		slog.Warn("failed to run aws api to describe tasks", "error", err)
		return []types.Task{}, false, err
	}

	if len(describeTasksOutput.Tasks) > 0 {
		if !filterDescribedByService {
			resultTasks = append(resultTasks, describeTasksOutput.Tasks...)
		} else {
			for _, t := range describeTasksOutput.Tasks {
				if serviceName != nil {
					if *serviceName == utils.GetServiceByTaskGroup(t.Group) {
						resultTasks = append(resultTasks, t)
					}
				} else {
					resultTasks = append(resultTasks, t)
				}
			}
		}
	}

	return resultTasks, warnFallbackFromRunningToStopped, nil
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
