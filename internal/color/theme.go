package color

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
)

type Theme struct {
	BgColor     string
	FgColor     string
	BorderColor string
	Black       string
	Yellow      string
	Green       string
	Red         string
	Blue        string
	Magenta     string
	Cyan        string
	Gray        string
}

func Color(c string) tcell.Color {
	if c == "default" {
		return tcell.ColorDefault
	}
	return tcell.GetColor(c).TrueColor()
}

// Init basic styles
func InitStyles() Theme {
	// Get from https://github.com/alacritty/alacritty-theme/blob/master/themes/gruvbox_dark.toml
	theme := Theme{
		BgColor:     "#282828",
		FgColor:     "#ebdbb2",
		BorderColor: "#8ec07c",
		Black:       "#282828",
		Red:         "#cc241d",
		Green:       "#98971a",
		Yellow:      "#d79921",
		Blue:        "#458588",
		Magenta:     "#b16286",
		Cyan:        "#689d6a",
		Gray:        "#808080",
	}
	viper.UnmarshalKey("theme", &theme)
	tview.Styles.PrimitiveBackgroundColor = Color(theme.BgColor)
	tview.Styles.ContrastBackgroundColor = Color(theme.BgColor)
	tview.Styles.MoreContrastBackgroundColor = Color(theme.BgColor)
	tview.Styles.PrimaryTextColor = Color(theme.FgColor)
	tview.Styles.BorderColor = Color(theme.BorderColor)
	tview.Styles.TitleColor = Color(theme.FgColor)
	tview.Styles.GraphicsColor = Color(theme.FgColor)
	tview.Styles.SecondaryTextColor = Color(theme.FgColor)
	tview.Styles.TertiaryTextColor = Color(theme.FgColor)
	tview.Styles.InverseTextColor = Color(theme.FgColor)
	tview.Styles.ContrastSecondaryTextColor = Color(theme.FgColor)

	theme.initFmt()

	return theme
}
