package api

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

const (
	MaxTaskDefinitionFamily   = 50
	MaxTaskDefinitionRevision = 20
)

// Equivalent to
// aws ecs describe-task-definition --task-definition ${taskDefinition}
func (store *Store) DescribeTaskDefinition(name *string) (types.TaskDefinition, error) {

	include := []types.TaskDefinitionField{
		types.TaskDefinitionFieldTags,
	}
	taskDefinition, err := store.ecs.DescribeTaskDefinition(context.Background(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: name,
		Include:        include,
	})
	if err != nil {
		logger.Printf("e1s - aws failed to describe task definition, error: %v\n", err)
		return types.TaskDefinition{}, err
	}

	return *taskDefinition.TaskDefinition, nil
}

type TaskDefinitionRevision = []string

// Equivalent to
// aws ecs list-task-definitions --family-prefix ${prefix}
func (store *Store) ListTaskDefinition(name *string) (TaskDefinitionRevision, error) {
	listTaskDefinitions, err := store.ecs.ListTaskDefinitions(context.Background(), &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: name,
		MaxResults:   aws.Int32(MaxTaskDefinitionRevision),
		Sort:         types.SortOrderDesc,
	})
	if err != nil {
		logger.Printf("e1s - aws failed to list task definitions, error: %v\n", err)
		return nil, err
	}

	return listTaskDefinitions.TaskDefinitionArns, nil
}

type FullTaskDefinition = map[string][]string

// Deprecated
// List all task definition family with revisions for service update form(not support yet)
func (store *Store) ListFullTaskDefinition() (FullTaskDefinition, error) {
	listFamily, err := store.ecs.ListTaskDefinitionFamilies(context.Background(), &ecs.ListTaskDefinitionFamiliesInput{
		MaxResults: aws.Int32(MaxTaskDefinitionFamily),
	})
	if err != nil {
		logger.Printf("e1s - aws failed to list task definition family, error: %v\n", err)
		return nil, err
	}
	results := make(map[string][]string)
	for _, family := range listFamily.Families {
		taskDefinition, err := store.ListTaskDefinition(aws.String(family))
		logger.Println(taskDefinition)
		if err != nil {
			logger.Printf("e1s - aws failed to list task definitions, error: %v\n", err)
		}
		results[family] = taskDefinition
	}
	return results, nil
}

// Equivalent to
// aws ecs list-task-definition-families
func (store *Store) ListTaskDefinitionFamilies() ([]string, error) {
	familiesOutput, err := store.ecs.ListTaskDefinitionFamilies(context.Background(), &ecs.ListTaskDefinitionFamiliesInput{
		Status: types.TaskDefinitionFamilyStatusActive,
	})
	if err != nil {
		logger.Printf("e1s - aws failed to list task definition families, error: %v\n", err)
		return nil, err
	}

	return familiesOutput.Families, nil
}
