package ui

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func TestGetJsonData(t *testing.T) {
	app, _ := newApp(Option{})
	view := newView(app, []KeyInput{}, secondaryPageKeyMap{
		DescriptionKind: []KeyInput{
			{key: string(fKey), description: toggleFullScreen},
		},
	})
	type input struct {
		entity Entity
	}
	cluster := &types.Cluster{
		ClusterName: aws.String(clusterName1),
	}
	events := []types.ServiceEvent{
		{
			Message: aws.String("good"),
		},
	}
	service := &types.Service{
		ServiceName: aws.String(serviceName1),
		Events:      events,
	}
	clusterBytes, _ := json.MarshalIndent(cluster, "", "  ")
	serviceBytes, _ := json.MarshalIndent(service, "", "  ")
	eventsBytes, _ := json.MarshalIndent(events, "", "  ")
	testCases := []struct {
		name       string
		input      input
		want       string
		changeKind func()
	}{
		{
			name: "cluster",
			input: input{
				entity: Entity{
					cluster: cluster,
				},
			},
			want: colorizeJSON(clusterBytes),
			changeKind: func() {
				app.kind = ClusterKind
			},
		},
		{
			name: "service",
			input: input{
				entity: Entity{
					service: service,
					events:  events,
				},
			},
			want: colorizeJSON(serviceBytes),
			changeKind: func() {
				app.kind = ServiceKind
			},
		},
		{
			name: "service events",
			input: input{
				entity: Entity{
					service: service,
					events:  events,
				},
			},
			want: colorizeJSON(eventsBytes),
			changeKind: func() {
				app.secondaryKind = ServiceEventsKind
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.changeKind()
			result := view.getJsonString(tc.input.entity)
			if string(result) != tc.want {
				t.Errorf("Name: %s, Got: %s, Want: %s\n", tc.name, result, tc.want)
			}
		})
	}
}
