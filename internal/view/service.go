package view

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type serviceView struct {
	view
	services []types.Service
}

func newServiceView(services []types.Service, app *App) *serviceView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["U"],
		hotKeyMap["w"],
		hotKeyMap["t"],
		hotKeyMap["L"],
		hotKeyMap["m"],
		hotKeyMap["a"],
		hotKeyMap["p"],
	}...)
	return &serviceView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind:   describePageKeys,
			LogKind:           logPageKeys,
			AutoScalingKind:   describePageKeys,
			ServiceEventsKind: otherDescribePageKeys,
		}),
		services: services,
	}
}

func (app *App) showServicesPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	var resources []types.Service
	var err error
	if len(app.bootstrapServices) > 0 && !reload {
		resources = app.bootstrapServices
		app.bootstrapServices = nil
	} else {
		resources, err = app.Store.ListServices(app.cluster.ClusterName)
	}

	// Set default service if provided through options
	if app.Option.Service != "" && !reload {
		for _, s := range resources {
			if *s.ServiceName == app.Option.Service {
				app.service = &s
				app.events = s.Events
				err = app.showPrimaryKindPage(TaskKind, false)
				if err != nil {
					return err
				}
				return nil
			}
		}
		// If service not found, reset the option and show warning
		slog.Warn("service not found", "service", app.Option.Service)
		app.Notice.Warnf("Service '%s' not found in cluster '%s'", app.Option.Service, *app.cluster.ClusterName)
		app.Option.Service = ""
	}

	err = buildResourcePage(resources, app, err, func() resourceViewBuilder {
		return newServiceView(resources, app)
	})
	return err
}

func (v *serviceView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.service
}

func (v *serviceView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.services))
	for i, s := range v.services {
		params = append(params, headerPageParam{
			title:      *s.ServiceName,
			entityName: *s.ServiceArn,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

// Generate info pages params
func (v *serviceView) headerPageItems(index int) (items []headerItem) {
	s := v.services[index]
	// publicIP
	ip := utils.EmptyText
	// security groups
	sgs := []string{}
	if s.NetworkConfiguration != nil && s.NetworkConfiguration.AwsvpcConfiguration != nil {
		sgs = append(sgs, s.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups...)
		ip = string(s.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp)
	}

	// target groups
	tgs := []string{}
	for _, lb := range s.LoadBalancers {
		tgs = append(tgs, utils.ArnToFullName(lb.TargetGroupArn))
	}

	// capacity provider strategy
	cps := []string{}
	for _, p := range s.CapacityProviderStrategy {
		cps = append(cps, *p.CapacityProvider)
	}
	cpsString := strings.Join(cps, ",")
	if len(cpsString) == 0 {
		cpsString = utils.EmptyText
	}

	// deployment circuit breaker enable
	dcbe := utils.EmptyText
	// deployment circuit breaker rollback
	dcbr := utils.EmptyText
	// deployment max percent
	dmaxp := utils.EmptyText
	// deployment min percent
	dminp := utils.EmptyText

	if s.DeploymentConfiguration != nil {
		dmaxp = strconv.Itoa(int(*s.DeploymentConfiguration.MaximumPercent)) + "%"
		dminp = strconv.Itoa(int(*s.DeploymentConfiguration.MinimumHealthyPercent)) + "%"
		if s.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			dcbe = strconv.FormatBool(s.DeploymentConfiguration.DeploymentCircuitBreaker.Enable)
			dcbr = strconv.FormatBool(s.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback)
		}
	}

	// deployment controller
	dc := utils.EmptyText
	if s.DeploymentController != nil {
		dc = string(s.DeploymentController.Type)
	}

	items = []headerItem{
		{name: "Name", value: utils.ShowString(s.ServiceName)},
		{name: "Cluster", value: utils.ArnToName(s.ClusterArn)},
		{name: "Capacity provider strategy", value: cpsString},
		{name: "Role arn", value: utils.ArnToName(s.RoleArn)},
		{name: "Task definition", value: utils.ArnToName(s.TaskDefinition)},
		{name: "Propagate tags", value: string(s.PropagateTags)},
		{name: "Scheduling strategy", value: string(s.SchedulingStrategy)},
		{name: "Deployment controller", value: dc},
		{name: "Deployment circuit breaker enable", value: dcbe},
		{name: "Deployment circuit breaker rollback", value: dcbr},
		{name: "Deployment maximum", value: dmaxp},
		{name: "Deployment minimum", value: dminp},
		{name: "Public IP", value: ip},
		{name: "Security groups", value: strings.Join(sgs, ",")},
		{name: "Target groups", value: strings.Join(tgs, ",")},
		{name: "Execute command", value: strconv.FormatBool(s.EnableExecuteCommand)},
		{name: "Created", value: utils.ShowTime(s.CreatedAt)},
		{name: "Created by", value: utils.ArnToName(s.CreatedBy)},
		{name: "Platform family", value: utils.ShowString(s.PlatformFamily)},
		{name: "Platform version", value: utils.ShowString(s.PlatformVersion)},
		{name: "Tags count", value: strconv.Itoa(len(s.Tags))},
	}
	return
}

// Generate table params
func (v *serviceView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, "Services", *v.app.cluster.ClusterName, len(v.services))
	headers = []string{
		"Name",
		"Status",
		"Tasks",
		"Pending",
		"Last deployment",
		"Execute command",
		"Task definition",
		"Age",
	}

	rowsBuilder = func() (data [][]string) {
		for _, s := range v.services {
			row := []string{}

			// tasks
			tasks := fmt.Sprintf("%d/%d Tasks running", s.RunningCount, s.DesiredCount)

			// last deployment
			lastDeployment := ""
			// last update time
			var lastUpdateTime *time.Time

			if len(s.Deployments) > 0 {
				rollout := string(s.Deployments[0].RolloutState)
				lastUpdateTime = s.Deployments[0].CreatedAt
				lastDeployment += fmt.Sprintf("%s - %s", utils.ShowGreenGrey(&rollout, "completed"), utils.Age(lastUpdateTime))
			}

			// enable execute command
			enableExecuteCommand := "False"
			if s.EnableExecuteCommand {
				enableExecuteCommand = "True"
			}

			row = append(row, utils.ShowString(s.ServiceName))
			row = append(row, utils.ShowGreenGrey(s.Status, "active"))
			row = append(row, tasks)
			row = append(row, utils.ShowInt(&s.PendingCount))
			row = append(row, lastDeployment)
			row = append(row, utils.ShowGreenGrey(&enableExecuteCommand, "true"))
			row = append(row, utils.ArnToName(s.TaskDefinition))
			row = append(row, utils.Age(s.CreatedAt))
			data = append(data, row)

			entity := Entity{service: &s, events: s.Events, entityName: *s.ServiceArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}
	return
}
