package readurl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type ReadURLTool struct{}

func NewReadURLTool(ctx context.Context) (tool.BaseTool, error) {
	t := &ReadURLTool{}
	return t.ToEinoTool()
}

func (t *ReadURLTool) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("read_url", "fetch content from a web URL and return it as text", t.Invoke)
}

type ReadURLReq struct {
	URL string `json:"url" jsonschema_description:"The web URL to read content from"`
}

type ReadURLRes struct {
	Content string `json:"content" jsonschema_description:"The text content of the web page"`
}

func (t *ReadURLTool) Invoke(ctx context.Context, req ReadURLReq) (ReadURLRes, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(req.URL)
	if err != nil {
		return ReadURLRes{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ReadURLRes{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReadURLRes{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Basic text extraction (could be improved with HTML to text library)
	return ReadURLRes{Content: string(body)}, nil
}
