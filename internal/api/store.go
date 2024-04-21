package api

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

type Store struct {
	*aws.Config
	ecs            *ecs.Client
	cloudwatch     *cloudwatch.Client
	cloudwatchlogs *cloudwatchlogs.Client
	autoScaling    *applicationautoscaling.Client
	ssm            *ssm.Client
}

func NewStore(logr *logrus.Logger) (*Store, error) {
	logger = logr
	logger.Infof("e1s load default config with AWS_PROFILE: %q, AWS_REGION: %q", os.Getenv("AWS_PROFILE"), os.Getenv("AWS_REGION"))
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		logger.Errorf("Failed to load aws SDK config, error: %v\n", err)
		return nil, err
	}
	ecsClient := ecs.NewFromConfig(cfg)
	return &Store{
		Config: &cfg,
		ecs:    ecsClient,
	}, nil
}

func (store *Store) initCloudwatchClient() {
	if store.cloudwatch == nil {
		store.cloudwatch = cloudwatch.NewFromConfig(*store.Config)
	}
}

func (store *Store) initCloudwatchlogsClient() {
	if store.cloudwatchlogs == nil {
		store.cloudwatchlogs = cloudwatchlogs.NewFromConfig(*store.Config)
	}
}

func (store *Store) initSsmClient() {
	if store.ssm == nil {
		store.ssm = ssm.NewFromConfig(*store.Config)
	}
}
