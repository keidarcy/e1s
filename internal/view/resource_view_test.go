package view

import (
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/rivo/tview"
)

type testResourceViewBuilder struct {
	v            *view
	footer       *tview.TextView
	headerParams []headerPageParam
	title        string
	headers      []string
	rows         [][]string
}

func (b *testResourceViewBuilder) getViewAndFooter() (*view, *tview.TextView) {
	return b.v, b.footer
}

func (b *testResourceViewBuilder) headerParamsBuilder() []headerPageParam {
	return b.headerParams
}

func (b *testResourceViewBuilder) headerPageItems(index int) []headerItem {
	_ = index
	return nil
}

func (b *testResourceViewBuilder) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	return b.title, b.headers, func() [][]string {
		return b.rows
	}
}

func TestResourceViewPreHandler(t *testing.T) {
	resources := []string{"resource"}

	t.Run("error input", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = TaskKind
		expectErr := errors.New("failed to load resources")

		gotErr := resourceViewPreHandler(resources, app, expectErr)
		if !errors.Is(gotErr, expectErr) {
			t.Errorf("Got: %v, Want: %v\n", gotErr, expectErr)
		}
		if app.kind != ServiceKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ServiceKind)
		}
	})

	t.Run("empty resources", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = ServiceKind

		gotErr := resourceViewPreHandler([]string{}, app, nil)
		if gotErr == nil {
			t.Errorf("Got nil error, Want non nil error\n")
		}
		if gotErr != nil && gotErr.Error() != "no services found" {
			t.Errorf("Got: %v, Want: %v\n", gotErr, "no services found")
		}
		if app.kind != ClusterKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ClusterKind)
		}
	})

	t.Run("success", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = ServiceKind

		gotErr := resourceViewPreHandler(resources, app, nil)
		if gotErr != nil {
			t.Errorf("Got: %v, Want: %v\n", gotErr, nil)
		}
		if app.kind != ServiceKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ServiceKind)
		}
	})
}

func TestBuildResourcePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = ClusterKind

		cluster := &types.Cluster{
			ClusterName: aws.String(clusterName1),
			ClusterArn:  aws.String(clusterArn1),
		}

		v := newView(app, basicKeyInputs, nil)
		v.originalRowReferences = []Entity{
			{
				cluster:    cluster,
				entityName: clusterArn1,
			},
		}
		footer := tview.NewTextView().SetDynamicColors(true)

		builder := &testResourceViewBuilder{
			v:      v,
			footer: footer,
			headerParams: []headerPageParam{
				{
					title:      clusterName1,
					entityName: clusterArn1,
					items: []headerItem{
						{name: "Name", value: clusterName1},
					},
				},
			},
			title:   clusterName1,
			headers: []string{"Name"},
			rows: [][]string{
				{clusterName1},
			},
		}

		err := buildResourcePage([]types.Cluster{*cluster}, app, nil, func() resourceViewBuilder {
			return builder
		})
		if err != nil {
			t.Errorf("Got: %v, Want: %v\n", err, nil)
		}
		if !app.Pages.HasPage(ClusterKind.getAppPageName("")) {
			t.Errorf("Page clusters should exist\n")
		}
		if v.mainFlex == nil {
			t.Errorf("mainFlex should be initialized\n")
		}
		if v.filterInput == nil {
			t.Errorf("filterInput should be initialized\n")
		}
		if v.table.GetRowCount() != 2 {
			t.Errorf("RowCount Got: %d, Want: %d\n", v.table.GetRowCount(), 2)
		}
		if app.cluster == nil || aws.ToString(app.cluster.ClusterArn) != clusterArn1 {
			t.Errorf("ClusterArn Got: %v, Want: %v\n", app.cluster, clusterArn1)
		}
		footerText := footer.GetText(false)
		if !strings.Contains(footerText, app.kind.String()) {
			t.Errorf("Footer Got: %s, Want contains: %s\n", footerText, app.kind.String())
		}
	})

	t.Run("pre handler error", func(t *testing.T) {
		app, _ := newApp(Option{})
		app.kind = TaskKind

		expectErr := errors.New("failed before builder")
		called := false
		err := buildResourcePage([]string{"resource"}, app, expectErr, func() resourceViewBuilder {
			called = true
			return nil
		})

		if !errors.Is(err, expectErr) {
			t.Errorf("Got: %v, Want: %v\n", err, expectErr)
		}
		if called {
			t.Errorf("builder should not be called when pre handler returns error\n")
		}
		if app.kind != ServiceKind {
			t.Errorf("Kind Got: %v, Want: %v\n", app.kind, ServiceKind)
		}
	})
}
