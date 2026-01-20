/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package einoagent

import (
	"context"

	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/skill"
	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/tool/bash"
	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/tool/gitclone"
	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/tool/open"
	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/tool/task"
	"github.com/cloudwego/eino/components/tool"
)

func GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	toolTask, err := NewTaskTool(ctx)
	if err != nil {
		return nil, err
	}

	toolOpen, err := NewOpenFileTool(ctx)
	if err != nil {
		return nil, err
	}

	toolGitClone, err := NewGitCloneFile(ctx)
	if err != nil {
		return nil, err
	}

	toolBash, err := NewBashTool(ctx)
	if err != nil {
		return nil, err
	}

	toolSkill, err := NewSkillTool(ctx)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{
		toolTask,
		toolOpen,
		toolGitClone,
		toolBash,
		toolSkill,
	}, nil
}

func NewOpenFileTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return open.NewOpenFileTool(ctx, nil)
}

func NewGitCloneFile(ctx context.Context) (tn tool.BaseTool, err error) {
	return gitclone.NewGitCloneFile(ctx, nil)
}

func NewTaskTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return task.NewTaskTool(ctx, nil)
}

func NewBashTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return bash.NewBashTool(ctx)
}

func NewSkillTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return skill.NewSkillManager(ctx, "skills")
}
