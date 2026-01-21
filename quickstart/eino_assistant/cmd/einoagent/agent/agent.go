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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/callbacks/apmplus"
	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/eino/einoagent"
	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/mem"
)

const ChatLogKey = "chat_log"

var memory = mem.GetDefaultMemory()

var cbHandler callbacks.Handler

var once sync.Once

func Init() error {
	var err error
	once.Do(func() {
		os.MkdirAll("log", 0755)
		var f *os.File
		f, err = os.OpenFile("log/eino.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return
		}

		cbConfig := &LogCallbackConfig{
			Detail: true,
			Writer: f,
		}
		if os.Getenv("DEBUG") == "true" {
			cbConfig.Debug = true
		}
		// this is for invoke option of WithCallback
		cbHandler = LogCallback(cbConfig)

		// init global callback, for trace and metrics
		callbackHandlers := make([]callbacks.Handler, 0)
		if os.Getenv("APMPLUS_APP_KEY") != "" {
			region := os.Getenv("APMPLUS_REGION")
			if region == "" {
				region = "cn-beijing"
			}
			fmt.Println("[eino agent] INFO: use apmplus as callback, watch at: https://console.volcengine.com/apmplus-server")
			cbh, _, err := apmplus.NewApmplusHandler(&apmplus.Config{
				Host:        fmt.Sprintf("apmplus-%s.volces.com:4317", region),
				AppKey:      os.Getenv("APMPLUS_APP_KEY"),
				ServiceName: "eino-assistant",
				Release:     "release/v0.0.1",
			})
			if err != nil {
				log.Fatal(err)
			}

			callbackHandlers = append(callbackHandlers, cbh)
		}

		if os.Getenv("LANGFUSE_PUBLIC_KEY") != "" && os.Getenv("LANGFUSE_SECRET_KEY") != "" {
			fmt.Println("[eino agent] INFO: use langfuse as callback, watch at: https://cloud.langfuse.com")
			cbh, _ := langfuse.NewLangfuseHandler(&langfuse.Config{
				Host:      "https://cloud.langfuse.com",
				PublicKey: os.Getenv("LANGFUSE_PUBLIC_KEY"),
				SecretKey: os.Getenv("LANGFUSE_SECRET_KEY"),
				Name:      "Eino Assistant",
				Public:    true,
				Release:   "release/v0.0.1",
				UserID:    "eino_god",
				Tags:      []string{"eino", "assistant"},
			})
			callbackHandlers = append(callbackHandlers, cbh)
		}
		if len(callbackHandlers) > 0 {
			callbacks.InitCallbackHandlers(callbackHandlers)
		}
	})
	return err
}

func RunAgent(ctx context.Context, id string, msg string, chatLog chan string) (*schema.StreamReader[*schema.Message], error) {
	if chatLog != nil {
		ctx = context.WithValue(ctx, ChatLogKey, chatLog)
	}

	runner, err := einoagent.BuildEinoAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build agent graph: %w", err)
	}

	conversation := memory.GetConversation(id, true)

	userMessage := &einoagent.UserMessage{
		ID:      id,
		Query:   msg,
		History: conversation.GetMessages(),
	}
	if os.Getenv("APMPLUS_APP_KEY") != "" {
		// set session info for apmplus callback
		ctx = apmplus.SetSession(ctx, apmplus.WithSessionID(id), apmplus.WithUserID("eino-assistant-user"))
	}
	// Add user query to conversation history BEFORE streaming starts
	conversation.Append(schema.UserMessage(msg))

	reqCallback := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info.Component == "Tool" {
				// 为每个工具调用生成唯一的 ID 并存储在 context 中
				eventID := fmt.Sprintf("evt_%d", time.Now().UnixNano())
				ctx = context.WithValue(ctx, "tool_event_id", eventID)

				var argsStr string
				if inStr, ok := input.(string); ok {
					argsStr = inStr
				}
				// 发送JSON格式的工具开始事件
				startEvent := map[string]interface{}{
					"event": "tool_start",
					"id":    eventID,
					"name":  info.Name,
					"args":  argsStr,
				}
				eventJSON, _ := json.Marshal(startEvent)
				eventStr := string(eventJSON)
				chatLog <- eventStr
				conversation.Append(&schema.Message{
					Role:    "assistant",
					Content: "__TOOL_STEP__" + eventStr,
				})
			} else if info.Component == "Skill" || info.Type == "Skill" {
				skillEvent := map[string]interface{}{
					"event": "skill_start",
					"name":  info.Name,
				}
				eventJSON, _ := json.Marshal(skillEvent)
				chatLog <- string(eventJSON)
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Component == "Tool" {
				var outStr string
				if output != nil {
					if s, ok := output.(string); ok {
						outStr = s
					}
				}
				// 从 context 中获取之前生成的 eventID
				eventID := "unknown"
				if v := ctx.Value("tool_event_id"); v != nil {
					if id, ok := v.(string); ok {
						eventID = id
					}
				}
				// 发送JSON格式的工具结束事件
				endEvent := map[string]interface{}{
					"event":  "tool_end",
					"id":     eventID,
					"name":   info.Name,
					"output": outStr,
				}
				eventJSON, _ := json.Marshal(endEvent)
				eventStr := string(eventJSON)
				chatLog <- eventStr
				conversation.Append(&schema.Message{
					Role:    "assistant",
					Content: "__TOOL_STEP__" + eventStr,
				})
			}
			return ctx
		}).Build()

	// Combine global and request callbacks
	sr, err := runner.Stream(ctx, userMessage, compose.WithCallbacks(cbHandler, reqCallback))
	if err != nil {
		return nil, fmt.Errorf("failed to stream: %w", err)
	}

	srs := sr.Copy(2)

	go func() {
		// for save to memory
		fullMsgs := make([]*schema.Message, 0)

		defer func() {
			// close chat log channel
			close(chatLog)

			// close stream if you used it
			srs[1].Close()

			fullMsg, err := schema.ConcatMessages(fullMsgs)
			if err != nil {
				fmt.Println("error concatenating messages: ", err.Error())
			}
			// add agent response to history
			conversation.Append(fullMsg)
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context done", ctx.Err())
				return
			default:
				chunk, err := srs[1].Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break outer
					}
				}

				fullMsgs = append(fullMsgs, chunk)
			}
		}
	}()

	return srs[0], nil
}

type LogCallbackConfig struct {
	Detail bool
	Debug  bool
	Writer io.Writer
}

func LogCallback(config *LogCallbackConfig) callbacks.Handler {
	if config == nil {
		config = &LogCallbackConfig{
			Detail: true,
			Writer: os.Stdout,
		}
	}
	if config.Writer == nil {
		config.Writer = os.Stdout
	}
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
		fmt.Fprintf(config.Writer, "[view]: start [%s:%s:%s]\n", info.Component, info.Type, info.Name)

		// Check for chat log channel in context
		// val := ctx.Value(ChatLogKey)
		// if val != nil {
		// 	if cl, ok := val.(chan string); ok {
		// 		if info.Type == "Tool" {
		// 			cl <- fmt.Sprintf("▶️ Executing Tool [%s]", info.Name)
		// 			if inStr, ok := input.(string); ok {
		// 				cl <- fmt.Sprintf("   Args: %s", inStr)
		// 			}
		// 		} else if info.Type == "Skill" {
		// 			cl <- fmt.Sprintf("🪄 Activating Skill [%s]", info.Name)
		// 		}
		// 	}
		// }

		if config.Detail {
			var b []byte
			if config.Debug {
				b, _ = json.MarshalIndent(input, "", "  ")
			} else {
				b, _ = json.Marshal(input)
			}
			fmt.Fprintf(config.Writer, "%s\n", string(b))
		}
		return ctx
	})
	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		fmt.Fprintf(config.Writer, "[view]: end [%s:%s:%s]\n", info.Component, info.Type, info.Name)

		// val := ctx.Value(ChatLogKey)
		// if val != nil {
		// 	if cl, ok := val.(chan string); ok {
		// 		if info.Type == "Tool" {
		// 			cl <- fmt.Sprintf("✅ Tool [%s] completed", info.Name)
		// 			if outStr, ok := output.(string); ok {
		// 				if len(outStr) > 500 {
		// 					outStr = outStr[:500] + "..."
		// 				}
		// 				cl <- fmt.Sprintf("   Output: %s", outStr)
		// 			}
		// 		}
		// 	}
		// }

		return ctx
	})
	return builder.Build()
}
