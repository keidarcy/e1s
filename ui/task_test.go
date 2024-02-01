package ui

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
)

var (
	taskArn1 = "arn:aws/123"
	taskArn2 = "arn:aws/124"
)

func getTaskViews() []TaskView {
	now := time.Now()
	newTask := func() types.Task {
		return types.Task{
			CreatedAt:     &now,
			StartedAt:     &now,
			PullStartedAt: &now,
			PullStoppedAt: &now,
		}
	}
	task1 := newTask()
	task1.TaskArn = aws.String(taskArn1)

	task2 := newTask()
	task2.TaskArn = aws.String(taskArn2)

	app, _ := newApp(false)
	app.cluster = &types.Cluster{
		ClusterName: aws.String(clusterName1),
	}
	app.service = &types.Service{
		ServiceName: aws.String(serviceName1),
	}
	TaskView1 := newTaskView([]types.Task{task1}, app)
	TaskView2 := newTaskView([]types.Task{task2}, app)

	return []TaskView{*TaskView1, *TaskView2}
}

func TestTaskPageParams(t *testing.T) {
	taskViews := getTaskViews()

	testCases := []struct {
		name string
		view TaskView
		want want
	}{
		{
			name: taskArn1,
			view: taskViews[0],
			want: want{
				taskID: util.ArnToName(&taskArn1),
			},
		},
		{
			name: taskArn2,
			view: taskViews[1],
			want: want{
				taskID: util.ArnToName(&taskArn2),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, c := range tc.view.tasks {
				items := tc.view.infoPagesParam(c)
				if items[0].value != tc.want.taskID {
					t.Errorf("%s Got: %s, Want: %s\n", items[0].name, items[0].value, tc.want.taskID)
				}
			}
		})
	}
}

func TestTaskTableParam(t *testing.T) {
	taskViews := getTaskViews()

	testCases := []struct {
		name string
		view TaskView
		want want
	}{
		{
			name: serviceName1,
			view: taskViews[0],
			want: want{
				taskID: util.ArnToName(&taskArn1),
			},
		},
		{
			name: clusterName2,
			view: taskViews[1],
			want: want{
				taskID: util.ArnToName(&taskArn2),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, dataBuilder := tc.view.tableParam()
			matrix := dataBuilder()
			if matrix[0][0] != tc.want.taskID {
				t.Errorf("Name Got: %s, Want: %s\n", matrix[0][0], tc.want.taskID)
			}
		})
	}
}
