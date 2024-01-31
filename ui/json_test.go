package ui

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func TestGetJsonData(t *testing.T) {
	app, _ := newApp()
	view := newView(app, ClusterPage, []KeyInput{}, secondaryPageKeyMap{
		JsonPage: []KeyInput{
			{key: string(fKey), description: toggleFullScreen},
		},
	})
	type input struct {
		entity Entity
		which  string
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
		name  string
		input input
		want  string
	}{
		{
			name: "cluster",
			input: input{
				entity: Entity{
					cluster: cluster,
				},
				which: "cluster",
			},
			want: colorizeJSON(clusterBytes),
		},
		{
			name: "service",
			input: input{
				entity: Entity{
					service: service,
					events:  events,
				},
				which: "service",
			},
			want: colorizeJSON(serviceBytes),
		},
		{
			name: "service",
			input: input{
				entity: Entity{
					service: service,
					events:  events,
				},
				which: "events",
			},
			want: colorizeJSON(eventsBytes),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := view.getJsonString(tc.input.entity, tc.input.which)
			if string(result) != tc.want {
				t.Errorf("Got: %s, Want: %s\n", result, tc.want)
			}
		})
	}
}
