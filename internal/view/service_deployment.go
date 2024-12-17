package view

import (
	"fmt"
	"log/slog"

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
		hotKeyMap["U"],
	}...)
	return &serviceDeploymentView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		serviceDeployments: serviceDeployments,
	}
}

// Show service deployment page
func (app *App) showServiceDeploymentPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	serviceDeployments, err := app.Store.ListServiceDeployments(app.cluster.ClusterName, app.service.ServiceName)
	if err != nil {
		slog.Warn("failed to show service deployment pages", "error", err)
		app.back()
		return err
	}

	if len(serviceDeployments) == 0 {
		app.back()
		return fmt.Errorf("no service deployments found")
	}

	view := newServiceDeploymentView(serviceDeployments, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
}

// Build info pages for service deployment page
func (v *serviceDeploymentView) headerBuilder() *tview.Pages {
	for _, d := range v.serviceDeployments {
		title := utils.ArnToName(d.ServiceDeploymentArn)
		entityName := *d.ServiceDeploymentArn
		items := v.headerPagesParam(d)

		v.buildHeaderPages(items, title, entityName)
	}

	if len(v.serviceDeployments) > 0 && v.serviceDeployments[0].ServiceDeploymentArn != nil {
		v.headerPages.SwitchToPage(*v.serviceDeployments[0].ServiceDeploymentArn)
		v.changeSelectedValues()
	}
	return v.headerPages
}

// Generate info pages params
func (v *serviceDeploymentView) headerPagesParam(d types.ServiceDeployment) (items []headerItem) {
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
				d.DeploymentConfiguration.MaximumPercent,
				d.DeploymentConfiguration.MinimumHealthyPercent)
		}()},
		{name: "Circuit Breaker Config", value: func() string {
			if d.DeploymentConfiguration == nil || d.DeploymentConfiguration.DeploymentCircuitBreaker == nil {
				return "-"
			}
			return fmt.Sprintf("Enable: %v, Rollback: %v",
				d.DeploymentConfiguration.DeploymentCircuitBreaker.Enable,
				d.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback)
		}()},
		{name: "Source revision", value: func() string {
			if len(d.SourceServiceRevisions) == 0 {
				return "-"
			}
			return fmt.Sprintf("%s (Running: %d, Pending: %d)",
				utils.ArnToName(d.SourceServiceRevisions[0].Arn),
				d.SourceServiceRevisions[0].RunningTaskCount,
				d.SourceServiceRevisions[0].PendingTaskCount)
		}()},
		{name: "Target revision", value: func() string {
			if d.TargetServiceRevision == nil {
				return "-"
			}
			return fmt.Sprintf("%s (Running: %d, Pending: %d, Requested: %d)",
				utils.ArnToName(d.TargetServiceRevision.Arn),
				d.TargetServiceRevision.RunningTaskCount,
				d.TargetServiceRevision.PendingTaskCount,
				d.TargetServiceRevision.RequestedTaskCount)
		}()},
		{name: "Created At", value: utils.ShowTime(d.CreatedAt)},
		{name: "Started At", value: utils.ShowTime(d.StartedAt)},
		{name: "Updated At", value: utils.ShowTime(d.UpdatedAt)},
		{name: "Finished At", value: utils.ShowTime(d.FinishedAt)},
	}
	return
}

// Build footer for service deployment page
func (v *serviceDeploymentView) footerBuilder() *tview.Flex {
	v.footer.serviceDeployment.SetText(fmt.Sprintf(color.FooterSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Build table for service deployment page
func (v *serviceDeploymentView) bodyBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.bodyPages
}

// Handlers for task definition table
func (v *serviceDeploymentView) tableHandler() {
	for row, deployment := range v.serviceDeployments {
		d := deployment
		v.table.GetCell(row+1, 0).SetReference(Entity{serviceDeployment: &d, entityName: *d.ServiceDeploymentArn})
	}
}

// Generate table params
func (v *serviceDeploymentView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	serviceName := ""
	if v.app.service.ServiceName != nil {
		serviceName = *v.app.service.ServiceName
	}

	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, serviceName, len(v.serviceDeployments))
	headers = []string{
		"Deployment ID ▾",
		"Status",
		"Revision",
		"Created At",
		"Started At",
		"Finished At",
		"Duration",
	}

	dataBuilder = func() (data [][]string) {
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
		}
		return data
	}

	return
}
