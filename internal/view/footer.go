package view

import (
	"fmt"

	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

// View footer struct
type footer struct {
	footerFlex        *tview.Flex
	cluster           *tview.TextView
	service           *tview.TextView
	task              *tview.TextView
	container         *tview.TextView
	instance       *tview.TextView
	taskDefinition    *tview.TextView
	serviceDeployment *tview.TextView
	help              *tview.TextView
}

func newFooter() *footer {
	footerFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	footerFlex.SetBackgroundColor(color.Color(theme.BgColor))
	return &footer{
		footerFlex:        footerFlex,
		cluster:           tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, ClusterKind)),
		service:           tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, ServiceKind)),
		task:              tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, TaskKind)),
		container:         tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, ContainerKind)),
		instance:       tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, InstanceKind)).SetTextAlign(L),
		taskDefinition:    tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, TaskDefinitionKind)).SetTextAlign(L),
		serviceDeployment: tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, ServiceDeploymentKind)).SetTextAlign(L),
		help:              tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterItemFmt, HelpKind)).SetTextAlign(L),
	}
}
func (v *view) addFooterItems() {
	// left resources
	v.footer.footerFlex.AddItem(v.footer.cluster, 13, 0, false).
		AddItem(v.footer.service, 13, 0, false).
		AddItem(v.footer.task, 10, 0, false).
		AddItem(v.footer.container, 15, 0, false)

	// keep middle space
	if v.app.kind == TaskDefinitionKind {
		v.footer.footerFlex.
			AddItem(tview.NewTextView(), 5, 0, false).
			AddItem(v.footer.taskDefinition, 0, 1, false)
	} else if v.app.kind == InstanceKind {
		v.footer.footerFlex.
			AddItem(tview.NewTextView(), 5, 0, false).
			AddItem(v.footer.instance, 0, 1, false)
	} else if v.app.kind == ServiceDeploymentKind {
		v.footer.footerFlex.
			AddItem(tview.NewTextView(), 5, 0, false).
			AddItem(v.footer.serviceDeployment, 0, 1, false)
	} else if v.app.kind == HelpKind {
		v.footer.footerFlex.
			AddItem(tview.NewTextView(), 5, 0, false).
			AddItem(v.footer.help, 0, 1, false)
	} else {
		v.footer.footerFlex.
			AddItem(tview.NewTextView(), 0, 1, false)
	}

	// right labels
	// aws cli awsInfo label
	awsInfo := v.app.region
	if v.app.profile != "" {
		awsInfo = fmt.Sprintf("%s:%s", v.app.profile, v.app.region)
	}
	awsInfoView := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterAwsFmt, awsInfo))
	v.footer.footerFlex.AddItem(awsInfoView, len(awsInfo)+3, 0, false)

	// e1s info label
	t := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.FooterE1sFmt, utils.AppName, utils.AppVersion))
	v.footer.footerFlex.AddItem(t, 14, 0, false)
}
