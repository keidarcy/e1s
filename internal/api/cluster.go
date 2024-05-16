package api

import (
	"context"
	"log/slog"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// aws ecs list-clusters
// aws ecs describe-clusters --clusters ${clusters}
func (store *Store) ListClusters() ([]types.Cluster, error) {
	limit := int32(100)
	clustersOutput, err := store.ecs.ListClusters(context.Background(), &ecs.ListClustersInput{
		MaxResults: &limit,
	})

	if err != nil {
		slog.Warn("failed to run aws api to list clusters", "error", err)
		return []types.Cluster{}, err
	}

	include := []types.ClusterField{
		types.ClusterFieldAttachments,
		types.ClusterFieldConfigurations,
		types.ClusterFieldSettings,
		types.ClusterFieldStatistics,
		types.ClusterFieldTags,
	}
	describeInput := &ecs.DescribeClustersInput{
		Clusters: clustersOutput.ClusterArns,
		Include:  include,
	}

	results := []types.Cluster{}

	describeClusterOutput, err := store.ecs.DescribeClusters(context.Background(), describeInput)
	if err != nil {
		slog.Warn("failed to run aws api to describe clusters", "error", err)
		return []types.Cluster{}, err
	}

	results = append(results, describeClusterOutput.Clusters...)

	// sort by running task count, name ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].RunningTasksCount > results[j].RunningTasksCount {
			return true
		} else if results[i].RunningTasksCount < results[j].RunningTasksCount {
			return false
		} else {
			return *results[i].ClusterName < *results[j].ClusterName
		}
	})

	return results, nil
}
