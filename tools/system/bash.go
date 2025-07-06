package system

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"opencode_nano/tools/core"
)

// BashTool 增强版 bash 执行工具
type BashTool struct {
	*core.BaseTool
}

// NewBashTool 创建 bash 工具
func NewBashTool() *BashTool {
	tool := &BashTool{
		BaseTool: core.NewBaseTool("bash", "system", "Execute shell commands with enhanced features"),
	}
	
	tool.SetRequiresPerm(true)
	tool.SetTags("system", "shell", "command", "execute")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"command": {
				Type:        "string",
				Description: "Command to execute",
			},
			"cwd": {
				Type:        "string",
				Description: "Working directory",
				Default:     "",
			},
			"env": {
				Type:        "object",
				Description: "Environment variables",
				Default:     map[string]string{},
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout in seconds (0 for no timeout)",
				Default:     300, // 5 minutes default
			},
			"shell": {
				Type:        "string",
				Description: "Shell to use",
				Default:     "",
			},
			"capture_output": {
				Type:        "boolean",
				Description: "Capture command output",
				Default:     true,
			},
			"combine_output": {
				Type:        "boolean",
				Description: "Combine stdout and stderr",
				Default:     true,
			},
		},
		Required: []string{"command"},
	})
	
	return tool
}

// Execute 执行命令
func (t *BashTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取参数
	command, err := params.GetString("command")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid command parameter")
	}
	
	// 安全检查
	if err := t.checkCommandSafety(command); err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("unsafe command: %v", err))
	}
	
	// 获取可选参数
	cwd := ""
	if params.Has("cwd") {
		cwd, _ = params.GetString("cwd")
		// 验证目录存在
		if cwd != "" {
			if info, err := os.Stat(cwd); err != nil || !info.IsDir() {
				return nil, core.ErrInvalidParams(t.Info().Name, "invalid working directory")
			}
		}
	}
	
	env := make(map[string]string)
	if params.Has("env") {
		if envRaw, err := params.Get("env"); err == nil {
			if envMap, ok := envRaw.(map[string]interface{}); ok {
				for k, v := range envMap {
					if s, ok := v.(string); ok {
						env[k] = s
					}
				}
			}
		}
	}
	
	timeout := 300
	if params.Has("timeout") {
		timeout, _ = params.GetInt("timeout")
	}
	
	shell := t.getShell()
	if params.Has("shell") {
		if customShell, _ := params.GetString("shell"); customShell != "" {
			shell = customShell
		}
	}
	
	captureOutput := true
	if params.Has("capture_output") {
		captureOutput, _ = params.GetBool("capture_output")
	}
	
	combineOutput := true
	if params.Has("combine_output") {
		combineOutput, _ = params.GetBool("combine_output")
	}
	
	// 创建命令
	var cmd *exec.Cmd
	if timeout > 0 {
		// 使用超时上下文
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(timeoutCtx, shell, "-c", command)
	} else {
		cmd = exec.CommandContext(ctx, shell, "-c", command)
	}
	
	// 设置工作目录
	if cwd != "" {
		cmd.Dir = cwd
	}
	
	// 设置环境变量
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	
	// 执行命令
	var stdout, stderr bytes.Buffer
	startTime := time.Now()
	
	if captureOutput {
		if combineOutput {
			cmd.Stdout = &stdout
			cmd.Stderr = &stdout
		} else {
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
		}
	}
	
	err = cmd.Run()
	duration := time.Since(startTime)
	
	// 创建结果
	var resultMsg string
	exitCode := 0
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			resultMsg = fmt.Sprintf("Command failed with exit code %d", exitCode)
		} else if ctx.Err() == context.DeadlineExceeded {
			resultMsg = "Command timed out"
			exitCode = -1
		} else {
			resultMsg = fmt.Sprintf("Command failed: %v", err)
			exitCode = -1
		}
	} else {
		resultMsg = "Command executed successfully"
	}
	
	result := core.NewSimpleResult(resultMsg)
	result.WithMetadata("command", command)
	result.WithMetadata("exit_code", exitCode)
	result.WithMetadata("duration_ms", duration.Milliseconds())
	
	if captureOutput {
		result.WithMetadata("stdout", stdout.String())
		if !combineOutput {
			result.WithMetadata("stderr", stderr.String())
		}
		
		// 添加输出到结果数据
		if combineOutput || stderr.Len() == 0 {
			result = core.NewSimpleResult(stdout.String())
		} else {
			result = core.NewSimpleResult(fmt.Sprintf("stdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String()))
		}
		
		// 重新添加元数据
		result.WithMetadata("command", command)
		result.WithMetadata("exit_code", exitCode)
		result.WithMetadata("duration_ms", duration.Milliseconds())
		result.WithMetadata("success", err == nil)
	}
	
	if cwd != "" {
		result.WithMetadata("cwd", cwd)
	}
	
	if len(env) > 0 {
		result.WithMetadata("env", env)
	}
	
	return result, nil
}

// getShell 获取默认 shell
func (t *BashTool) getShell() string {
	if runtime.GOOS == "windows" {
		// Windows 上使用 cmd 或 powershell
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell"
		}
		return "cmd"
	}
	
	// Unix 系统
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	
	// 默认 shell
	shells := []string{"bash", "sh", "zsh", "fish"}
	for _, shell := range shells {
		if _, err := exec.LookPath(shell); err == nil {
			return shell
		}
	}
	
	return "sh" // 最后的后备选项
}

// checkCommandSafety 检查命令安全性
func (t *BashTool) checkCommandSafety(command string) error {
	// 危险命令列表
	dangerousCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero",
		"mkfs",
		"format",
		":(){ :|:& };:", // Fork bomb
	}
	
	// 危险模式
	dangerousPatterns := []string{
		"> /dev/sda",
		"> /dev/null 2>&1 &",
		"chmod -R 777 /",
		"chown -R",
	}
	
	// 转换为小写进行比较
	lowerCommand := strings.ToLower(command)
	
	// 检查危险命令
	for _, dangerous := range dangerousCommands {
		if strings.Contains(lowerCommand, strings.ToLower(dangerous)) {
			return fmt.Errorf("potentially dangerous command detected: %s", dangerous)
		}
	}
	
	// 检查危险模式
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCommand, strings.ToLower(pattern)) {
			return fmt.Errorf("potentially dangerous pattern detected: %s", pattern)
		}
	}
	
	// 警告：这只是基本的安全检查，不能保证完全安全
	return nil
}

// PipelineTool 管道执行工具
type PipelineTool struct {
	*core.BaseTool
	bashTool *BashTool
}

// NewPipelineTool 创建管道工具
func NewPipelineTool() *PipelineTool {
	tool := &PipelineTool{
		BaseTool: core.NewBaseTool("pipeline", "system", "Execute commands in a pipeline"),
		bashTool: NewBashTool(),
	}
	
	tool.SetRequiresPerm(true)
	tool.SetTags("system", "shell", "pipeline", "chain")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"commands": {
				Type:        "array",
				Description: "List of commands to execute in sequence",
			},
			"stop_on_error": {
				Type:        "boolean",
				Description: "Stop pipeline on first error",
				Default:     true,
			},
			"parallel": {
				Type:        "boolean",
				Description: "Execute commands in parallel",
				Default:     false,
			},
			"cwd": {
				Type:        "string",
				Description: "Working directory for all commands",
				Default:     "",
			},
			"env": {
				Type:        "object",
				Description: "Environment variables for all commands",
				Default:     map[string]string{},
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout for each command in seconds",
				Default:     300,
			},
		},
		Required: []string{"commands"},
	})
	
	return tool
}

// Execute 执行管道
func (t *PipelineTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取命令列表
	commandsRaw, err := params.Get("commands")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid commands parameter")
	}
	
	commands, err := t.parseCommands(commandsRaw)
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("invalid commands: %v", err))
	}
	
	// 获取可选参数
	stopOnError := true
	if params.Has("stop_on_error") {
		stopOnError, _ = params.GetBool("stop_on_error")
	}
	
	parallel := false
	if params.Has("parallel") {
		parallel, _ = params.GetBool("parallel")
	}
	
	// 获取公共参数
	commonParams := core.NewMapParameters(make(map[string]any))
	if params.Has("cwd") {
		if cwd, _ := params.GetString("cwd"); cwd != "" {
			commonParams.Set("cwd", cwd)
		}
	}
	if params.Has("env") {
		if env, err := params.Get("env"); err == nil {
			commonParams.Set("env", env)
		}
	}
	if params.Has("timeout") {
		if timeout, _ := params.GetInt("timeout"); timeout > 0 {
			commonParams.Set("timeout", timeout)
		}
	}
	
	// 执行命令
	var results []map[string]interface{}
	successCount := 0
	failCount := 0
	
	if parallel {
		// 并行执行
		results = t.executeParallel(ctx, commands, commonParams)
	} else {
		// 顺序执行
		results = t.executeSequential(ctx, commands, commonParams, stopOnError)
	}
	
	// 统计结果
	for _, r := range results {
		if success, ok := r["success"].(bool); ok && success {
			successCount++
		} else {
			failCount++
		}
	}
	
	// 创建结果
	result := core.NewSimpleResult(fmt.Sprintf("Executed %d commands: %d succeeded, %d failed", 
		len(commands), successCount, failCount))
	result.WithMetadata("results", results)
	result.WithMetadata("total_commands", len(commands))
	result.WithMetadata("success_count", successCount)
	result.WithMetadata("fail_count", failCount)
	result.WithMetadata("parallel", parallel)
	
	return result, nil
}

// parseCommands 解析命令列表
func (t *PipelineTool) parseCommands(raw interface{}) ([]string, error) {
	var commands []string
	
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if cmd, ok := item.(string); ok {
				commands = append(commands, cmd)
			} else {
				return nil, fmt.Errorf("command must be a string")
			}
		}
	case []string:
		commands = v
	default:
		return nil, fmt.Errorf("commands must be an array of strings")
	}
	
	if len(commands) == 0 {
		return nil, fmt.Errorf("at least one command is required")
	}
	
	return commands, nil
}

// executeSequential 顺序执行命令
func (t *PipelineTool) executeSequential(ctx context.Context, commands []string, commonParams core.Parameters, stopOnError bool) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(commands))
	
	for i, cmd := range commands {
		// 创建命令参数
		cmdParams := core.NewMapParameters(map[string]any{
			"command": cmd,
		})
		
		// 复制公共参数
		if cwd, err := commonParams.GetString("cwd"); err == nil {
			cmdParams.Set("cwd", cwd)
		}
		if env, err := commonParams.Get("env"); err == nil {
			cmdParams.Set("env", env)
		}
		if timeout, err := commonParams.GetInt("timeout"); err == nil {
			cmdParams.Set("timeout", timeout)
		}
		
		// 执行命令
		result, err := t.bashTool.Execute(ctx, cmdParams)
		
		cmdResult := map[string]interface{}{
			"index":   i,
			"command": cmd,
		}
		
		if err != nil {
			cmdResult["success"] = false
			cmdResult["error"] = err.Error()
			results = append(results, cmdResult)
			
			if stopOnError {
				break
			}
		} else {
			cmdResult["success"] = true
			cmdResult["output"] = result.Data()
			cmdResult["metadata"] = result.Metadata()
			results = append(results, cmdResult)
		}
	}
	
	return results
}

// executeParallel 并行执行命令
func (t *PipelineTool) executeParallel(ctx context.Context, commands []string, commonParams core.Parameters) []map[string]interface{} {
	results := make([]map[string]interface{}, len(commands))
	done := make(chan struct{}, len(commands))
	
	for i, cmd := range commands {
		go func(index int, command string) {
			defer func() { done <- struct{}{} }()
			
			// 创建命令参数
			cmdParams := core.NewMapParameters(map[string]any{
				"command": command,
			})
			
			// 复制公共参数
			if cwd, err := commonParams.GetString("cwd"); err == nil {
				cmdParams.Set("cwd", cwd)
			}
			if env, err := commonParams.Get("env"); err == nil {
				cmdParams.Set("env", env)
			}
			if timeout, err := commonParams.GetInt("timeout"); err == nil {
				cmdParams.Set("timeout", timeout)
			}
			
			// 执行命令
			result, err := t.bashTool.Execute(ctx, cmdParams)
			
			cmdResult := map[string]interface{}{
				"index":   index,
				"command": command,
			}
			
			if err != nil {
				cmdResult["success"] = false
				cmdResult["error"] = err.Error()
			} else {
				cmdResult["success"] = true
				cmdResult["output"] = result.Data()
				cmdResult["metadata"] = result.Metadata()
			}
			
			results[index] = cmdResult
		}(i, cmd)
	}
	
	// 等待所有命令完成
	for i := 0; i < len(commands); i++ {
		<-done
	}
	
	return results
}