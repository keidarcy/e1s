package api

import (
	"context"
	"math"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// aws ecs list-services --cluster ${cluster}
// aws ecs describe-services --cluster ${cluster} --services ${service}
func (store *Store) ListServices(clusterName *string) ([]types.Service, error) {
	limit := int32(100)
	params := &ecs.ListServicesInput{
		Cluster:    clusterName,
		MaxResults: &limit,
	}

	listServicesOutput, err := store.ecs.ListServices(context.Background(), params)

	if err != nil {
		logger.Printf("e1s - aws failed to list services, err: %v\n", err)
		return []types.Service{}, err
	}

	if len(listServicesOutput.ServiceArns) == 0 {
		return nil, nil
	}

	include := []types.ServiceField{
		types.ServiceFieldTags,
	}

	results := []types.Service{}

	// You may specify up to 10 services to describe
	// If over 10, loop and slice by 10
	for i := 0; i <= len(listServicesOutput.ServiceArns)/10; i++ {
		describeServicesOutput, err := store.ecs.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
			Services: listServicesOutput.ServiceArns[i*10 : int(math.Min(float64((i+1)*10), float64(len(listServicesOutput.ServiceArns))))],
			Cluster:  clusterName,
			Include:  include,
		})
		if err != nil {
			logger.Printf("e1s - aws failed to describe services, err: %v\n", err)
			return []types.Service{}, err
		}
		results = append(results, describeServicesOutput.Services...)
	}

	// sort by desire count, name ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].DesiredCount > results[j].DesiredCount {
			return true
		} else if results[i].DesiredCount < results[j].DesiredCount {
			return false
		} else {
			return *results[i].ServiceName < *results[j].ServiceName
		}
	})

	return results, nil
}

// Equivalent to
// aws ecs update-service --cluster ${cluster} --service ${service} --task-definition ${task-definition} --desired-count ${count} --force-new-deployment
func (store *Store) UpdateService(input *ecs.UpdateServiceInput) (*types.Service, error) {
	logger.Printf("cluster: %s, service: %s, desiredCount: %d, taskDef: %s, force: %t\n", *input.Cluster, *input.Service, *input.DesiredCount, *input.TaskDefinition, input.ForceNewDeployment)
	updateOutput, err := store.ecs.UpdateService(context.Background(), input)

	if err != nil {
		logger.Printf("e1s - aws failed to update service, err: %v\n", err)
		return nil, err
	}
	return updateOutput.Service, nil
}
