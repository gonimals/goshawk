package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"regexp"
	"time"

	yaml "gopkg.in/yaml.v3"

	"github.com/gonimals/goshawk/pkg/util"
)

func LoadConfig(configFile, expectedHash string) (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])

	if len(expectedHash) == 0 {
		slog.Warn("no expected hash provided, skipping hash check", "calculatedHash", actualHash)
	} else if actualHash[:len(expectedHash)] != expectedHash {
		slog.Warn("config hash mismatch", "calculatedHash", actualHash)
		return nil, fmt.Errorf("configuration hash mismatch")
	}

	return ParseConfigBytes(data)
}

func ParseConfigBytes(data []byte) (*Config, error) {
	var config Config
	var err error
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	// Handle incorrect or missing configurations
	if config.Services == nil {
		config.Services = make(map[string]*Service)
	}
	if config.AuthenticatedHosts == nil {
		config.AuthenticatedHosts = make(map[string]string)
	}
	if len(config.Services) == 0 && len(config.AuthenticatedHosts) == 0 {
		return nil, fmt.Errorf("configuration error: no services defined")
	}

	if config.ListenAddress == "" && len(config.AuthenticatedHosts) > 0 {
		return nil, fmt.Errorf("configuration error: authenticated hosts require a valid listen address")
	}

	config.TemplateTitleParsed, err = template.New("title").Parse(config.TemplateTitle)
	if err != nil {
		return nil, fmt.Errorf("could not parse template title: %w", err)
	}
	config.TemplateBodyParsed, err = template.New("body").Parse(config.TemplateBody)
	if err != nil {
		return nil, fmt.Errorf("could not parse template body down: %w", err)
	}

	// Create auxiliary maps
	config.HostKeys = make(map[string]string)
	for host, key := range config.AuthenticatedHosts {
		config.HostKeys[key] = host
	}

	config.HostsStatus = util.NewSyncMap[string, AssetStatus]()
	for host := range config.AuthenticatedHosts {
		config.HostsStatus.Set(host, AssetStatus{
			ServiceName: host,
			IsActive:    true,
			Notified:    true,
			LastCheck:   time.Now(),
		})
	}

	// Handle default values
	if config.DefaultServiceFrequency == 0 {
		config.DefaultServiceFrequency = 60
	}
	if config.DefaultServiceMaxFails == 0 {
		config.DefaultServiceMaxFails = 3
	}
	if config.DefaultServiceTimeout == 0 {
		config.DefaultServiceTimeout = 10
	}
	if config.NotificationRateLimit == 0 {
		config.NotificationRateLimit = 10
	}
	if config.HostMaxSeconds == 0 {
		config.HostMaxSeconds = 60
	}

	// Initialize services
	for serviceName, service := range config.Services {
		if service.FrequencySeconds == 0 {
			service.FrequencySeconds = config.DefaultServiceFrequency
		}
		if service.MaxFails == 0 {
			service.MaxFails = config.DefaultServiceMaxFails
		}
		if service.TimeoutSeconds == 0 {
			service.TimeoutSeconds = config.DefaultServiceTimeout
		}
		if service.WebRequest != nil {
			if service.WebRequest.ExpectedOutput == "" {
				service.WebRequest.ExpectedOutput = ".*"
			}
			service.WebRequest.ExpectedOutputRegexp, err = regexp.Compile(service.WebRequest.ExpectedOutput)
			if err != nil {
				return nil, fmt.Errorf("could not compile regexp: %w", err)
			}
		}
		if service.BashScript != nil {
			if service.BashScript.ExpectedOutput == "" {
				service.BashScript.ExpectedOutput = ".*"
			}
			service.BashScript.ExpectedOutputRegexp, err = regexp.Compile(service.BashScript.ExpectedOutput)
			if err != nil {
				return nil, fmt.Errorf("could not compile regexp: %w", err)
			}
		}
		service.Status = AssetStatus{
			ServiceName: serviceName,
			IsActive:    true,
			Notified:    true,
		}
	}

	return &config, nil
}
