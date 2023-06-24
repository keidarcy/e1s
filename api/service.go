package api

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// aws ecs list-services --cluster ${cluster}
// aws ecs describe-services --cluster ${cluster} --services ${service}
func (store *Store) ListServices(clusterName *string) ([]types.Service, error) {
	params := &ecs.ListServicesInput{
		Cluster: clusterName,
	}

	r, err := store.ecs.ListServices(context.Background(), params)

	if err != nil {
		logger.Printf("aws failed to list services, err: %v\n", err)
		return []types.Service{}, err
	}
	if len(r.ServiceArns) == 0 {
		return nil, nil
	}

	include := []types.ServiceField{
		types.ServiceFieldTags,
	}

	describeServicesOutput, err := store.ecs.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Services: r.ServiceArns,
		Cluster:  clusterName,
		Include:  include,
	})
	if err != nil {
		logger.Printf("aws failed to describe services, err: %v\n", err)
		return []types.Service{}, err
	}

	sort.Slice(describeServicesOutput.Services, func(i, j int) bool {
		return describeServicesOutput.Services[i].DesiredCount > describeServicesOutput.Services[j].DesiredCount
	})

	return describeServicesOutput.Services, nil
}

// Equivalent to
// aws ecs update-service --cluster ${cluster} --service ${service} --desired-count ${count} --force-new-deployment
func (store *Store) UpdateService(input *ecs.UpdateServiceInput) (*types.Service, error) {
	logger.Printf("cluster: %s, service: %s, desiredCount: %d, taskDef: %s, force: %t\n", *input.Cluster, *input.Service, *input.DesiredCount, *input.TaskDefinition, input.ForceNewDeployment)
	updateOutput, err := store.ecs.UpdateService(context.Background(), input)

	if err != nil {
		logger.Printf("aws failed to update service, err: %v\n", err)
		return nil, err
	}
	return updateOutput.Service, nil
}
