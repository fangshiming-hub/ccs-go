package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/weishiren/ccs-go/internal/config"
	"github.com/weishiren/ccs-go/internal/utils"
	"github.com/weishiren/ccs-go/internal/validator"
)

// Version is the application version
const Version = "1.0.0"

var (
	blue   = color.New(color.FgBlue).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	gray   = color.New(color.FgHiBlack).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

func main() {
	// Define flags
	listFlag := flag.Bool("list", false, "列出所有可用模型")
	listShort := flag.Bool("l", false, "列出所有可用模型 (简写)")
	infoFlag := flag.Bool("info", false, "显示当前 env 配置")
	infoShort := flag.Bool("i", false, "显示当前 env 配置 (简写)")
	validateFlag := flag.Bool("validate", false, "验证配置文件格式")
	validateShort := flag.Bool("V", false, "验证配置文件格式 (简写)")
	verboseFlag := flag.Bool("verbose", false, "详细输出")
	helpFlag := flag.Bool("help", false, "显示帮助信息")
	helpShort := flag.Bool("h", false, "显示帮助信息 (简写)")
	versionFlag := flag.Bool("version", false, "显示版本号")
	versionShort := flag.Bool("v", false, "显示版本号 (简写)")

	flag.Usage = printUsage
	flag.Parse()

	// Handle help
	if *helpFlag || *helpShort {
		printUsage()
		return
	}

	// Handle version
	if *versionFlag || *versionShort {
		fmt.Printf("ccs version %s\n", Version)
		return
	}

	// Create config manager
	cm, err := config.NewConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", red("❌ 错误:"), err.Error())
		os.Exit(1)
	}

	// Handle different commands
	if *listFlag || *listShort {
		handleListModels(cm)
	} else if *infoFlag || *infoShort {
		handleShowCurrentConfigInfo(cm)
	} else if *validateFlag || *validateShort {
		handleValidateConfig(cm, *verboseFlag)
	} else {
		// Check for model name argument
		args := flag.Args()
		if len(args) > 0 {
			handleSwitchByModel(cm, args[0])
		} else {
			handleInteractiveMode(cm)
		}
	}
}

func printUsage() {
	fmt.Println("用法: ccs [模型名称] [选项]")
	fmt.Println()
	fmt.Println("Claude 配置切换器 - 快速切换不同的 Claude 环境配置")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  ccs              交互式选择模型")
	fmt.Println("  ccs <model>      直接切换到指定模型 (如: ccs work)")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -l, --list       列出所有可用模型")
	fmt.Println("  -i, --info       显示当前 env 配置")
	fmt.Println("  -V, --validate   验证配置文件格式")
	fmt.Println("      --verbose    详细输出")
	fmt.Println("  -h, --help       显示帮助信息")
	fmt.Println("  -v, --version    显示版本号")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  ccs              # 交互式选择")
	fmt.Println("  ccs work         # 切换到 work 环境")
	fmt.Println("  ccs -l           # 列出所有模型")
	fmt.Println("  ccs -i           # 查看当前配置")
	fmt.Println("  ccs -V           # 验证配置文件")
}

func handleInteractiveMode(cm *config.ConfigManager) {
	fmt.Println(blue("🤖 Claude 配置切换器"))
	fmt.Println(gray("====================="))
	fmt.Println()

	models, err := cm.ScanConfigs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", red("❌ 错误:"), err.Error())
		os.Exit(1)
	}

	if len(models) == 0 {
		fmt.Println(yellow("未找到模型配置"))
		fmt.Println(gray(fmt.Sprintf("配置文件: %s", cm.GetConfigFilePath())))
		fmt.Println(gray("\n请创建配置文件，格式如下:"))
		exampleConfig := map[string]interface{}{
			"work": map[string]interface{}{
				"ANTHROPIC_API_KEY":  "your-key",
				"ANTHROPIC_BASE_URL": "https://api.anthropic.com",
			},
			"personal": map[string]interface{}{
				"ANTHROPIC_API_KEY":  "your-key",
				"ANTHROPIC_BASE_URL": "https://custom-api.example.com",
			},
		}
		jsonData, _ := json.MarshalIndent(exampleConfig, "", "  ")
		fmt.Println(green(string(jsonData)))
		return
	}

	fmt.Println(blue(fmt.Sprintf("📁 配置文件: %s", cm.GetConfigFilePath())))
	fmt.Println(blue(fmt.Sprintf("📊 找到 %d 个模型配置:", len(models))))
	fmt.Println()

	// Show current config
	currentConfig, err := cm.GetCurrentConfig()
	if err == nil && currentConfig != nil && currentConfig.Env != nil {
		fmt.Println(cyan("⚡ 当前 env 配置:"))
		jsonData, _ := json.MarshalIndent(currentConfig.Env, "", "  ")
		fmt.Println(gray(string(jsonData)))
		fmt.Println()
	}

	// Build choices for prompt
	var items []string
	for _, model := range models {
		items = append(items, model.Name)
	}

	prompt := promptui.Select{
		Label: "请选择要切换的模型",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		if strings.Contains(err.Error(), "interrupt") {
			fmt.Println(yellow("取消操作"))
			return
		}
		fmt.Fprintf(os.Stderr, "%s %s\n", red("❌ 错误:"), err.Error())
		os.Exit(1)
	}

	doSwitchConfig(cm, models[idx].Name)
}

func handleListModels(cm *config.ConfigManager) {
	models, err := cm.ScanConfigs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", red("❌ 错误:"), err.Error())
		os.Exit(1)
	}

	fmt.Println(blue(fmt.Sprintf("📁 配置文件: %s", cm.GetConfigFilePath())))

	if len(models) == 0 {
		fmt.Println(yellow("\n未找到模型配置"))
		fmt.Println(gray("\n请创建配置文件，格式如下:"))
		exampleConfig := map[string]interface{}{
			"work": map[string]interface{}{
				"ANTHROPIC_API_KEY":  "your-key",
				"ANTHROPIC_BASE_URL": "https://api.anthropic.com",
			},
			"personal": map[string]interface{}{
				"ANTHROPIC_API_KEY":  "your-key",
				"ANTHROPIC_BASE_URL": "https://custom-api.example.com",
			},
		}
		jsonData, _ := json.MarshalIndent(exampleConfig, "", "  ")
		fmt.Println(green(string(jsonData)))
		return
	}

	fmt.Println(blue(fmt.Sprintf("\n📊 找到 %d 个模型配置:", len(models))))
	fmt.Println()

	for i, model := range models {
		fmt.Printf("%d. %s\n", i+1, green(model.Name))
	}

	// Show current config
	currentConfig, err := cm.GetCurrentConfig()
	if err == nil && currentConfig != nil && currentConfig.Env != nil {
		fmt.Println(cyan("\n⚡ 当前 env 配置:"))
		jsonData, _ := json.MarshalIndent(currentConfig.Env, "", "  ")
		fmt.Println(gray(string(jsonData)))
	}
}

func handleSwitchByModel(cm *config.ConfigManager, modelName string) {
	doSwitchConfig(cm, modelName)
}

func doSwitchConfig(cm *config.ConfigManager, modelName string) {
	fmt.Println(blue("🔄 正在切换配置..."))
	fmt.Println(gray(fmt.Sprintf("模型: %s", modelName)))

	err := cm.SwitchConfig(modelName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", red("❌ 错误:"), err.Error())
		os.Exit(1)
	}

	fmt.Println(green(fmt.Sprintf("✅ 已切换到模型 \"%s\"", modelName)))
	fmt.Println(gray(fmt.Sprintf("配置文件: %s", cm.GetTargetFilePath())))
}

func handleValidateConfig(cm *config.ConfigManager, verbose bool) {
	configPath := cm.GetConfigFilePath()

	fmt.Println(blue(fmt.Sprintf("📁 验证配置文件: %s", configPath)))
	fmt.Println()

	if !utils.FileExists(configPath) {
		fmt.Println(yellow("⚠️  配置文件不存在"))
		fmt.Println(gray("\n请创建配置文件，格式如下:"))
		exampleConfig := map[string]interface{}{
			"work": map[string]interface{}{
				"ANTHROPIC_API_KEY":  "your-key",
				"ANTHROPIC_BASE_URL": "https://api.anthropic.com",
			},
			"personal": map[string]interface{}{
				"ANTHROPIC_API_KEY":  "your-key",
				"ANTHROPIC_BASE_URL": "https://custom-api.example.com",
			},
		}
		jsonData, _ := json.MarshalIndent(exampleConfig, "", "  ")
		fmt.Println(green(string(jsonData)))
		return
	}

	result, configData := validator.ValidateConfigFile(configPath)
	report := validator.GenerateValidationReport(result, verbose, configData)
	fmt.Println(report)
}

func handleShowCurrentConfigInfo(cm *config.ConfigManager) {
	currentConfig, err := cm.GetCurrentConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", red("❌ 错误:"), err.Error())
		os.Exit(1)
	}

	if currentConfig == nil {
		fmt.Println(yellow("⚠️  未找到当前配置文件"))
		fmt.Println(gray(fmt.Sprintf("目标文件: %s", cm.GetTargetFilePath())))
		return
	}

	fmt.Println(blue("\n📄 当前配置详情"))
	fmt.Println(gray("=================="))
	fmt.Println(green(fmt.Sprintf("文件: %s", currentConfig.Name)))
	fmt.Println(gray(fmt.Sprintf("路径: %s", currentConfig.Path)))
	fmt.Println(gray(fmt.Sprintf("大小: %s", utils.FormatFileSize(currentConfig.Size))))
	fmt.Println(gray(fmt.Sprintf("修改时间: %s", currentConfig.Modified.Format("2006-01-02 15:04:05"))))

	if currentConfig.Env != nil {
		fmt.Println(blue("\n📋 env 配置:"))
		jsonData, _ := json.MarshalIndent(currentConfig.Env, "", "  ")
		fmt.Println(green(string(jsonData)))

		// Validate env config
		envMap := make(map[string]interface{})
		for k, v := range currentConfig.Env {
			envMap[k] = v
		}
		validation := validator.ValidateEnvConfig(envMap)
		fmt.Println(blue("\n🔍 配置验证:"))
		if validation.IsValid {
			fmt.Println(green("✅ env 配置有效"))
		} else {
			fmt.Println(red("❌ env 配置存在以下问题:"))
			for _, e := range validation.Errors {
				fmt.Println(red(fmt.Sprintf("  - %s", e)))
			}
		}

		if len(validation.Warnings) > 0 {
			fmt.Println(yellow("\n⚠️  警告:"))
			for _, w := range validation.Warnings {
				fmt.Println(yellow(fmt.Sprintf("  - %s", w)))
			}
		}
	} else {
		fmt.Println(yellow("\n⚠️  当前配置文件中没有 env 字段"))
	}

	fmt.Println()
}