package api

import (
	"context"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
	"golang.org/x/sync/errgroup"
)

const (
	MaxTaskDefinitionFamily   = 50
	MaxTaskDefinitionRevision = 20
)

// Equivalent to
// aws ecs describe-task-definition --task-definition ${taskDefinition}
func (store *Store) DescribeTaskDefinition(tdArn *string) (types.TaskDefinition, error) {

	include := []types.TaskDefinitionField{
		types.TaskDefinitionFieldTags,
	}
	taskDefinition, err := store.ecs.DescribeTaskDefinition(context.Background(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: tdArn,
		Include:        include,
	})
	if err != nil {
		logger.Warnf("Failed to run aws api to describe task definition, error: %v\n", err)
		return types.TaskDefinition{}, err
	}

	return *taskDefinition.TaskDefinition, nil
}

type TaskDefinitionRevision = []string

// Equivalent to
// aws ecs list-task-definitions --family-prefix ${prefix}
func (store *Store) ListTaskDefinition(familyName *string) (TaskDefinitionRevision, error) {
	listTaskDefinitions, err := store.ecs.ListTaskDefinitions(context.Background(), &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: familyName,
		MaxResults:   aws.Int32(MaxTaskDefinitionRevision),
		Sort:         types.SortOrderDesc,
	})
	if err != nil {
		logger.Warnf("Failed to run aws api to list task definitions, error: %v\n", err)
		return nil, err
	}

	return listTaskDefinitions.TaskDefinitionArns, nil
}

// List given task definition revision with contents
// Equivalent to
// aws ecs list-task-definitions --family-prefix ${prefix}
// aws ecs describe-task-definition --task-definition ${taskDefinition}
func (store *Store) ListFullTaskDefinition(taskDefinition *string) ([]types.TaskDefinition, error) {
	td := strings.Split(util.ArnToName(taskDefinition), ":")
	familyName := td[0]
	list, err := store.ListTaskDefinition(&familyName)

	if err != nil {
		logger.Warnf("Failed to run aws api to run list task definition in ListFullTaskDefinition, error: %v\n", err)
		return []types.TaskDefinition{}, err
	}

	results := []types.TaskDefinition{}
	g := new(errgroup.Group)

	for _, t := range list {
		g.Go(func() error {
			d, err := store.DescribeTaskDefinition(&t)
			if err != nil {
				logger.Warnf("Failed to run aws api to describe task definition , err: %v", err)
				return err
			}
			results = append(results, d)
			return nil
		})
	}

	err = g.Wait()

	// sort by desire count, name ascending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Revision > results[j].Revision
	})
	return results, err
}

// Equivalent to
// aws ecs list-task-definition-families
func (store *Store) ListTaskDefinitionFamilies() ([]string, error) {
	familiesOutput, err := store.ecs.ListTaskDefinitionFamilies(context.Background(), &ecs.ListTaskDefinitionFamiliesInput{
		Status: types.TaskDefinitionFamilyStatusActive,
	})
	if err != nil {
		logger.Warnf("Failed to run aws api to list task definition families, error: %v\n", err)
		return nil, err
	}

	return familiesOutput.Families, nil
}
