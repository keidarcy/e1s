package api

import (
	"context"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
)

var logger = util.Logger

type Store struct {
	*aws.Config
	ecs         *ecs.Client
	cloudWatch  *cloudwatch.Client
	autoScaling *applicationautoscaling.Client
}

func NewStore() *Store {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		logger.Printf("unable to load SDK config, error: %v\n", err)
	}
	ecsClient := ecs.NewFromConfig(cfg)
	return &Store{
		Config: &cfg,
		ecs:    ecsClient,
	}
}

// Equivalent to
// aws ecs list-clusters
// aws ecs describe-clusters --clusters ${clusters}
func (store *Store) ListClusters() ([]types.Cluster, error) {
	clustersOutput, err := store.ecs.ListClusters(context.Background(), &ecs.ListClustersInput{})

	if err != nil {
		logger.Printf("aws failed to list clusters, err: %v\n", err)
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

	describeClusterOutput, err := store.ecs.DescribeClusters(context.Background(), describeInput)
	if err != nil {
		logger.Printf("aws failed to describe clusters, err: %v\n", err)
		return []types.Cluster{}, err
	}

	sort.Slice(describeClusterOutput.Clusters, func(i, j int) bool {
		return describeClusterOutput.Clusters[i].RunningTasksCount > describeClusterOutput.Clusters[j].RunningTasksCount
	})

	return describeClusterOutput.Clusters, nil
}
