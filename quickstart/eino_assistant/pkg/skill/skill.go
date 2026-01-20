package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type SkillManager struct {
	BaseDir string
}

func NewSkillManager(ctx context.Context, baseDir string) (tool.BaseTool, error) {
	if baseDir == "" {
		baseDir = "skills"
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	t := &SkillManager{BaseDir: baseDir}
	return t.ToEinoTool()
}

func (t *SkillManager) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("skill_manager", "Manager for Agent Skills. Use this to save, list, or retrieve specific complex procedures (Bash scripts, SOPs, etc).", t.Invoke)
}

type Action string

const (
	ActionSave Action = "save"
	ActionList Action = "list"
	ActionGet  Action = "get"
)

type SkillReq struct {
	Action      Action `json:"action" jsonschema_description:"action to perform: save, list, get"`
	Name        string `json:"name,omitempty" jsonschema_description:"name of the skill (e.g. 'check_system_health')"`
	Description string `json:"description,omitempty" jsonschema_description:"description of what the skill does"`
	Content     string `json:"content,omitempty" jsonschema_description:"the content/procedure of the skill (e.g. bash script or steps)"`
}

type SkillRes struct {
	Status string   `json:"status"`
	Skills []string `json:"skills,omitempty"`
	Skill  *Skill   `json:"skill,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

func (t *SkillManager) Invoke(ctx context.Context, req SkillReq) (SkillRes, error) {
	switch req.Action {
	case ActionSave:
		if req.Name == "" || req.Content == "" {
			return SkillRes{Status: "error", Error: "name and content are required for save"}, nil
		}
		filename := filepath.Join(t.BaseDir, req.Name+".md")

		// Standardized Skill Format with YAML Frontmatter
		data := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n%s",
			req.Name, req.Description, req.Content)

		if err := os.WriteFile(filename, []byte(data), 0644); err != nil {
			return SkillRes{Status: "error", Error: err.Error()}, nil
		}
		return SkillRes{Status: "success"}, nil

	case ActionList:
		files, err := os.ReadDir(t.BaseDir)
		if err != nil {
			return SkillRes{Status: "error", Error: err.Error()}, nil
		}
		var skillNames []string
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				skillNames = append(skillNames, strings.TrimSuffix(f.Name(), ".md"))
			}
		}
		return SkillRes{Status: "success", Skills: skillNames}, nil

	case ActionGet:
		if req.Name == "" {
			return SkillRes{Status: "error", Error: "name is required for get"}, nil
		}
		filename := filepath.Join(t.BaseDir, req.Name+".md")
		data, err := os.ReadFile(filename)
		if err != nil {
			return SkillRes{Status: "error", Error: "skill not found"}, nil
		}

		content := string(data)
		name := req.Name
		desc := ""

		// Simple YAML Frontmatter parsing
		if strings.HasPrefix(content, "---") {
			parts := strings.SplitN(content, "---", 3)
			if len(parts) >= 3 {
				yamlLines := strings.Split(parts[1], "\n")
				for _, line := range yamlLines {
					if strings.HasPrefix(line, "name:") {
						name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
					} else if strings.HasPrefix(line, "description:") {
						desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
					}
				}
				content = strings.TrimSpace(parts[2])
			}
		}

		return SkillRes{Status: "success", Skill: &Skill{
			Name:        name,
			Description: desc,
			Content:     content,
		}}, nil

	default:
		return SkillRes{Status: "error", Error: "unknown action"}, nil
	}
}
