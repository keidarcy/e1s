package api

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
)

type AutoScalingData struct {
	Targets    []types.ScalableTarget
	Policies   []types.ScalingPolicy
	Actions    []types.ScheduledAction
	Activities []types.ScalingActivity
}

func (store *Store) GetAutoscaling(serviceArn *string) (*AutoScalingData, error) {
	targets, err := store.describeScalableTargets(serviceArn)

	if err != nil {
		return nil, err
	}

	policies, err := store.describeScalingPolicies(serviceArn)

	if err != nil {
		return nil, err
	}

	activities, err := store.describeScalingActivities(serviceArn)

	if err != nil {
		return nil, err
	}

	actions, err := store.describeScheduledAction(serviceArn)

	if err != nil {
		return nil, err
	}

	return &AutoScalingData{
		Targets:    targets,
		Policies:   policies,
		Actions:    actions,
		Activities: activities,
	}, nil

}

// Equivalent to
// aws application-autoscaling describe-scaling-activities --service-namespace ecs --resource-id {ServiceArn}
// Auto scaling logs
func (store *Store) describeScalingActivities(serviceArn *string) ([]types.ScalingActivity, error) {
	activitiesInput := &applicationautoscaling.DescribeScalingActivitiesInput{
		ServiceNamespace: "ecs",
		ResourceId:       serviceArn,
		MaxResults:       aws.Int32(10),
	}
	activitiesOutput, err := store.autoScaling.DescribeScalingActivities(context.Background(), activitiesInput)

	if err != nil {
		logger.Printf("e1s - aws failed to auto scaling activities serviceArn: \"%s\", err: %v\n", *serviceArn, err)
		return nil, err
	}

	return activitiesOutput.ScalingActivities, nil
}

// Equivalent to
// aws application-autoscaling describe-scalable-targets --service-namespace ecs --resource-ids {[ServiceArn]}
func (store *Store) describeScalableTargets(serviceArn *string) ([]types.ScalableTarget, error) {
	store.getAutoScalingClient()
	targetsInput := &applicationautoscaling.DescribeScalableTargetsInput{
		ServiceNamespace: "ecs",
		ResourceIds:      []string{*serviceArn},
	}
	targetsOutput, err := store.autoScaling.DescribeScalableTargets(context.Background(), targetsInput)

	if err != nil {
		logger.Printf("e1s - aws failed to auto scaling activities serviceArn: \"%s\", err: %v\n", *serviceArn, err)
		return nil, err
	}

	return targetsOutput.ScalableTargets, nil
}

// Equivalent to
// aws application-autoscaling describe-scaling-policies --service-namespace ecs --resource-id "service/<ClusterName>/<ServiceName>"
func (store *Store) describeScalingPolicies(serviceArn *string) ([]types.ScalingPolicy, error) {
	store.getAutoScalingClient()
	policiesInput := &applicationautoscaling.DescribeScalingPoliciesInput{
		ServiceNamespace: "ecs",
		ResourceId:       serviceArn,
	}
	policiesOutput, err := store.autoScaling.DescribeScalingPolicies(context.Background(), policiesInput)

	if err != nil {
		logger.Printf("e1s - aws failed to auto scaling activities serviceArn: \"%s\", err: %v\n", *serviceArn, err)
		return nil, err
	}

	return policiesOutput.ScalingPolicies, nil
}

func (store *Store) getAutoScalingClient() {
	if store.autoScaling == nil {
		store.autoScaling = applicationautoscaling.NewFromConfig(*store.Config)
	}
}

// Equivalent to
// aws application-autoscaling describe-scheduled-actions --service-namespace ecs --resource-id "service/<ClusterName>/<ServiceName>"
func (store *Store) describeScheduledAction(serviceArn *string) ([]types.ScheduledAction, error) {
	store.getAutoScalingClient()
	actionsInput := &applicationautoscaling.DescribeScheduledActionsInput{
		ServiceNamespace: "ecs",
		ResourceId:       serviceArn,
	}
	actionsOutput, err := store.autoScaling.DescribeScheduledActions(context.Background(), actionsInput)

	if err != nil {
		logger.Printf("e1s - aws failed to auto scaling scheduled actions serviceArn: \"%s\", err: %v\n", *serviceArn, err)
		return nil, err
	}

	return actionsOutput.ScheduledActions, nil
}
