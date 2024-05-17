package api

import (
	"context"
	"log/slog"
	"math"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"golang.org/x/sync/errgroup"
)

// Equivalent to
// aws ecs list-clusters
// aws ecs describe-clusters --clusters ${clusters}
func (store *Store) ListClusters() ([]types.Cluster, error) {
	batchSize := 100
	limit := int32(batchSize)
	clusterARNs := []string{}
	params := &ecs.ListClustersInput{
		MaxResults: &limit,
	}

	for {
		clustersOutput, err := store.ecs.ListClusters(context.Background(), params)
		if err != nil {
			slog.Warn("failed to run aws api to list clusters", "error", err)
			if len(clusterARNs) == 0 {
				return []types.Cluster{}, err
			}
			continue
		}
		clusterARNs = append(clusterARNs, clustersOutput.ClusterArns...)
		if clustersOutput.NextToken != nil {
			params.NextToken = clustersOutput.NextToken
		} else {
			break
		}
	}

	include := []types.ClusterField{
		types.ClusterFieldAttachments,
		types.ClusterFieldConfigurations,
		types.ClusterFieldSettings,
		types.ClusterFieldStatistics,
		types.ClusterFieldTags,
	}

	results := []types.Cluster{}
	g := new(errgroup.Group)

	clusterCount := len(clusterARNs)
	loopCount := clusterCount / batchSize

	if clusterCount%batchSize == 0 {
		loopCount = loopCount - 1
	}

	for i := 0; i <= loopCount; i++ {
		i := i
		g.Go(func() error {
			clusters := clusterARNs[i*batchSize : int(math.Min(float64((i+1)*batchSize), float64(clusterCount)))]

			// If describe more than 100, InvalidParameterException: Clusters cannot have more than 100 elements
			describeClusterOutput, err := store.ecs.DescribeClusters(context.Background(), &ecs.DescribeClustersInput{
				Clusters: clusters,
				Include:  include,
			})
			if err != nil {
				slog.Warn("failed to run aws api to describe clusters", "error", err)
				return err
			}

			results = append(results, describeClusterOutput.Clusters...)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return []types.Cluster{}, err
	}

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
