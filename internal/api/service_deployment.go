package api

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// aws ecs list-service-deployments --cluster ${cluster} --service ${service}
// aws ecs describe-service-deployments --service-deployment-arns ${arn1} ${arn2}
func (store *Store) ListServiceDeployments(cluster, service *string) ([]types.ServiceDeployment, error) {
	batchSize := 20
	limit := int32(batchSize)
	deploymentARNs := []string{}
	params := &ecs.ListServiceDeploymentsInput{
		Cluster:    cluster,
		Service:    service,
		MaxResults: &limit,
	}

	listServiceDeploymentsOutput, err := store.ecs.ListServiceDeployments(context.Background(), params)
	if err != nil {
		slog.Warn("failed to run aws api to list service deployments", "error", err)
		return []types.ServiceDeployment{}, err
	}

	for _, deployment := range listServiceDeploymentsOutput.ServiceDeployments {
		deploymentARNs = append(deploymentARNs, *deployment.ServiceDeploymentArn)
	}

	describeOutput, err := store.ecs.DescribeServiceDeployments(context.Background(), &ecs.DescribeServiceDeploymentsInput{
		ServiceDeploymentArns: deploymentARNs,
	})

	if err != nil {
		slog.Warn("failed to run aws api to describe service deployments", "error", err)
		return []types.ServiceDeployment{}, err
	}

	// // sort by running task count, name ascending
	// sort.Slice(results, func(i, j int) bool {
	// return results[i].StartedAt.After(u time.Time)
	// })

	return describeOutput.ServiceDeployments, nil
}

// Equivalent to
// aws ecs describe-service-revisions --service-revision-arns ${arn1}
func (store *Store) GetServiceRevision(serviceRevisionArn *string) (*types.ServiceRevision, error) {
	describeServiceRevisionOutput, err := store.ecs.DescribeServiceRevisions(context.Background(), &ecs.DescribeServiceRevisionsInput{
		ServiceRevisionArns: []string{*serviceRevisionArn},
	})
	if err != nil {
		slog.Warn("failed to run aws api to describe service revision", "error", err)
		return nil, err
	}
	return &describeServiceRevisionOutput.ServiceRevisions[0], nil
}
