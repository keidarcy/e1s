package api

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/account"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	regionsLib "github.com/keidarcy/aws-regions/v3"
)

// SwitchAwsConfig switches to a different AWS profile and reinitializes clients
func (store *Store) SwitchAwsConfig(profile string, region string) error {
	os.Setenv("AWS_PROFILE", profile)
	os.Setenv("AWS_REGION", region)

	// Load new configuration with the updated profile
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("failed to load aws SDK config with new profile", "profile", profile, "error", err)
		return err
	}

	// Update store with new configuration
	store.Config = &cfg

	// Reinitialize all clients with new configuration
	store.ecs = ecs.NewFromConfig(cfg)
	store.cloudwatch = nil     // Will be lazy-loaded with new config
	store.cloudwatchlogs = nil // Will be lazy-loaded with new config
	store.ssm = nil            // Will be lazy-loaded with new config
	store.autoScaling = nil    // Will be lazy-loaded with new config
	store.account = nil

	slog.Info("switched AWS profile", slog.String("AWS_PROFILE", profile), slog.String("AWS_REGION", region))
	return nil
}

// Profile summarizes one named profile from local AWS config and credentials files.
type Profile struct {
	Name          string
	Source        string // config, credentials, or both
	DefaultRegion string
	AuthStyle     string
}

func (store *Store) ListProfiles() ([]Profile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := os.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = filepath.Join(homeDir, ".aws", "config")
	}
	configSections := parseAWSConfigProfiles(configPath)

	credentialsPath := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if credentialsPath == "" {
		credentialsPath = filepath.Join(homeDir, ".aws", "credentials")
	}
	credNames, credStatic := parseAWSCredentialsProfiles(credentialsPath)

	nameSet := make(map[string]struct{})
	for n := range configSections {
		nameSet[n] = struct{}{}
	}
	for n := range credNames {
		nameSet[n] = struct{}{}
	}
	if _, ok := nameSet["default"]; !ok {
		nameSet["default"] = struct{}{}
	}

	names := make([]string, 0, len(nameSet))
	for n := range nameSet {
		names = append(names, n)
	}
	sort.Strings(names)

	out := make([]Profile, 0, len(names))
	for _, name := range names {
		_, inConfig := configSections[name]
		inCreds := credNames[name]
		p := Profile{Name: name}

		switch {
		case inConfig && inCreds:
			p.Source = "both"
		case inConfig:
			p.Source = "config"
		default:
			p.Source = "credentials"
		}

		if kv := configSections[name]; kv != nil {
			p.DefaultRegion = strings.TrimSpace(kv["region"])
			p.AuthStyle = inferAuthStyle(kv, credStatic[name])
		} else {
			if credStatic[name] {
				p.AuthStyle = "static"
			} else {
				p.AuthStyle = "—"
			}
		}

		out = append(out, p)
	}
	return out, nil
}

func inferAuthStyle(kv map[string]string, credStatic bool) string {
	if kv == nil {
		if credStatic {
			return "static"
		}
		return "—"
	}
	if strings.TrimSpace(kv["sso_start_url"]) != "" || strings.TrimSpace(kv["sso_session"]) != "" {
		return "SSO"
	}
	if strings.TrimSpace(kv["web_identity_token_file"]) != "" {
		return "web identity"
	}
	if strings.TrimSpace(kv["credential_process"]) != "" {
		return "credential process"
	}
	if strings.TrimSpace(kv["role_arn"]) != "" {
		return "assume role"
	}
	if credStatic {
		return "static"
	}
	return "inherit"
}

// parseAWSConfigProfiles returns per-profile key/value settings from ~/.aws/config (keys lowercased).
func parseAWSConfigProfiles(path string) map[string]map[string]string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	sections := make(map[string]map[string]string)
	var current string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			header := strings.TrimSpace(line[1 : len(line)-1])
			switch {
			case header == "default":
				current = "default"
			case strings.HasPrefix(strings.ToLower(header), "profile "):
				current = strings.TrimSpace(header[len("profile "):])
			default:
				current = ""
			}
			if current != "" && sections[current] == nil {
				sections[current] = make(map[string]string)
			}
			continue
		}
		if current == "" {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		sections[current][key] = val
	}
	return sections
}

func parseAWSCredentialsProfiles(path string) (names map[string]bool, staticKeys map[string]bool) {
	names = make(map[string]bool)
	staticKeys = make(map[string]bool)

	file, err := os.Open(path)
	if err != nil {
		return names, staticKeys
	}
	defer file.Close()

	var current string
	var hasAccessID, hasSecret bool
	flush := func() {
		if current == "" {
			return
		}
		names[current] = true
		if hasAccessID || hasSecret {
			staticKeys[current] = true
		}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			flush()
			current = strings.TrimSpace(line[1 : len(line)-1])
			hasAccessID, hasSecret = false, false
			continue
		}
		if current == "" {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		switch key {
		case "aws_access_key_id":
			if val != "" {
				hasAccessID = true
			}
		case "aws_secret_access_key":
			if val != "" {
				hasSecret = true
			}
		}
	}
	flush()
	return names, staticKeys
}

type Region struct {
	Code    string
	Name    string
	Enabled string
}

func (store *Store) ListRegions() ([]Region, error) {
	store.initAccountClient()
	limit := int32(50)
	regionsOutput, err := store.account.ListRegions(context.Background(), &account.ListRegionsInput{MaxResults: &limit})
	if err != nil {
		slog.Warn("failed to run aws api list regions", "error", err)
		return nil, err
	}
	regions := regionsOutput.Regions

	regionsList, err := regionsLib.List()
	if err != nil {
		return []Region{}, nil
	}

	results := []Region{}
	for _, region := range regions {
		result := Region{}
		for _, r := range regionsList {
			if r.Code == *region.RegionName {
				result.Code = r.Code
				result.Name = r.FullName
				enabled := "Enabled by default"
				if string(region.RegionOptStatus) == "DISABLED" {
					enabled = "Disabled"
				}
				result.Enabled = enabled
				break
			}
		}
		if result.Code != "" {
			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Enabled == "Disabled" {
			return false
		}
		return true
	})

	return results, nil
}
