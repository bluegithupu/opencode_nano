package core

import (
	"context"
	"fmt"
	"sync"
)

// PipelineStep 管道步骤
type PipelineStep struct {
	Tool   Tool
	Params Parameters
}

// ToolPipeline 工具管道实现
type ToolPipeline struct {
	steps []PipelineStep
}

// NewPipeline 创建新的管道
func NewPipeline() *ToolPipeline {
	return &ToolPipeline{
		steps: make([]PipelineStep, 0),
	}
}

// Add 添加工具到管道
func (p *ToolPipeline) Add(tool Tool, params Parameters) Pipeline {
	p.steps = append(p.steps, PipelineStep{
		Tool:   tool,
		Params: params,
	})
	return p
}

// Execute 执行管道
func (p *ToolPipeline) Execute(ctx context.Context) ([]Result, error) {
	results := make([]Result, 0, len(p.steps))
	
	for i, step := range p.steps {
		select {
		case <-ctx.Done():
			return results, fmt.Errorf("pipeline cancelled at step %d: %v", i, ctx.Err())
		default:
		}
		
		// 如果不是第一步，可以使用前一步的结果作为输入
		if i > 0 && len(results) > 0 {
			prevResult := results[i-1]
			if prevResult.Success() {
				// 将前一步的结果注入到当前参数中
				if mapParams, ok := step.Params.(*MapParameters); ok {
					mapParams.data["_previous_result"] = prevResult.Data()
				}
			}
		}
		
		// 执行当前步骤
		result, err := step.Tool.Execute(ctx, step.Params)
		if err != nil {
			// 创建错误结果
			errResult := NewErrorResult(err)
			errResult.WithMetadata("step", i)
			errResult.WithMetadata("tool", step.Tool.Info().Name)
			results = append(results, errResult)
			
			// 如果某步失败，停止执行后续步骤
			return results, fmt.Errorf("pipeline failed at step %d (%s): %v", 
				i, step.Tool.Info().Name, err)
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// ExecuteAsync 异步执行管道
func (p *ToolPipeline) ExecuteAsync(ctx context.Context) <-chan Result {
	resultChan := make(chan Result, len(p.steps))
	
	go func() {
		defer close(resultChan)
		
		for i, step := range p.steps {
			select {
			case <-ctx.Done():
				// 发送取消错误
				errResult := NewErrorResult(ctx.Err())
				errResult.WithMetadata("step", i)
				errResult.WithMetadata("cancelled", true)
				resultChan <- errResult
				return
			default:
			}
			
			// 执行当前步骤
			result, err := step.Tool.Execute(ctx, step.Params)
			if err != nil {
				// 创建错误结果
				errResult := NewErrorResult(err)
				errResult.WithMetadata("step", i)
				errResult.WithMetadata("tool", step.Tool.Info().Name)
				resultChan <- errResult
				return
			}
			
			resultChan <- result
			
			// 如果工具支持异步执行，可以并发
			if asyncTool, ok := step.Tool.(AsyncTool); ok && i < len(p.steps)-1 {
				// 检查下一步是否依赖当前结果
				// 这里简化处理，实际可以更复杂
				_ = asyncTool
			}
		}
	}()
	
	return resultChan
}

// ParallelPipeline 并行管道实现
type ParallelPipeline struct {
	steps []PipelineStep
}

// NewParallelPipeline 创建并行管道
func NewParallelPipeline() *ParallelPipeline {
	return &ParallelPipeline{
		steps: make([]PipelineStep, 0),
	}
}

// Add 添加工具到并行管道
func (p *ParallelPipeline) Add(tool Tool, params Parameters) *ParallelPipeline {
	p.steps = append(p.steps, PipelineStep{
		Tool:   tool,
		Params: params,
	})
	return p
}

// Execute 并行执行所有工具
func (p *ParallelPipeline) Execute(ctx context.Context) ([]Result, error) {
	results := make([]Result, len(p.steps))
	errors := make([]error, len(p.steps))
	
	var wg sync.WaitGroup
	wg.Add(len(p.steps))
	
	for i, step := range p.steps {
		go func(idx int, s PipelineStep) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				errors[idx] = ctx.Err()
				results[idx] = NewErrorResult(ctx.Err())
				return
			default:
			}
			
			result, err := s.Tool.Execute(ctx, s.Params)
			if err != nil {
				errors[idx] = err
				results[idx] = NewErrorResult(err)
			} else {
				results[idx] = result
			}
		}(i, step)
	}
	
	wg.Wait()
	
	// 检查是否有错误
	var firstError error
	for i, err := range errors {
		if err != nil && firstError == nil {
			firstError = fmt.Errorf("tool %s failed: %v", p.steps[i].Tool.Info().Name, err)
		}
	}
	
	return results, firstError
}

// ConditionalPipeline 条件管道
type ConditionalPipeline struct {
	steps      []ConditionalStep
	defaultStep *PipelineStep
}

// ConditionalStep 条件步骤
type ConditionalStep struct {
	Condition func(prevResults []Result) bool
	Step      PipelineStep
}

// NewConditionalPipeline 创建条件管道
func NewConditionalPipeline() *ConditionalPipeline {
	return &ConditionalPipeline{
		steps: make([]ConditionalStep, 0),
	}
}

// AddIf 添加条件步骤
func (p *ConditionalPipeline) AddIf(condition func([]Result) bool, tool Tool, params Parameters) *ConditionalPipeline {
	p.steps = append(p.steps, ConditionalStep{
		Condition: condition,
		Step: PipelineStep{
			Tool:   tool,
			Params: params,
		},
	})
	return p
}

// SetDefault 设置默认步骤
func (p *ConditionalPipeline) SetDefault(tool Tool, params Parameters) *ConditionalPipeline {
	p.defaultStep = &PipelineStep{
		Tool:   tool,
		Params: params,
	}
	return p
}

// Execute 执行条件管道
func (p *ConditionalPipeline) Execute(ctx context.Context, prevResults []Result) (Result, error) {
	// 检查条件
	for _, step := range p.steps {
		if step.Condition(prevResults) {
			return step.Step.Tool.Execute(ctx, step.Step.Params)
		}
	}
	
	// 执行默认步骤
	if p.defaultStep != nil {
		return p.defaultStep.Tool.Execute(ctx, p.defaultStep.Params)
	}
	
	return NewSimpleResult("no condition matched and no default step"), nil
}