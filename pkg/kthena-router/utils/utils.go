/*
Copyright The Volcano Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/volcano-sh/kthena/pkg/kthena-router/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

var (
	KVCacheUsage      = "kv_cache_usage"
	RequestWaitingNum = "request_waiting_num"
	RequestRunningNum = "request_running_num"
	TPOT              = "TPOT"
	TTFT              = "TTFT"
)

func GetNamespaceName(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func ParsePrompt(body map[string]interface{}) (*common.ChatMessage, error) {
	if prompt, ok := body["prompt"]; ok {
		promptStr, ok := prompt.(string)
		if !ok {
			return nil, fmt.Errorf("prompt is not a string")
		}
		return &common.ChatMessage{
			Text: promptStr,
		}, nil
	}

	if messages, ok := body["messages"]; ok {
		messageList, ok := messages.([]interface{})
		if !ok {
			return nil, fmt.Errorf("messages is not a list")
		}

		msgs := make([]common.Message, 0, len(messageList))
		for _, message := range messageList {
			msgMap, ok := message.(map[string]interface{})
			if !ok {
				continue
			}

			role, ok := msgMap["role"].(string)
			if !ok {
				continue
			}

			content, err := parseMessageContent(msgMap["content"])
			if err != nil {
				return nil, err
			}

			msgs = append(msgs, common.Message{
				Role:    role,
				Content: content,
			})
		}

		if len(msgs) == 0 {
			return nil, fmt.Errorf("messages contains no usable message")
		}

		return &common.ChatMessage{
			Messages: msgs,
		}, nil
	}

	return nil, fmt.Errorf("prompt or messages not found in request body")
}

// parseMessageContent extracts the text of a single chat message. The OpenAI chat
// completions API allows "content" to be a plain string, an array of content parts,
// or null (for assistant messages that only carry tool_calls), so all three forms
// have to be accepted here. Anything else is a malformed request.
func parseMessageContent(content interface{}) (string, error) {
	switch c := content.(type) {
	case nil:
		// A message with null content still occupies a turn in the conversation,
		// so it is kept with an empty content rather than dropped.
		return "", nil
	case string:
		return c, nil
	case []interface{}:
		var text strings.Builder
		for _, part := range c {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("message content part is not an object")
			}
			// Non-text parts (image_url, input_audio, ...) carry no text and are
			// skipped; only their text siblings contribute to the prompt.
			if partType, ok := partMap["type"].(string); !ok || partType != "text" {
				continue
			}
			partText, ok := partMap["text"].(string)
			if !ok {
				return "", fmt.Errorf("text content part has no string text field")
			}
			text.WriteString(partText)
		}
		return text.String(), nil
	default:
		return "", fmt.Errorf("message content is neither a string nor a list of content parts")
	}
}

func GetPromptString(chatMessage *common.ChatMessage) string {
	// If Text field is present, return text directly (for prompt format)
	if chatMessage.Text != "" {
		return chatMessage.Text
	}

	// For chat messages, convert to ChatML format
	var result strings.Builder
	for _, msg := range chatMessage.Messages {
		fmt.Fprintf(&result, "<|im_start|>%s\n%s<|im_end|>\n", msg.Role, msg.Content)
	}
	return result.String()
}

func LoadEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		klog.Warningf("environment variable %s is not set, using default value: %s", key, defaultValue)
		return defaultValue
	}
	return value
}
