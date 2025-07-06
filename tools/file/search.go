package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"opencode_nano/tools/core"
)

// SearchTool 文件内容搜索工具
type SearchTool struct {
	*core.BaseTool
}

// NewSearchTool 创建搜索工具
func NewSearchTool() *SearchTool {
	tool := &SearchTool{
		BaseTool: core.NewBaseTool("search", "file", "Search file contents with regex support"),
	}
	
	tool.SetTags("file", "search", "grep", "find")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"pattern": {
				Type:        "string",
				Description: "Search pattern (regex supported)",
			},
			"path": {
				Type:        "string",
				Description: "Directory or file path to search in",
				Default:     ".",
			},
			"file_pattern": {
				Type:        "string",
				Description: "File name pattern to match (e.g., '*.go')",
				Default:     "*",
			},
			"case_sensitive": {
				Type:        "boolean",
				Description: "Case sensitive search",
				Default:     true,
			},
			"recursive": {
				Type:        "boolean",
				Description: "Search recursively in subdirectories",
				Default:     true,
			},
			"max_results": {
				Type:        "integer",
				Description: "Maximum number of results to return",
				Default:     100,
			},
			"context_lines": {
				Type:        "integer",
				Description: "Number of context lines before and after match",
				Default:     0,
			},
		},
		Required: []string{"pattern"},
	})
	
	return tool
}

// Execute 执行搜索
func (t *SearchTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取参数
	pattern, err := params.GetString("pattern")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid pattern parameter")
	}
	
	searchPath := "."
	if params.Has("path") {
		searchPath, _ = params.GetString("path")
	}
	
	filePattern := "*"
	if params.Has("file_pattern") {
		filePattern, _ = params.GetString("file_pattern")
	}
	
	caseSensitive := true
	if params.Has("case_sensitive") {
		caseSensitive, _ = params.GetBool("case_sensitive")
	}
	
	recursive := true
	if params.Has("recursive") {
		recursive, _ = params.GetBool("recursive")
	}
	
	maxResults := 100
	if params.Has("max_results") {
		maxResults, _ = params.GetInt("max_results")
	}
	
	contextLines := 0
	if params.Has("context_lines") {
		contextLines, _ = params.GetInt("context_lines")
	}
	
	// 编译正则表达式
	var re *regexp.Regexp
	if caseSensitive {
		re, err = regexp.Compile(pattern)
	} else {
		re, err = regexp.Compile("(?i)" + pattern)
	}
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("invalid regex pattern: %v", err))
	}
	
	// 搜索文件
	matches := make([]SearchMatch, 0)
	matchCount := 0
	fileCount := 0
	
	err = t.searchFiles(ctx, searchPath, filePattern, recursive, func(path string) error {
		if matchCount >= maxResults {
			return fmt.Errorf("max results reached")
		}
		
		fileMatches, err := t.searchInFile(path, re, contextLines, maxResults-matchCount)
		if err != nil {
			return nil // 忽略单个文件的错误
		}
		
		if len(fileMatches) > 0 {
			fileCount++
			matches = append(matches, fileMatches...)
			matchCount += len(fileMatches)
		}
		
		return nil
	})
	
	// 创建结果
	result := core.NewSimpleResult(fmt.Sprintf("Found %d matches in %d files", matchCount, fileCount))
	result.WithMetadata("matches", matches)
	result.WithMetadata("total_matches", matchCount)
	result.WithMetadata("files_with_matches", fileCount)
	result.WithMetadata("pattern", pattern)
	
	return result, nil
}

// SearchMatch 搜索匹配结果
type SearchMatch struct {
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Column     int      `json:"column"`
	Match      string   `json:"match"`
	Context    []string `json:"context,omitempty"`
	LineText   string   `json:"line_text"`
}

// searchFiles 搜索文件
func (t *SearchTool) searchFiles(ctx context.Context, searchPath, filePattern string, recursive bool, handler func(string) error) error {
	// 检查是否为单个文件
	info, err := os.Stat(searchPath)
	if err != nil {
		return err
	}
	
	if !info.IsDir() {
		// 单个文件
		matched, _ := filepath.Match(filePattern, filepath.Base(searchPath))
		if matched || filePattern == "*" {
			return handler(searchPath)
		}
		return nil
	}
	
	// 目录搜索
	if recursive {
		return filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 忽略错误，继续搜索
			}
			
			// 检查上下文取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			
			if info.IsDir() {
				return nil
			}
			
			matched, _ := filepath.Match(filePattern, filepath.Base(path))
			if matched || filePattern == "*" {
				return handler(path)
			}
			
			return nil
		})
	} else {
		// 非递归搜索
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			return err
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			matched, _ := filepath.Match(filePattern, entry.Name())
			if matched || filePattern == "*" {
				if err := handler(filepath.Join(searchPath, entry.Name())); err != nil {
					return err
				}
			}
		}
		
		return nil
	}
}

// searchInFile 在文件中搜索
func (t *SearchTool) searchInFile(filePath string, re *regexp.Regexp, contextLines, maxMatches int) ([]SearchMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	matches := make([]SearchMatch, 0)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	lines := make([]string, 0)
	
	// 如果需要上下文，先读取所有行
	if contextLines > 0 {
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		scanner = nil
	}
	
	// 搜索匹配
	if contextLines > 0 {
		// 有上下文的搜索
		for i, line := range lines {
			if len(matches) >= maxMatches {
				break
			}
			
			if loc := re.FindStringIndex(line); loc != nil {
				match := SearchMatch{
					File:     filePath,
					Line:     i + 1,
					Column:   loc[0] + 1,
					Match:    line[loc[0]:loc[1]],
					LineText: line,
				}
				
				// 添加上下文
				if contextLines > 0 {
					context := make([]string, 0, contextLines*2+1)
					for j := i - contextLines; j <= i+contextLines; j++ {
						if j >= 0 && j < len(lines) {
							context = append(context, lines[j])
						}
					}
					match.Context = context
				}
				
				matches = append(matches, match)
			}
		}
	} else {
		// 无上下文的搜索
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			
			if len(matches) >= maxMatches {
				break
			}
			
			if loc := re.FindStringIndex(line); loc != nil {
				matches = append(matches, SearchMatch{
					File:     filePath,
					Line:     lineNum,
					Column:   loc[0] + 1,
					Match:    line[loc[0]:loc[1]],
					LineText: line,
				})
			}
		}
	}
	
	return matches, scanner.Err()
}

// GlobTool 文件通配符匹配工具
type GlobTool struct {
	*core.BaseTool
}

// NewGlobTool 创建通配符工具
func NewGlobTool() *GlobTool {
	tool := &GlobTool{
		BaseTool: core.NewBaseTool("glob", "file", "Find files matching glob patterns"),
	}
	
	tool.SetTags("file", "glob", "find", "pattern")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"pattern": {
				Type:        "string",
				Description: "Glob pattern to match (e.g., '**/*.go')",
			},
			"path": {
				Type:        "string",
				Description: "Base directory to search from",
				Default:     ".",
			},
			"exclude": {
				Type:        "array",
				Description: "Patterns to exclude",
				Default:     []string{},
			},
			"include_dirs": {
				Type:        "boolean",
				Description: "Include directories in results",
				Default:     false,
			},
			"follow_symlinks": {
				Type:        "boolean",
				Description: "Follow symbolic links",
				Default:     false,
			},
			"max_results": {
				Type:        "integer",
				Description: "Maximum number of results",
				Default:     1000,
			},
		},
		Required: []string{"pattern"},
	})
	
	return tool
}

// Execute 执行通配符匹配
func (t *GlobTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取参数
	pattern, err := params.GetString("pattern")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid pattern parameter")
	}
	
	basePath := "."
	if params.Has("path") {
		basePath, _ = params.GetString("path")
	}
	
	excludePatterns := []string{}
	if params.Has("exclude") {
		if excludeRaw, err := params.Get("exclude"); err == nil {
			if excludeList, ok := excludeRaw.([]interface{}); ok {
				for _, e := range excludeList {
					if s, ok := e.(string); ok {
						excludePatterns = append(excludePatterns, s)
					}
				}
			}
		}
	}
	
	includeDirs := false
	if params.Has("include_dirs") {
		includeDirs, _ = params.GetBool("include_dirs")
	}
	
	maxResults := 1000
	if params.Has("max_results") {
		maxResults, _ = params.GetInt("max_results")
	}
	
	// 执行通配符匹配
	matches := []string{}
	
	// 处理 ** 模式
	if strings.Contains(pattern, "**") {
		err = t.globRecursive(ctx, basePath, pattern, excludePatterns, includeDirs, maxResults, &matches)
	} else {
		// 简单匹配
		globPattern := filepath.Join(basePath, pattern)
		files, err := filepath.Glob(globPattern)
		if err == nil {
			for _, file := range files {
				if len(matches) >= maxResults {
					break
				}
				
				// 检查排除模式
				excluded := false
				for _, exclude := range excludePatterns {
					if matched, _ := filepath.Match(exclude, filepath.Base(file)); matched {
						excluded = true
						break
					}
				}
				
				if !excluded {
					info, err := os.Stat(file)
					if err == nil && (includeDirs || !info.IsDir()) {
						matches = append(matches, file)
					}
				}
			}
		}
	}
	
	// 创建结果
	result := core.NewSimpleResult(fmt.Sprintf("Found %d files matching pattern", len(matches)))
	result.WithMetadata("files", matches)
	result.WithMetadata("count", len(matches))
	result.WithMetadata("pattern", pattern)
	
	return result, nil
}

// globRecursive 递归通配符匹配
func (t *GlobTool) globRecursive(ctx context.Context, basePath, pattern string, excludes []string, includeDirs bool, maxResults int, matches *[]string) error {
	// 分解 ** 模式
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return fmt.Errorf("invalid ** pattern")
	}
	
	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")
	
	return filepath.Walk(filepath.Join(basePath, prefix), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误
		}
		
		// 检查上下文取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if len(*matches) >= maxResults {
			return fmt.Errorf("max results reached")
		}
		
		// 检查是否匹配后缀
		relPath, _ := filepath.Rel(filepath.Join(basePath, prefix), path)
		if matched, _ := filepath.Match(suffix, relPath); matched {
			// 检查排除模式
			excluded := false
			for _, exclude := range excludes {
				if matched, _ := filepath.Match(exclude, filepath.Base(path)); matched {
					excluded = true
					break
				}
			}
			
			if !excluded && (includeDirs || !info.IsDir()) {
				*matches = append(*matches, path)
			}
		}
		
		return nil
	})
}