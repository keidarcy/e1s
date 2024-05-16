package color

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
)

var logger *slog.Logger

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

type themeConfig struct {
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
func InitStyles(theme string, appLogger *slog.Logger) Colors {
	logger = appLogger
	colors := Colors{
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
	colors.updateByTheme(theme)

	viper.UnmarshalKey("colors", &colors)
	logger.Debug("colors", "colors", colors)
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

func (c *Colors) updateByTheme(theme string) {
	if theme == "" {
		return
	}
	// Fetch the TOML content from the URL
	url := fmt.Sprintf("https://raw.githubusercontent.com/keidarcy/alacritty-theme/master/themes/%s.toml", theme)
	resp, err := http.Get(url)
	if err != nil {
		logger.Warn("failed fetching TOML data", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("failed fetching TOML data", "HTTP status code", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("failed reading TOML data", "error", err)
		return
	}

	var config themeConfig
	// Decode the TOML content
	if _, err := toml.Decode(string(body), &config); err != nil {
		logger.Warn("failed decoding TOML data", "error", err)
		return
	}
	c.BgColor = config.Colors.Primary.Background
	c.FgColor = config.Colors.Primary.Foreground
	c.BorderColor = config.Colors.Primary.Foreground
	c.Black = config.Colors.Normal.Black
	c.Yellow = config.Colors.Normal.Yellow
	c.Green = config.Colors.Normal.Green
	c.Red = config.Colors.Normal.Red
	c.Blue = config.Colors.Normal.Blue
	c.Magenta = config.Colors.Normal.Magenta
	c.Cyan = config.Colors.Normal.Cyan
}
