package view

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func TestValidateContainerSessionTarget(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		runtimeId, containerName, err := validateContainerSessionTarget(Entity{
			container: &types.Container{
				RuntimeId: aws.String("runtime-id"),
				Name:      aws.String("app"),
			},
		})
		if err != nil {
			t.Errorf("Got: %v, Want: %v", err, nil)
		}
		if runtimeId != "runtime-id" {
			t.Errorf("runtimeId Got: %q, Want: %q", runtimeId, "runtime-id")
		}
		if containerName != "app" {
			t.Errorf("containerName Got: %q, Want: %q", containerName, "app")
		}
	})

	testCases := []struct {
		name     string
		selected Entity
	}{
		{name: "missing container", selected: Entity{}},
		{name: "missing runtime id", selected: Entity{container: &types.Container{Name: aws.String("app")}}},
		{name: "empty runtime id", selected: Entity{container: &types.Container{RuntimeId: aws.String(""), Name: aws.String("app")}}},
		{name: "missing name", selected: Entity{container: &types.Container{RuntimeId: aws.String("runtime-id")}}},
		{name: "empty name", selected: Entity{container: &types.Container{RuntimeId: aws.String("runtime-id"), Name: aws.String("")}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := validateContainerSessionTarget(tc.selected); err == nil {
				t.Errorf("Got nil error, Want non nil error")
			}
		})
	}
}
