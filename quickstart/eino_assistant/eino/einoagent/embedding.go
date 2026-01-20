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
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	if openAIKey := os.Getenv("OPENAI_API_KEY"); openAIKey != "" {
		baseURL := os.Getenv("OPENAI_EMBED_BASE_URL")
		if baseURL == "" {
			baseURL = os.Getenv("OPENAI_BASE_URL")
		}
		model := os.Getenv("OPENAI_EMBEDDING_MODEL")
		log.Printf("[embedding] Creating OpenAI embedder: base=%s, model=%s", baseURL, model)
		config := &openai.EmbeddingConfig{
			Model:   model,
			APIKey:  openAIKey,
			BaseURL: baseURL,
		}
		return openai.NewEmbedder(ctx, config)
	}

	// Default to Ark
	config := &ark.EmbeddingConfig{
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"),
		APIKey: os.Getenv("ARK_API_KEY"),
	}
	eb, err = ark.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
