package color

import (
	"fmt"
	"io"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logger *logrus.Logger

type Colors struct {
	BgColor     string `toml:"background"`
	FgColor     string `toml:"foreground"`
	BorderColor string `toml:"border"` // Assuming there's a border color in the TOML
	Black       string `toml:"black"`
	Yellow      string `toml:"yellow"`
	Green       string `toml:"green"`
	Red         string `toml:"red"`
	Blue        string `toml:"blue"`
	Magenta     string `toml:"magenta"`
	Cyan        string `toml:"cyan"`
	Gray        string `toml:"white"` // Assuming dim white is considered gray
}

type tomlConfig struct {
	Colors struct {
		Primary struct {
			Background string `toml:"background"`
			Foreground string `toml:"foreground"`
		} `toml:"primary"`
		Normal struct {
			Black   string `toml:"black"`
			Red     string `toml:"red"`
			Green   string `toml:"green"`
			Yellow  string `toml:"yellow"`
			Blue    string `toml:"blue"`
			Magenta string `toml:"magenta"`
			Cyan    string `toml:"cyan"`
			White   string `toml:"white"`
		} `toml:"normal"`
	} `toml:"colors"`
}

func Color(c string) tcell.Color {
	if c == "default" {
		return tcell.ColorDefault
	}
	return tcell.GetColor(c).TrueColor()
}

// Init basic styles
func InitStyles(theme string) Colors {
	var colors Colors
	if theme != "" {
		// Fetch the TOML content from the URL
		url := fmt.Sprintf("https://raw.githubusercontent.com/keidarcy/alacritty-theme/master/themes/%s.toml", theme)
		resp, err := http.Get(url)
		if err != nil {
			logger.Warnf("Error fetching TOML data: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warnf("Error fetching TOML data: HTTP status code %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Warnf("Error reading TOML data: %s", err)
		}

		// Decode the TOML content
		var config tomlConfig
		if _, err := toml.Decode(string(body), &config); err != nil {
			logger.Warnf("Error decoding TOML data: %s", err)
		}

		// Map the decoded TOML to the Theme struct
		colors = Colors{
			BgColor:     config.Colors.Primary.Background,
			FgColor:     config.Colors.Primary.Foreground,
			BorderColor: config.Colors.Primary.Foreground, // Assuming border is the same as foreground
			Black:       config.Colors.Normal.Black,
			Yellow:      config.Colors.Normal.Yellow,
			Green:       config.Colors.Normal.Green,
			Red:         config.Colors.Normal.Red,
			Blue:        config.Colors.Normal.Blue,
			Magenta:     config.Colors.Normal.Magenta,
			Cyan:        config.Colors.Normal.Cyan,
		}
	} else {
		// Default theme gruvbox_dark
		// Get from https://github.com/alacritty/alacritty-colors/blob/master/themes/gruvbox_dark.toml
		colors = Colors{
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
		}
	}

	colors.Gray = "#808080"

	viper.UnmarshalKey("colors", &colors)
	tview.Styles.PrimitiveBackgroundColor = Color(colors.BgColor)
	tview.Styles.ContrastBackgroundColor = Color(colors.BgColor)
	tview.Styles.MoreContrastBackgroundColor = Color(colors.BgColor)
	tview.Styles.PrimaryTextColor = Color(colors.FgColor)
	tview.Styles.BorderColor = Color(colors.BorderColor)
	tview.Styles.TitleColor = Color(colors.FgColor)
	tview.Styles.GraphicsColor = Color(colors.FgColor)
	tview.Styles.SecondaryTextColor = Color(colors.FgColor)
	tview.Styles.TertiaryTextColor = Color(colors.FgColor)
	tview.Styles.InverseTextColor = Color(colors.FgColor)
	tview.Styles.ContrastSecondaryTextColor = Color(colors.FgColor)

	colors.initFmt()

	return colors
}
