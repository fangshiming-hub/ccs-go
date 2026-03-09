package validator

import (
	"encoding/json"
	"fmt"
	"os"
)

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid bool
	Errors  []string
	Warnings []string
}

// ValidateJSON validates if data is valid JSON
func ValidateJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return result, nil
}

// ValidateConfigFile validates the Claude config file format
func ValidateConfigFile(filePath string) (*ValidationResult, map[string]interface{}) {
	result := &ValidationResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("配置文件不存在: %s", filePath))
		return result, nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("无法读取文件: %s", err))
		return result, nil
	}

	// Parse JSON
	config, err := ValidateJSON(data)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("JSON 格式错误: %s", err))
		return result, nil
	}

	// Validate config structure
	return ValidateClaudeConfig(config)
}

// ValidateClaudeConfig validates the config structure
func ValidateClaudeConfig(config map[string]interface{}) (*ValidationResult, map[string]interface{}) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if config == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "配置必须是有效的 JSON 对象")
		return result, config
	}

	// Check if config is empty
	if len(config) == 0 {
		result.Warnings = append(result.Warnings, "配置文件为空，没有定义任何模型")
		return result, config
	}

	// Validate each model config
	for modelName, modelConfig := range config {
		envConfig, ok := modelConfig.(map[string]interface{})
		if !ok {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("模型 \"%s\" 的配置必须是对象", modelName))
			continue
		}

		// Check environment variables are strings
		for key, value := range envConfig {
			if _, isString := value.(string); !isString {
				result.IsValid = false
				result.Errors = append(result.Errors, fmt.Sprintf("模型 \"%s\" 的环境变量 \"%s\" 必须是字符串", modelName, key))
			}
		}
	}

	return result, config
}

// ValidateEnvConfig validates the env field in settings.json
func ValidateEnvConfig(envConfig map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if envConfig == nil {
		result.Warnings = append(result.Warnings, "当前没有设置 env 配置")
		return result
	}

	for key, value := range envConfig {
		if _, isString := value.(string); !isString {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("环境变量 \"%s\" 必须是字符串", key))
		}
	}

	return result
}

// GenerateValidationReport generates a formatted validation report
func GenerateValidationReport(result *ValidationResult, showContent bool, config map[string]interface{}) string {
	report := ""

	if result.IsValid {
		report += "✅ 配置验证通过\n"
	} else {
		report += "❌ 配置验证失败\n"
		for _, err := range result.Errors {
			report += fmt.Sprintf("  - 错误: %s\n", err)
		}
	}

	if len(result.Warnings) > 0 {
		report += "\n⚠️  配置警告\n"
		for _, warning := range result.Warnings {
			report += fmt.Sprintf("  - 警告: %s\n", warning)
		}
	}

	if showContent && config != nil {
		report += "\n📋 配置内容:\n"
		jsonData, _ := json.MarshalIndent(config, "", "  ")
		report += string(jsonData) + "\n"
	}

	return report
}