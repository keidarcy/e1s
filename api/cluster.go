package api

import (
	"context"
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
		logger.Printf("e1s - aws failed to list clusters, err: %v\n", err)
		return []types.Cluster{}, err
	}

	include := []types.ClusterField{
		types.ClusterFieldAttachments,
		types.ClusterFieldConfigurations,
		types.ClusterFieldSettings,
		types.ClusterFieldStatistics,
		types.ClusterFieldTags,
	}
	// describeInput := &ecs.DescribeClustersInput{
	// 	Clusters: clustersOutput.ClusterArns,
	// 	Include:  include,
	// }

	// results := []types.Cluster{}

	// describeClusterOutput, err := store.ecs.DescribeClusters(context.Background(), describeInput)
	// if err != nil {
	// 	logger.Printf("e1s - aws failed to describe clusters, err: %v\n", err)
	// 	return []types.Cluster{}, err
	// }

	// results = append(results, describeClusterOutput.Clusters...)

	// describe each cluster to accept Deny specific cluster policy
	//  {
	//         "Effect": "Deny",
	//         "Action": "ecs:DescribeClusters",
	//         "Resource": "arn:aws:ecs:*:*:cluster/<specific cluster>"
	//  }
	results := []types.Cluster{}
	for _, clusterArn := range clustersOutput.ClusterArns {
		describeInput := &ecs.DescribeClustersInput{
			Clusters: []string{clusterArn},
			Include:  include,
		}
		describeClusterOutput, err := store.ecs.DescribeClusters(context.Background(), describeInput)
		if err != nil {
			logger.Printf("e1s - aws failed to describe clusters, err: %v\n", err)
			continue
		}
		results = append(results, describeClusterOutput.Clusters...)
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
