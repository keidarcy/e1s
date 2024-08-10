package api

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Store struct {
	*aws.Config
	Region         string
	Profile        string
	ecs            *ecs.Client
	cloudwatch     *cloudwatch.Client
	cloudwatchlogs *cloudwatchlogs.Client
	autoScaling    *applicationautoscaling.Client
	ssm            *ssm.Client
	lambda         *lambda.Client
}

func NewStore() (*Store, error) {
	profile := os.Getenv("AWS_PROFILE")
	region := os.Getenv("AWS_REGION")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		slog.Error("failed to load aws SDK config", "error", err)
		return nil, err
	}
	ecsClient := ecs.NewFromConfig(cfg)
	slog.Info("load config", slog.String("AWS_PROFILE", profile), slog.String("AWS_REGION", cfg.Region))

	lambdaClient := lambda.NewFromConfig(cfg)
	slog.Info("load lambda config", slog.String("AWS_PROFILE", profile), slog.String("AWS_REGION", cfg.Region))

	return &Store{
		Config:  &cfg,
		Region:  cfg.Region,
		Profile: profile,
		ecs:     ecsClient,
		lambda:  lambdaClient,
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
