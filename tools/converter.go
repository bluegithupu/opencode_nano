package tools

import (
	"context"
	"opencode_nano/permission"
	"opencode_nano/tools/core"
	"opencode_nano/tools/file"
	"opencode_nano/tools/system"
	"opencode_nano/tools/task"
)

// CreateToolSet creates tools compatible with the old interface directly
func CreateToolSet(perm permission.Manager) ([]Tool, error) {
	// Create tools list
	var tools []Tool
	
	// Add file read tool (no permission needed)
	readTool := file.NewReadTool()
	tools = append(tools, &CoreToolAdapter{tool: readTool})
	
	// Add file write tool (needs permission) 
	writeTool := file.NewWriteTool()
	tools = append(tools, &CoreToolAdapter{
		tool: writeTool,
		needsPerm: true,
		perm: perm,
	})
	
	// Add bash tool (needs permission)
	bashTool := system.NewBashTool()
	tools = append(tools, &CoreToolAdapter{
		tool: bashTool,
		needsPerm: true,
		perm: perm,
	})
	
	// Add task/todo tool (no permission needed)
	taskTool, err := task.NewTaskTool()
	if err != nil {
		return nil, err
	}
	tools = append(tools, &CoreToolAdapter{tool: taskTool})
	
	return tools, nil
}

// CoreToolAdapter adapts core.Tool to the old Tool interface
type CoreToolAdapter struct {
	tool      core.Tool
	needsPerm bool
	perm      permission.Manager
}

func (a *CoreToolAdapter) Name() string {
	return a.tool.Info().Name
}

func (a *CoreToolAdapter) Description() string {
	return a.tool.Info().Description
}

func (a *CoreToolAdapter) Parameters() map[string]interface{} {
	schema := a.tool.Schema()
	params := make(map[string]interface{})
	
	params["type"] = "object"
	params["required"] = schema.Required
	
	properties := make(map[string]interface{})
	for key, prop := range schema.Properties {
		propMap := map[string]interface{}{
			"type":        prop.Type,
			"description": prop.Description,
		}
		if len(prop.Enum) > 0 {
			propMap["enum"] = prop.Enum
		}
		if prop.Default != nil {
			propMap["default"] = prop.Default
		}
		properties[key] = propMap
	}
	params["properties"] = properties
	
	return params
}

func (a *CoreToolAdapter) Execute(params map[string]interface{}) (string, error) {
	// Check permission if needed
	if a.needsPerm {
		description := a.tool.Info().Description
		if cmd, ok := params["command"].(string); ok {
			description = "Execute command: " + cmd
		} else if path, ok := params["path"].(string); ok {
			description = "Write to file: " + path
		} else if filePath, ok := params["file_path"].(string); ok {
			description = "Write to file: " + filePath
		}
		
		if !a.perm.Request(a.tool.Info().Name, description) {
			return "", core.ErrPermissionDenied(a.tool.Info().Name, "permission denied by user")
		}
	}
	
	// Execute the tool
	coreParams := core.NewMapParameters(params)
	result, err := a.tool.Execute(context.Background(), coreParams)
	if err != nil {
		return "", err
	}
	
	return result.String(), nil
}