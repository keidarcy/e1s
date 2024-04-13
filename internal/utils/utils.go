package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	EmptyText     = "<empty>"
	greenFmt      = "[green]%s[-:-:-]"
	greyFmt       = "[grey]%s[-:-:-]"
	clusterFmt    = "https://%s.console.aws.amazon.com/ecs/v2/clusters/%s"
	regionFmt     = "?region=%s"
	serviceFmt    = "/services/%s"
	taskFmt       = "/tasks/%s"
	clusterURLFmt = clusterFmt + regionFmt
	serviceURLFmt = clusterFmt + serviceFmt + regionFmt
	taskURLFmt    = clusterFmt + serviceFmt + taskFmt + regionFmt

	AppVersion = "v1.0.26"
	AppName    = "e1s"
)

// GetLogger returns a *logrus.Logger configured to write to the specified file path.
// It also returns the log file *os.File  itself, such that callers can close the
// file if/when needed.
func GetLogger(path string, json bool, debug bool) (*logrus.Logger, *os.File) {
	logger := logrus.New()

	if debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.SetOutput(file)
	} else {
		logger.Error("Failed to log to file, using default stderr")
	}

	if json {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339, // Customize the timestamp format
		})
	} else {
		// Add colored output to the console with a custom timestamp format
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339, // Customize the timestamp format
		})
	}

	// https://github.com/sirupsen/logrus?tab=readme-ov-file#thread-safety
	logger.SetNoLock()

	return logger, file
}

func ArnToName(arn *string) string {
	if arn == nil {
		return EmptyText
	}
	ss := strings.Split(*arn, "/")
	return ss[len(ss)-1]
}

func ArnToFullName(arn *string) string {
	if arn == nil {
		return EmptyText
	}
	ss := strings.Split(*arn, ":")
	return ss[len(ss)-1]
}

func ShowString(s *string) string {
	if s == nil {
		return EmptyText
	}
	return *s
}

func ShowArray(s []string) string {
	if len(s) == 0 {
		return EmptyText
	}
	return strings.Join(s, ",")
}

func ShowTime(at *time.Time) string {
	if at == nil {
		return EmptyText
	}
	return at.Format(time.RFC3339)
}

func ShowInt(p *int32) string {
	if p == nil {
		return EmptyText
	}
	return strconv.Itoa(int(*p))
}

func ShowGreenGrey(inputStr *string, greenStr string) string {
	if inputStr == nil {
		return EmptyText
	}

	str := *inputStr
	if str == "" {
		return EmptyText
	}
	outputStr := strings.ToUpper(string(str[0])) + strings.ToLower(string(str[1:]))
	if strings.ToLower(str) == greenStr {
		return fmt.Sprintf(greenFmt, outputStr)
	}
	return fmt.Sprintf(greyFmt, outputStr)
}

// Convert ARN to AWS web console URL
// TaskARN not contains service but need service name as second argument
func ArnToUrl(arn string, taskService string) string {
	components := strings.Split(arn, ":")
	resources := components[len(components)-1]
	names := strings.Split(resources, "/")

	region := components[3]
	clusterName := ""
	serviceName := ""
	taskName := ""

	switch names[0] {
	case "cluster":
		clusterName = names[1]
		return fmt.Sprintf(clusterURLFmt, region, clusterName, region)
	case "service":
		clusterName = names[1]
		serviceName = names[2]
		return fmt.Sprintf(serviceURLFmt, region, clusterName, serviceName, region)
	case "task", "container":
		clusterName = names[1]
		taskName = names[2]
		return fmt.Sprintf(taskURLFmt, region, clusterName, taskService, taskName, region)
	default:
		return ""
	}
}

func OpenURL(url string) error {
	var err error

	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		return err
	}
	return nil
}

func BuildMeterText(f float64) string {
	const yesBlock = "█"
	const noBlock = "▒"
	i := int(f)

	yesNum := i / 5
	if yesNum == 0 {
		yesNum++
	}
	noNum := 20 - yesNum
	meterVal := strings.Join([]string{
		strings.Repeat(yesBlock, yesNum),
		strings.Repeat(noBlock, noNum),
	}, "")

	return meterVal + " " + fmt.Sprintf("%.2f", f) + "%"
}

func ShowVersion() string {
	type ghRes struct {
		Name string `json:"name"`
	}
	resp, err := http.Get("https://api.github.com/repos/keidarcy/e1s/releases/latest")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var rsp ghRes
	if err := json.Unmarshal(body, &rsp); err != nil {
		log.Fatal(err)
	}
	latestVersion := rsp.Name

	message := ""
	if latestVersion != AppVersion {
		message = "\nPlease upgrade e1s to latest version on https://github.com/keidarcy/e1s/releases"
	}

	return fmt.Sprintf("\nCurrent: %s\nLatest: %s%s", AppVersion, latestVersion, message)
}

func Age(t *time.Time) string {
	if t == nil {
		return EmptyText
	}
	now := time.Now()
	if now.Before(*t) {
		return "0s"
	}
	duration := now.Sub(*t)

	if years := int(duration.Hours() / 24 / 365); years > 0 {
		return fmt.Sprintf("%dy", years)
	}
	if months := int(duration.Hours() / 24 / 30); months > 0 {
		return fmt.Sprintf("%dmo", months)
	}
	if weeks := int(duration.Hours() / 24 / 7); weeks > 0 {
		return fmt.Sprintf("%dw", weeks)
	}
	if days := int(duration.Hours() / 24); days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	if hours := int(duration.Hours()); hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if minutes := int(duration.Minutes()); minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", int(duration.Seconds()))
}

// Return docker image registry and image name with tag
func ImageInfo(imageURL *string) (string, string) {
	if imageURL == nil {
		return EmptyText, EmptyText
	}
	url := *imageURL
	// Map of known registry domains to their short names
	registryMap := map[string]string{
		"docker.io":           "Docker Hub",
		"ecr.aws":             "Amazon ECR Public",
		".amazonaws.com":      "Amazon ECR",
		"gcr.io":              "Google GCR",
		"azurecr.io":          "Azure ACR",
		"registry.gitlab.com": "GitLab",
		"ghcr.io":             "GitHub",
		"quay.io":             "Quay",
	}

	// Default registry short name
	defaultRegistry := "Docker Hub"

	// Extract the domain and path from the image URL
	domain := url
	path := ""
	if strings.Contains(url, "/") {
		parts := strings.SplitN(url, "/", 2)
		domain = parts[0]
		path = parts[1]
	} else {
		// If there's no '/', it's an official image on Docker Hub
		path = url
		domain = "docker.io"
	}

	// Check for known registries by domain
	registryShortName := defaultRegistry
	for key, shortName := range registryMap {
		if strings.Contains(domain, key) {
			registryShortName = shortName
			break
		}
	}

	if strings.Contains(path, ":") {
		parts := strings.SplitN(path, ":", 2)
		imageName := parts[0]
		tag := parts[1]
		if len(tag) > 8 {
			tag = tag[:8] + "..."
		}
		path = imageName + ":" + tag
	}

	// Return the registry short name and the image name with tag
	return registryShortName, path
}
