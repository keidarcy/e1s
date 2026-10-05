package api

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func TestResolveLogLocation(t *testing.T) {
	awslogsContainer := types.ContainerDefinition{
		Name: aws.String("app"),
		LogConfiguration: &types.LogConfiguration{
			LogDriver: types.LogDriverAwslogs,
			Options: map[string]string{
				"awslogs-group":         "/ecs/app",
				"awslogs-stream-prefix": "ecs",
			},
		},
	}
	firelensContainer := types.ContainerDefinition{
		Name: aws.String("app"),
		LogConfiguration: &types.LogConfiguration{
			LogDriver: types.LogDriverAwsfirelens,
		},
	}

	tests := []struct {
		name       string
		overrides  []CloudwatchLogOverride
		container  types.ContainerDefinition
		wantGroup  string
		wantStream string
		wantOk     bool
	}{
		{
			name:       "awslogs without config",
			container:  awslogsContainer,
			wantGroup:  "/ecs/app",
			wantStream: "ecs/app/task-id",
			wantOk:     true,
		},
		{
			name:      "firelens without config",
			container: firelensContainer,
			wantOk:    false,
		},
		{
			name:       "firelens with matching config",
			overrides:  []CloudwatchLogOverride{{Family: "web-*", Container: "app", Group: "/firelens/web", Stream: "{container}-firelens-{taskId}"}},
			container:  firelensContainer,
			wantGroup:  "/firelens/web",
			wantStream: "app-firelens-task-id",
			wantOk:     true,
		},
		{
			name:       "empty patterns match every container",
			overrides:  []CloudwatchLogOverride{{Group: "/firelens/web"}},
			container:  firelensContainer,
			wantGroup:  "/firelens/web",
			wantStream: "app/task-id",
			wantOk:     true,
		},
		{
			name:      "not matching family falls back",
			overrides: []CloudwatchLogOverride{{Family: "worker-*", Group: "/firelens/worker"}},
			container: firelensContainer,
			wantOk:    false,
		},
		{
			name:       "config wins over awslogs options",
			overrides:  []CloudwatchLogOverride{{Container: "app", Group: "/firelens/web"}},
			container:  awslogsContainer,
			wantGroup:  "/firelens/web",
			wantStream: "app/task-id",
			wantOk:     true,
		},
		{
			name:      "entry without group keeps the log driver defaults",
			overrides: []CloudwatchLogOverride{{Container: "app"}, {Group: "/firelens/web"}},
			container: firelensContainer,
			wantOk:    false,
		},
		{
			name:       "entry without group falls back to awslogs options",
			overrides:  []CloudwatchLogOverride{{Container: "app"}, {Group: "/firelens/web"}},
			container:  awslogsContainer,
			wantGroup:  "/ecs/app",
			wantStream: "ecs/app/task-id",
			wantOk:     true,
		},
		{
			name:      "no log configuration",
			container: types.ContainerDefinition{Name: aws.String("app")},
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, stream, ok := resolveLogLocation(tt.overrides, "web-app", tt.container, "task-id")
			if ok != tt.wantOk {
				t.Fatalf("ok got %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if group != tt.wantGroup {
				t.Errorf("group got %s, want %s", group, tt.wantGroup)
			}
			if stream != tt.wantStream {
				t.Errorf("stream got %s, want %s", stream, tt.wantStream)
			}
		})
	}
}
