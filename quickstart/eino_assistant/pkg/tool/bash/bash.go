package bash

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type BashTool struct{}

func NewBashTool(ctx context.Context) (tool.BaseTool, error) {
	t := &BashTool{}
	return t.ToEinoTool()
}

func (t *BashTool) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("bash_executor", "Execute a bash command on the local system and return the output.", t.Invoke)
}

type BashReq struct {
	Command string `json:"command" jsonschema_description:"The bash command to execute"`
}

type BashRes struct {
	Stdout   string `json:"stdout" jsonschema_description:"The standard output of the command"`
	Stderr   string `json:"stderr" jsonschema_description:"The standard error of the command"`
	ExitCode int    `json:"exit_code" jsonschema_description:"The exit code of the command"`
}

func (t *BashTool) Invoke(ctx context.Context, req BashReq) (BashRes, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", req.Command)
	} else {
		cmd = exec.Command("bash", "-c", req.Command)
	}

	stdout, err := cmd.Output()
	res := BashRes{
		Stdout: string(stdout),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.Stderr = string(exitErr.Stderr)
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.Stderr = err.Error()
			res.ExitCode = -1
		}
	} else {
		res.ExitCode = 0
	}

	return res, nil
}
