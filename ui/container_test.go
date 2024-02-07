package ui

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var (
	containerName1 = "container1"
	containerName2 = "container2"
)

func getContainerViews() []ContainerView {
	newContainer := func() types.Container {
		return types.Container{}
	}
	container1 := newContainer()
	container1.Name = aws.String(containerName1)

	container2 := newContainer()
	container2.Name = aws.String(containerName2)

	app, _ := newApp(Option{})
	app.cluster = &types.Cluster{
		ClusterName: aws.String(clusterName1),
	}
	app.service = &types.Service{
		ServiceName: aws.String(serviceName1),
	}
	app.task = &types.Task{
		TaskArn: aws.String(taskArn1),
	}
	ContainerView1 := newContainerView([]types.Container{container1}, app)
	ContainerView2 := newContainerView([]types.Container{container2}, app)

	return []ContainerView{*ContainerView1, *ContainerView2}
}

func TestContainerPageParams(t *testing.T) {
	containerViews := getContainerViews()

	testCases := []struct {
		name string
		view ContainerView
		want want
	}{
		{
			name: containerName1,
			view: containerViews[0],
			want: want{
				containerName: containerName1,
			},
		},
		{
			name: containerName2,
			view: containerViews[1],
			want: want{
				containerName: containerName2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, c := range tc.view.containers {
				items := tc.view.infoPagesParam(c)
				if items[0].value != tc.want.containerName {
					t.Errorf("%s Got: %s, Want: %s\n", items[0].name, items[0].value, tc.want.containerName)
				}
			}
		})
	}
}

func TestContainerTableParam(t *testing.T) {
	containerViews := getContainerViews()

	testCases := []struct {
		name string
		view ContainerView
		want want
	}{
		{
			name: containerName1,
			view: containerViews[0],
			want: want{
				containerName: containerName1,
			},
		},
		{
			name: clusterName2,
			view: containerViews[1],
			want: want{
				containerName: containerName2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, dataBuilder := tc.view.tableParam()
			matrix := dataBuilder()
			if matrix[0][0] != tc.want.containerName {
				t.Errorf("Name Got: %s, Want: %s\n", matrix[0][0], tc.want.containerName)
			}
		})
	}
}
