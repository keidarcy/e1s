package api

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// ListContainerInstances gets container instances in an ECS cluster
// Equivalent to:
// aws ecs list-container-instances --cluster ${cluster}
// aws ecs describe-container-instances --cluster ${cluster} --container-instances ${instance1} ${instance2}
func (store *Store) ListContainerInstances(cluster *string) ([]types.ContainerInstance, error) {
	batchSize := 100
	limit := int32(batchSize)
	params := &ecs.ListContainerInstancesInput{
		Cluster:    cluster,
		MaxResults: &limit,
	}

	listOutput, err := store.ecs.ListContainerInstances(context.Background(), params)
	if err != nil {
		slog.Warn("failed to run aws api to list container instances", "error", err)
		return []types.ContainerInstance{}, err
	}

	// If no instances found, return empty slice
	if len(listOutput.ContainerInstanceArns) == 0 {
		return []types.ContainerInstance{}, nil
	}

	// Get detailed information about the container instances
	describeOutput, err := store.ecs.DescribeContainerInstances(context.Background(), &ecs.DescribeContainerInstancesInput{
		Cluster:            cluster,
		ContainerInstances: listOutput.ContainerInstanceArns,
	})

	if err != nil {
		slog.Warn("failed to run aws api to describe container instances", "error", err)
		return []types.ContainerInstance{}, err
	}

	return describeOutput.ContainerInstances, nil
}
