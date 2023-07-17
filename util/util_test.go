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
