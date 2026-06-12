package api

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Equivalent to
// aws ecs list-daemons --cluster ${cluster}
func (store *Store) ListDaemons(clusterArn *string) ([]types.DaemonSummary, error) {
	limit := int32(100)
	params := &ecs.ListDaemonsInput{
		ClusterArn: clusterArn,
		MaxResults: &limit,
	}

	results := []types.DaemonSummary{}

	for {
		output, err := store.ecs.ListDaemons(context.Background(), params)
		if err != nil {
			slog.Warn("failed to run aws api to list daemons", "error", err)
			if len(results) == 0 {
				return nil, err
			}
			break
		}

		results = append(results, output.DaemonSummariesList...)

		if output.NextToken != nil {
			params.NextToken = output.NextToken
		} else {
			break
		}
	}

	return results, nil
}

// Equivalent to
// aws ecs describe-daemon --daemon-arn ${daemonArn}
func (store *Store) DescribeDaemon(daemonArn *string) (*types.DaemonDetail, error) {
	output, err := store.ecs.DescribeDaemon(context.Background(), &ecs.DescribeDaemonInput{
		DaemonArn: daemonArn,
	})
	if err != nil {
		slog.Warn("failed to run aws api to describe daemon", "error", err)
		return nil, err
	}
	return output.Daemon, nil
}

// Equivalent to
// aws ecs list-daemon-task-definitions --family ${family}
// aws ecs describe-daemon-task-definition --daemon-task-definition ${arn}
func (store *Store) ListDaemonTaskDefinitions(family *string) ([]types.DaemonTaskDefinition, error) {
	limit := int32(20)
	output, err := store.ecs.ListDaemonTaskDefinitions(context.Background(), &ecs.ListDaemonTaskDefinitionsInput{
		Family:    family,
		MaxResults: &limit,
		Sort:      types.SortOrderDesc,
	})
	if err != nil {
		slog.Warn("failed to run aws api to list daemon task definitions", "error", err)
		return nil, err
	}

	results := []types.DaemonTaskDefinition{}
	for _, summary := range output.DaemonTaskDefinitions {
		if summary.Arn == nil {
			continue
		}
		td, err := store.DescribeDaemonTaskDefinition(summary.Arn)
		if err != nil {
			slog.Warn("failed to describe daemon task definition", "arn", *summary.Arn, "error", err)
			continue
		}
		results = append(results, *td)
	}

	return results, nil
}

// Equivalent to
// aws ecs describe-daemon-task-definition --daemon-task-definition ${arn}
func (store *Store) DescribeDaemonTaskDefinition(arn *string) (*types.DaemonTaskDefinition, error) {
	output, err := store.ecs.DescribeDaemonTaskDefinition(context.Background(), &ecs.DescribeDaemonTaskDefinitionInput{
		DaemonTaskDefinition: arn,
	})
	if err != nil {
		slog.Warn("failed to run aws api to describe daemon task definition", "error", err)
		return nil, err
	}
	return output.DaemonTaskDefinition, nil
}

// Equivalent to
// aws ecs describe-daemon-revisions --daemon-revision-arns ${arns}
func (store *Store) DescribeDaemonRevisions(arns []string) ([]types.DaemonRevision, error) {
	output, err := store.ecs.DescribeDaemonRevisions(context.Background(), &ecs.DescribeDaemonRevisionsInput{
		DaemonRevisionArns: arns,
	})
	if err != nil {
		slog.Warn("failed to run aws api to describe daemon revisions", "error", err)
		return nil, err
	}
	return output.DaemonRevisions, nil
}
