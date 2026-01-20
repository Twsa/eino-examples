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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// newLambda component initialization function of node 'InputToQuery' in graph 'EinoAgent'
func newLambda(ctx context.Context, input *UserMessage, opts ...any) (output string, err error) {
	return input.Query, nil
}

// newLambda2 component initialization function of node 'InputToHistory' in graph 'EinoAgent'
func newLambda2(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	// Preload Skills
	skillsInfo := "No skills available."
	files, err := os.ReadDir("skills")
	if err == nil {
		var skillList []string
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				name := strings.TrimSuffix(f.Name(), ".md")
				// Parse YAML Frontmatter for description
				content, _ := os.ReadFile(filepath.Join("skills", f.Name()))
				sContent := string(content)
				desc := "No description."
				if strings.HasPrefix(sContent, "---") {
					parts := strings.SplitN(sContent, "---", 3)
					if len(parts) >= 3 {
						yamlLines := strings.Split(parts[1], "\n")
						for _, line := range yamlLines {
							if strings.HasPrefix(line, "description:") {
								desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
								break
							}
						}
					}
				}
				skillList = append(skillList, fmt.Sprintf("- **%s**: %s", name, desc))
			}
		}
		if len(skillList) > 0 {
			skillsInfo = strings.Join(skillList, "\n")
		}
	}

	// Preload Tools
	toolsInfo := "No tools available."
	ts, err := GetTools(ctx)
	if err == nil {
		var toolList []string
		for _, t := range ts {
			// Try to get info using the Eino Tool interface
			if it, ok := t.(interface {
				Info(ctx context.Context) (*schema.ToolInfo, error)
			}); ok {
				info, err := it.Info(ctx)
				if err == nil {
					toolList = append(toolList, fmt.Sprintf("- **%s**: %s", info.Name, info.Desc))
				}
			}
		}
		if len(toolList) > 0 {
			toolsInfo = strings.Join(toolList, "\n")
		}
	}

	return map[string]any{
		"content": input.Query,
		"history": input.History,
		"date":    time.Now().Format("2006-01-02 15:04:05"),
		"skills":  skillsInfo,
		"tools":   toolsInfo,
	}, nil
}
