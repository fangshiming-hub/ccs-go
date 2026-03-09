package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/weishiren/ccs-go/internal/utils"
)

// ModelConfig represents a single model configuration
type ModelConfig struct {
	Name string
	Path string
}

// CurrentConfigInfo represents information about the current config
type CurrentConfigInfo struct {
	Name     string
	Path     string
	Size     int64
	Modified time.Time
	Env      map[string]interface{}
}

// ConfigManager manages the claude configuration
type ConfigManager struct {
	configDir  string
	configFile string
	targetFile string
}

// NewConfigManager creates a new ConfigManager
func NewConfigManager() (*ConfigManager, error) {
	configDir, err := utils.GetConfigDir()
	if err != nil {
		return nil, err
	}

	configFile, err := utils.GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	targetFile, err := utils.GetSettingsFilePath()
	if err != nil {
		return nil, err
	}

	return &ConfigManager{
		configDir:  configDir,
		configFile: configFile,
		targetFile: targetFile,
	}, nil
}

// EnsureConfigFile ensures the config file exists
func (cm *ConfigManager) EnsureConfigFile() error {
	if err := utils.EnsureDir(cm.configDir); err != nil {
		return err
	}

	if !utils.FileExists(cm.configFile) {
		// Create empty config
		emptyConfig := map[string]interface{}{}
		data, err := json.MarshalIndent(emptyConfig, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(cm.configFile, data, 0644)
	}

	return nil
}

// ScanConfigs scans the config file and returns a list of models
func (cm *ConfigManager) ScanConfigs() ([]ModelConfig, error) {
	if err := cm.EnsureConfigFile(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("配置文件 JSON 格式错误: %w", err)
	}

	var models []ModelConfig
	for name := range config {
		models = append(models, ModelConfig{
			Name: name,
			Path: cm.configFile,
		})
	}

	return models, nil
}

// GetEnvConfig gets the env config for a specific model
func (cm *ConfigManager) GetEnvConfig(modelName string) (map[string]interface{}, error) {
	if err := cm.EnsureConfigFile(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	envConfig, exists := config[modelName]
	if !exists {
		availableModels := make([]string, 0, len(config))
		for name := range config {
			availableModels = append(availableModels, name)
		}
		availableStr := "无"
		if len(availableModels) > 0 {
			availableStr = ""
			for i, name := range availableModels {
				if i > 0 {
					availableStr += ", "
				}
				availableStr += name
			}
		}
		return nil, fmt.Errorf("找不到模型 \"%s\"\n可用的模型: %s", modelName, availableStr)
	}

	envMap, ok := envConfig.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("模型 \"%s\" 的配置格式错误", modelName)
	}

	return envMap, nil
}

// ReadSettings reads the settings.json file
func (cm *ConfigManager) ReadSettings() (map[string]interface{}, error) {
	if !utils.FileExists(cm.targetFile) {
		return map[string]interface{}{}, nil
	}

	data, err := os.ReadFile(cm.targetFile)
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return map[string]interface{}{}, nil
	}

	return settings, nil
}

// SwitchConfig switches to the specified model
func (cm *ConfigManager) SwitchConfig(modelName string) error {
	envConfig, err := cm.GetEnvConfig(modelName)
	if err != nil {
		return err
	}

	settings, err := cm.ReadSettings()
	if err != nil {
		return err
	}

	settings["env"] = envConfig

	// Ensure target directory exists
	targetDir := filepath.Dir(cm.targetFile)
	if err := utils.EnsureDir(targetDir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.targetFile, data, 0644)
}

// GetCurrentConfig returns the current config information
func (cm *ConfigManager) GetCurrentConfig() (*CurrentConfigInfo, error) {
	if !utils.FileExists(cm.targetFile) {
		return nil, nil
	}

	info, err := os.Stat(cm.targetFile)
	if err != nil {
		return nil, err
	}

	settings, err := cm.ReadSettings()
	if err != nil {
		return nil, err
	}

	var env map[string]interface{}
	if envValue, exists := settings["env"]; exists {
		if envMap, ok := envValue.(map[string]interface{}); ok {
			env = envMap
		}
	}

	return &CurrentConfigInfo{
		Name:     "settings.json",
		Path:     cm.targetFile,
		Size:     info.Size(),
		Modified: info.ModTime(),
		Env:      env,
	}, nil
}

// GetConfigFilePath returns the config file path
func (cm *ConfigManager) GetConfigFilePath() string {
	return cm.configFile
}

// GetTargetFilePath returns the target file path
func (cm *ConfigManager) GetTargetFilePath() string {
	return cm.targetFile
}

// GetConfigDir returns the config directory path
func (cm *ConfigManager) GetConfigDir() string {
	return cm.configDir
}