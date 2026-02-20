package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

// Add new type for service deployment view
type serviceDeploymentView struct {
	view
	serviceDeployments []types.ServiceDeployment
}

// Constructor for service deployment view
func newServiceDeploymentView(serviceDeployments []types.ServiceDeployment, app *App) *serviceDeploymentView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["v"],
		hotKeyMap["R"],
	}...)
	return &serviceDeploymentView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind:     describePageKeys,
			ServiceRevisionKind: describePageKeys,
		}),
		serviceDeployments: serviceDeployments,
	}
}

// Show service deployment page
func (app *App) showServiceDeploymentPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	resources, err := app.Store.ListServiceDeployments(app.cluster.ClusterName, app.service.ServiceName)
	err = buildResourcePage(resources, app, err, func() resourceViewBuilder {
		return newServiceDeploymentView(resources, app)
	})
	return err
}

func (v *serviceDeploymentView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.serviceDeployment
}

// Build info pages for service deployment page
func (v *serviceDeploymentView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.serviceDeployments))
	for i, d := range v.serviceDeployments {
		params = append(params, headerPageParam{
			title:      utils.ArnToName(d.ServiceDeploymentArn),
			entityName: *d.ServiceDeploymentArn,
			items:      v.headerPageItems(i),
		})
	}

	return params
}

// Generate info pages params
func (v *serviceDeploymentView) headerPageItems(index int) (items []headerItem) {
	d := v.serviceDeployments[index]
	items = []headerItem{
		{name: "Status", value: string(d.Status)},
		{name: "Status Reason", value: utils.ShowString(d.StatusReason)},
		{name: "Alarm status", value: string(d.Alarms.Status)},
		{name: "Circuit Breaker Status", value: string(d.DeploymentCircuitBreaker.Status)},
		{name: "Circuit Breaker Failure Count", value: func() string {
			if d.DeploymentCircuitBreaker == nil {
				return "-"
			}
			return fmt.Sprintf("%d/%d",
				d.DeploymentCircuitBreaker.FailureCount,
				d.DeploymentCircuitBreaker.Threshold)
		}()},
		{name: "Deployment Config", value: func() string {
			if d.DeploymentConfiguration == nil {
				return "-"
			}
			return fmt.Sprintf("Max: %d%%, Min: %d%%",
				*d.DeploymentConfiguration.MaximumPercent,
				*d.DeploymentConfiguration.MinimumHealthyPercent)
		}()},
		{name: "Circuit Breaker Config", value: func() string {
			if d.DeploymentConfiguration == nil || d.DeploymentConfiguration.DeploymentCircuitBreaker == nil {
				return "-"
			}
			return fmt.Sprintf("Enable: %v, Rollback: %v",
				d.DeploymentConfiguration.DeploymentCircuitBreaker.Enable,
				d.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback)
		}()},
		{name: "Target task count", value: func() string {
			if d.TargetServiceRevision == nil {
				return "-"
			}
			return fmt.Sprintf("Requested: %d, Running: %d, Pending: %d",
				d.TargetServiceRevision.RequestedTaskCount,
				d.TargetServiceRevision.RunningTaskCount,
				d.TargetServiceRevision.PendingTaskCount)
		}()},
		{name: "Source task count", value: func() string {
			if len(d.SourceServiceRevisions) == 0 {
				return "-"
			}
			return fmt.Sprintf("Requested: %d, Running: %d, Pending: %d",
				d.SourceServiceRevisions[0].RequestedTaskCount,
				d.SourceServiceRevisions[0].RunningTaskCount,
				d.SourceServiceRevisions[0].PendingTaskCount)
		}()},
		{name: "Created At", value: utils.ShowTime(d.CreatedAt)},
		{name: "Started At", value: utils.ShowTime(d.StartedAt)},
		{name: "Updated At", value: utils.ShowTime(d.UpdatedAt)},
		{name: "Finished At", value: utils.ShowTime(d.FinishedAt)},
	}
	return
}

// Generate table params
func (v *serviceDeploymentView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	serviceName := ""
	if v.app.service.ServiceName != nil {
		serviceName = *v.app.service.ServiceName
	}

	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, serviceName, len(v.serviceDeployments))
	headers = []string{
		"Deployment ID",
		"Status",
		"Target service revision",
		"Created At",
		"Started At",
		"Finished At",
		"Duration",
	}

	rowsBuilder = func() (data [][]string) {
		for _, d := range v.serviceDeployments {
			duration := "-"
			if d.StartedAt != nil && d.FinishedAt != nil {
				duration = utils.Duration(*d.StartedAt, *d.FinishedAt)
			}

			status := string(d.Status)
			row := []string{
				utils.ArnToName(d.ServiceDeploymentArn),
				utils.ShowGreenGrey(&status, "successful"),
				utils.ArnToName(d.TargetServiceRevision.Arn),
				utils.ShowTime(d.CreatedAt),
				utils.ShowTime(d.StartedAt),
				utils.ShowTime(d.FinishedAt),
				duration,
			}
			data = append(data, row)

			entity := Entity{serviceDeployment: &d, entityName: *d.ServiceDeploymentArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}

	return
}
