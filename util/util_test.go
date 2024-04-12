package util

import (
	"fmt"
	"testing"
)

const (
	testRegion    = "us-east-1"
	clusterArnFmt = "arn:aws:ecs:%s:111111:cluster/%s"
	serviceArnFmt = "arn:aws:ecs:%s:111111:service/%s/%s"
	taskArnFmt    = "arn:aws:ecs:%s:111111:task/%s/%s"
)

func TestArnToURL(t *testing.T) {
	type Args struct {
		arn         string
		taskService string
	}
	cluster1 := "cluster1"
	service1 := "service1"
	task1 := "task1"
	taskService1 := "taskService1"
	arn1 := fmt.Sprintf(clusterArnFmt, testRegion, cluster1)
	url1 := fmt.Sprintf(clusterURLFmt, testRegion, cluster1, testRegion)
	arn2 := fmt.Sprintf(serviceArnFmt, testRegion, cluster1, service1)
	url2 := fmt.Sprintf(serviceURLFmt, testRegion, cluster1, service1, testRegion)
	arn3 := fmt.Sprintf(taskArnFmt, testRegion, cluster1, task1)
	url3 := fmt.Sprintf(taskURLFmt, testRegion, cluster1, taskService1, task1, testRegion)

	testCases := []struct {
		name string
		args Args
		want string
	}{
		{
			name: "cluster arn convert",
			args: Args{
				arn:         arn1,
				taskService: "",
			},
			want: url1,
		},
		{
			name: "service arn convert",
			args: Args{
				arn:         arn2,
				taskService: "",
			},
			want: url2,
		},
		{
			name: "task arn convert",
			args: Args{
				arn:         arn3,
				taskService: taskService1,
			},
			want: url3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ArnToUrl(tc.args.arn, tc.args.taskService)
			if result != tc.want {
				t.Errorf("Arn: %s, taskService: %s, Got: %s, Want: %s\n", tc.args.arn, tc.args.taskService, result, tc.want)
			}
		})
	}
}

func TextBuildMeterText(t *testing.T) {
	testCases := []struct {
		name  string
		input float64
		want  string
	}{
		{
			name:  "value=2.123",
			input: 2.123,
			want:  "█▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒",
		},
		{
			name:  "value=12.123",
			input: 12.123,
			want:  "██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒",
		},
		{
			name:  "value=2.123",
			input: 52.123,
			want:  "██████████▒▒▒▒▒▒▒▒▒▒",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildMeterText(tc.input)
			if result != tc.want {
				t.Errorf("input: %v, want: %v, results %v\n", tc.input, tc.want, result)
			}
		})
	}

}

func TestGetRegistryInfo(t *testing.T) {
	tests := []struct {
		imageURL          string
		expectedRegistry  string
		expectedImageName string
	}{
		{"ubuntu:latest", "Docker Hub", "ubuntu:latest"},
		{"ubuntu:1234567890", "Docker Hub", "ubuntu:12345678..."},
		{"123456789012.dkr.ecr.region.amazonaws.com/my-image:tag", "Amazon ECR", "my-image:tag"},
		{"public.ecr.aws/username/my-image:tag", "Amazon ECR Public", "username/my-image:tag"},
		{"gcr.io/my-project/my-image:tag", "Google GCR", "my-project/my-image:tag"},
		{"myregistry.azurecr.io/my-image:tag", "Azure ACR", "my-image:tag"},
		{"registry.gitlab.com/my-group/my-project/my-image:tag", "GitLab", "my-group/my-project/my-image:tag"},
		{"ghcr.io/username/my-image:tag", "GitHub", "username/my-image:tag"},
		{"quay.io/username/my-image:tag", "Quay", "username/my-image:tag"},
	}

	for _, test := range tests {
		t.Run(test.imageURL, func(t *testing.T) {
			registry, imageName := ImageInfo(&test.imageURL)
			if registry != test.expectedRegistry {
				t.Errorf("getRegistryInfo(%q) got registry %q, want %q", test.imageURL, registry, test.expectedRegistry)
			}
			if imageName != test.expectedImageName {
				t.Errorf("getRegistryInfo(%q) got image name %q, want %q", test.imageURL, imageName, test.expectedImageName)
			}
		})
	}
}
