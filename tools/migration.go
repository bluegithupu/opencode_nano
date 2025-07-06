package tools

import (
	"context"
	"opencode_nano/permission"
	"opencode_nano/tools/core"
)

// CreateLegacyToolSet creates tools compatible with the old interface
func CreateLegacyToolSet(perm permission.Manager) ([]Tool, error) {
	// Initialize the new registry
	registry, err := InitializeRegistry()
	if err != nil {
		return nil, err
	}
	
	// Create legacy tools list
	legacyTools := []Tool{}
	
	// Add file read tool (no permission needed)
	if tool, err := registry.Get("read"); err == nil {
		legacyTools = append(legacyTools, NewLegacyAdapter(tool))
	}
	
	// Add file write tool (needs permission)
	if tool, err := registry.Get("write"); err == nil {
		// Wrap with permission check
		wrappedTool := &PermissionWrappedTool{
			tool: tool,
			perm: perm,
		}
		legacyTools = append(legacyTools, NewLegacyAdapter(wrappedTool))
	}
	
	// Add bash tool (needs permission)
	if tool, err := registry.Get("bash"); err == nil {
		// Wrap with permission check
		wrappedTool := &PermissionWrappedTool{
			tool: tool,
			perm: perm,
		}
		legacyTools = append(legacyTools, NewLegacyAdapter(wrappedTool))
	}
	
	// Add task/todo tool (no permission needed)
	if tool, err := registry.Get("todo"); err == nil {
		legacyTools = append(legacyTools, NewLegacyAdapter(tool))
	}
	
	return legacyTools, nil
}

// PermissionWrappedTool wraps a core.Tool with permission checks
type PermissionWrappedTool struct {
	tool core.Tool
	perm permission.Manager
}

// Info returns tool information
func (w *PermissionWrappedTool) Info() core.ToolInfo {
	return w.tool.Info()
}

// Execute executes the tool with permission check
func (w *PermissionWrappedTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	info := w.tool.Info()
	
	// Get command/action description for permission check
	description := info.Description
	if cmdParam, err := params.GetString("command"); err == nil {
		description = "Execute command: " + cmdParam
	} else if pathParam, err := params.GetString("path"); err == nil {
		description = "Write to file: " + pathParam
	}
	
	// Check permission
	if !w.perm.Request(info.Name, description) {
		return nil, core.ErrPermissionDenied(info.Name, "permission denied by user")
	}
	
	// Execute the actual tool
	return w.tool.Execute(ctx, params)
}

// Schema returns the tool's parameter schema
func (w *PermissionWrappedTool) Schema() core.ParameterSchema {
	return w.tool.Schema()
}