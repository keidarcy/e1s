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
// aws ecs list-services --cluster ${cluster}
// aws ecs describe-services --cluster ${cluster} --services ${service}
func (store *Store) ListServices(clusterName *string) ([]types.Service, error) {
	limit := int32(100)
	params := &ecs.ListServicesInput{
		Cluster:    clusterName,
		MaxResults: &limit,
	}

	serviceARNs := []string{}

	for {
		listServicesOutput, err := store.ecs.ListServices(context.Background(), params)
		if err != nil {
			logger.Warn("failed to run aws api to list services", "error", err)
			// If first run failed return err
			if len(serviceARNs) == 0 {
				return []types.Service{}, err
			}
			continue
		}

		if len(listServicesOutput.ServiceArns) == 0 {
			return nil, nil
		}

		serviceARNs = append(serviceARNs, listServicesOutput.ServiceArns...)

		if listServicesOutput.NextToken != nil {
			params.NextToken = listServicesOutput.NextToken
		} else {
			break
		}
	}

	// You may specify up to 10 services to describe.
	// If there are > 10 services in the cluster, loop and slice by 10
	// to describe them in batches of <= 10.
	batchSize := 10
	serviceCount := len(serviceARNs)
	loopCount := serviceCount / batchSize

	// If the number of services is divisible by batchSize, it's necessary to loop one less
	// time to describe all services in batches of batchSize.
	// Otherwise, we'll attempt to describe an empty slice of services, which results in a
	// HTTP 400: InvalidParameterException: Services cannot be empty.
	if serviceCount%batchSize == 0 {
		loopCount = loopCount - 1
	}

	results := []types.Service{}
	g := new(errgroup.Group)

	for i := 0; i <= loopCount; i++ {
		i := i

		g.Go(func() error {
			services := serviceARNs[i*batchSize : int(math.Min(float64((i+1)*batchSize), float64(serviceCount)))]

			describeServicesOutput, err := store.ecs.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
				Services: services,
				Cluster:  clusterName,
				Include: []types.ServiceField{
					types.ServiceFieldTags,
				},
			})
			if err != nil {
				logger.Warn("failed to run aws api to describe services in i:%d times loop", "error", i, err)
				return err
			}

			results = append(results, describeServicesOutput.Services...)

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return []types.Service{}, err
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
	taskDefinition := "no task definition"
	if input.TaskDefinition != nil {
		taskDefinition = *input.TaskDefinition
	}

	desiredCount := -1
	if input.DesiredCount != nil {
		desiredCount = int(*input.DesiredCount)
	}

	logger.Info("update service",
		slog.Group("parameters",
			slog.String("cluster", *input.Cluster),
			slog.String("service", *input.Service),
			slog.Int("desiredCount", desiredCount),
			slog.String("taskDefinition", taskDefinition),
			slog.Bool("forceNewDeployment", input.ForceNewDeployment),
		),
	)

	updateOutput, err := store.ecs.UpdateService(context.Background(), input)
	if err != nil {
		logger.Warn("failed to run aws api to update service", "error", err)
		return nil, err
	}
	return updateOutput.Service, nil
}
