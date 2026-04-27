package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

var codexModelMap = map[string]string{
	"gpt-5.5":                    "gpt-5.5",
	"gpt-5":                      "gpt-5",
	"gpt-5-codex":                "gpt-5-codex",
	"gpt-5-codex-mini":           "gpt-5-codex-mini",
	"gpt-5.1":                    "gpt-5.1",
	"gpt-5.1-none":               "gpt-5.1",
	"gpt-5.1-low":                "gpt-5.1",
	"gpt-5.1-medium":             "gpt-5.1",
	"gpt-5.1-high":               "gpt-5.1",
	"gpt-5.1-chat-latest":        "gpt-5.1",
	"gpt-5.1-codex":              "gpt-5.1-codex",
	"gpt-5.1-codex-low":          "gpt-5.1-codex",
	"gpt-5.1-codex-medium":       "gpt-5.1-codex",
	"gpt-5.1-codex-high":         "gpt-5.1-codex",
	"gpt-5.1-codex-max":          "gpt-5.1-codex-max",
	"gpt-5.1-codex-max-low":      "gpt-5.1-codex-max",
	"gpt-5.1-codex-max-medium":   "gpt-5.1-codex-max",
	"gpt-5.1-codex-max-high":     "gpt-5.1-codex-max",
	"gpt-5.1-codex-max-xhigh":    "gpt-5.1-codex-max",
	"gpt-5.1-codex-mini":         "gpt-5.1-codex-mini",
	"gpt-5.1-codex-mini-medium":  "gpt-5.1-codex-mini",
	"gpt-5.1-codex-mini-high":    "gpt-5.1-codex-mini",
	"codex-mini-latest":          "gpt-5.1-codex-mini",
	"gpt-5.2":                    "gpt-5.2",
	"gpt-5.2-none":               "gpt-5.2",
	"gpt-5.2-low":                "gpt-5.2",
	"gpt-5.2-medium":             "gpt-5.2",
	"gpt-5.2-high":               "gpt-5.2",
	"gpt-5.2-xhigh":              "gpt-5.2",
	"gpt-5.2-codex":              "gpt-5.2-codex",
	"gpt-5.2-codex-low":          "gpt-5.2-codex",
	"gpt-5.2-codex-medium":       "gpt-5.2-codex",
	"gpt-5.2-codex-high":         "gpt-5.2-codex",
	"gpt-5.2-codex-xhigh":        "gpt-5.2-codex",
	"gpt-5.3-codex":              "gpt-5.3-codex",
	"gpt-5.3-codex-low":          "gpt-5.3-codex",
	"gpt-5.3-codex-medium":       "gpt-5.3-codex",
	"gpt-5.3-codex-high":         "gpt-5.3-codex",
	"gpt-5.3-codex-xhigh":        "gpt-5.3-codex",
	"gpt-5.3-codex-spark":        "gpt-5.3-codex",
	"gpt-5.3-codex-spark-low":    "gpt-5.3-codex",
	"gpt-5.3-codex-spark-medium": "gpt-5.3-codex",
	"gpt-5.3-codex-spark-high":   "gpt-5.3-codex",
	"gpt-5.3-codex-spark-xhigh":  "gpt-5.3-codex",
	"gpt-5.4":                    "gpt-5.4",
	"gpt-5.4-none":               "gpt-5.4",
	"gpt-5.4-low":                "gpt-5.4",
	"gpt-5.4-medium":             "gpt-5.4",
	"gpt-5.4-high":               "gpt-5.4",
	"gpt-5.4-xhigh":              "gpt-5.4",
	"gpt-5.4-chat-latest":        "gpt-5.4",
	"gpt-5.4-mini":               "gpt-5.4-mini",
	"gpt-5.4-nano":               "gpt-5.4-nano",
}

type codexTransformResult struct {
	Modified                        bool
	NormalizedModel                 string
	PromptCacheKey                  string
	DroppedNativeItemReferenceCount int
}

var chatGPTOAuthLegacyModelMap = map[string]string{
	"gpt-5.1-codex":      "gpt-5.3-codex",
	"gpt-5.2-codex":      "gpt-5.3-codex",
	"gpt-5.1-codex-mini": "gpt-5.4-mini",
}

func applyCodexOAuthTransform(reqBody map[string]any, isCodexCLI bool, isCompact bool) codexTransformResult {
	return applyCodexOAuthTransformWithOptions(reqBody, isCodexCLI, isCompact, codexTransformOptions{})
}

type codexTransformOptions struct {
	DropStoreFalseNativeItemReferences bool
}

func applyCodexOAuthTransformWithOptions(
	reqBody map[string]any,
	isCodexCLI bool,
	isCompact bool,
	opts codexTransformOptions,
) codexTransformResult {
	result := codexTransformResult{}
	// 工具续链需求会影响存储策略与 input 过滤逻辑。
	needsToolContinuation := NeedsToolContinuation(reqBody)

	model := ""
	if v, ok := reqBody["model"].(string); ok {
		model = v
	}
	normalizedModel, _ := normalizeChatGPTOAuthModel(model)
	if normalizedModel != "" {
		if model != normalizedModel {
			reqBody["model"] = normalizedModel
			result.Modified = true
		}
		result.NormalizedModel = normalizedModel
	}

	if isCompact {
		if _, ok := reqBody["store"]; ok {
			delete(reqBody, "store")
			result.Modified = true
		}
		if _, ok := reqBody["stream"]; ok {
			delete(reqBody, "stream")
			result.Modified = true
		}
	} else {
		// OAuth 走 ChatGPT internal API 时，store 必须为 false；显式 true 也会强制覆盖。
		// 避免上游返回 "Store must be set to false"。
		if v, ok := reqBody["store"].(bool); !ok || v {
			reqBody["store"] = false
			result.Modified = true
		}
		if v, ok := reqBody["stream"].(bool); !ok || !v {
			reqBody["stream"] = true
			result.Modified = true
		}
	}

	// Strip parameters unsupported by codex models via the Responses API.
	for _, key := range []string{
		"max_output_tokens",
		"max_completion_tokens",
		"temperature",
		"top_p",
		"frequency_penalty",
		"presence_penalty",
		"prompt_cache_retention",
	} {
		if _, ok := reqBody[key]; ok {
			delete(reqBody, key)
			result.Modified = true
		}
	}

	// 兼容遗留的 functions 和 function_call，转换为 tools 和 tool_choice
	if functionsRaw, ok := reqBody["functions"]; ok {
		if functions, k := functionsRaw.([]any); k {
			tools := make([]any, 0, len(functions))
			for _, f := range functions {
				tools = append(tools, map[string]any{
					"type":     "function",
					"function": f,
				})
			}
			reqBody["tools"] = tools
		}
		delete(reqBody, "functions")
		result.Modified = true
	}

	if fcRaw, ok := reqBody["function_call"]; ok {
		if fcStr, ok := fcRaw.(string); ok {
			// e.g. "auto", "none"
			reqBody["tool_choice"] = fcStr
		} else if fcObj, ok := fcRaw.(map[string]any); ok {
			// e.g. {"name": "my_func"}
			if name, ok := fcObj["name"].(string); ok && strings.TrimSpace(name) != "" {
				reqBody["tool_choice"] = map[string]any{
					"type": "function",
					"function": map[string]any{
						"name": name,
					},
				}
			}
		}
		delete(reqBody, "function_call")
		result.Modified = true
	}

	if normalizeCodexTools(reqBody) {
		result.Modified = true
	}
	if normalizeCodexToolChoice(reqBody) {
		result.Modified = true
	}

	if v, ok := reqBody["prompt_cache_key"].(string); ok {
		result.PromptCacheKey = strings.TrimSpace(v)
	}

	// 提取 input 中 role:"system" 消息至 instructions（OAuth 上游不支持 system role）。
	if extractSystemMessagesFromInput(reqBody) {
		result.Modified = true
	}

	// instructions 处理逻辑：根据是否是 Codex CLI 分别调用不同方法
	if applyInstructions(reqBody, isCodexCLI) {
		result.Modified = true
	}

	storeDisabled := false
	if store, ok := reqBody["store"].(bool); ok && !store {
		storeDisabled = true
	}

	// 续链场景仅保留有效的工具续链引用与必要 id；
	// 是否过滤 store=false 下的 native item_reference 由账号级开关控制。
	if input, ok := reqBody["input"].([]any); ok {
		if normalizedInput, modified := normalizeCodexToolRoleMessages(input); modified {
			input = normalizedInput
			result.Modified = true
		}
		if normalizedInput, modified := normalizeCodexMessageContentText(input); modified {
			input = normalizedInput
			result.Modified = true
		}
		var droppedCount int
		input, droppedCount = filterCodexInput(
			input,
			needsToolContinuation,
			storeDisabled,
			opts.DropStoreFalseNativeItemReferences,
		)
		reqBody["input"] = input
		result.DroppedNativeItemReferenceCount += droppedCount
		result.Modified = true
	} else if inputStr, ok := reqBody["input"].(string); ok {
		// ChatGPT codex endpoint requires input to be a list, not a string.
		// Convert string input to the expected message array format.
		trimmed := strings.TrimSpace(inputStr)
		if trimmed != "" {
			reqBody["input"] = []any{
				map[string]any{
					"type":    "message",
					"role":    "user",
					"content": inputStr,
				},
			}
		} else {
			reqBody["input"] = []any{}
		}
		result.Modified = true
	}

	return result
}

func normalizeCodexToolChoice(reqBody map[string]any) bool {
	choice, ok := reqBody["tool_choice"]
	if !ok || choice == nil {
		return false
	}
	choiceMap, ok := choice.(map[string]any)
	if !ok {
		return false
	}
	choiceType := strings.TrimSpace(firstNonEmptyString(choiceMap["type"]))
	if choiceType == "" || codexToolsContainType(reqBody["tools"], choiceType) {
		return false
	}
	reqBody["tool_choice"] = "auto"
	return true
}

func codexToolsContainType(rawTools any, toolType string) bool {
	tools, ok := rawTools.([]any)
	if !ok || strings.TrimSpace(toolType) == "" {
		return false
	}
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(firstNonEmptyString(tool["type"])) == toolType {
			return true
		}
	}
	return false
}

func normalizeCodexToolRoleMessages(input []any) ([]any, bool) {
	if len(input) == 0 {
		return input, false
	}

	modified := false
	normalized := make([]any, 0, len(input))
	for _, item := range input {
		m, ok := item.(map[string]any)
		if !ok {
			normalized = append(normalized, item)
			continue
		}
		role, _ := m["role"].(string)
		if strings.TrimSpace(role) != "tool" {
			normalized = append(normalized, item)
			continue
		}

		callID := strings.TrimSpace(firstNonEmptyString(m["call_id"], m["tool_call_id"], m["id"]))
		if callID == "" {
			fallback := make(map[string]any, len(m))
			for key, value := range m {
				fallback[key] = value
			}
			fallback["role"] = "user"
			delete(fallback, "tool_call_id")
			normalized = append(normalized, fallback)
			modified = true
			continue
		}

		output := extractTextFromContent(m["content"])
		if output == "" {
			if value, ok := m["output"].(string); ok {
				output = value
			}
		}
		if output == "" && m["content"] != nil {
			if b, err := json.Marshal(m["content"]); err == nil {
				output = string(b)
			}
		}

		normalized = append(normalized, map[string]any{
			"type":    "function_call_output",
			"call_id": callID,
			"output":  output,
		})
		modified = true
	}
	if !modified {
		return input, false
	}
	return normalized, true
}

func normalizeCodexMessageContentText(input []any) ([]any, bool) {
	if len(input) == 0 {
		return input, false
	}

	modified := false
	normalized := make([]any, 0, len(input))
	for _, item := range input {
		m, ok := item.(map[string]any)
		if !ok || strings.TrimSpace(firstNonEmptyString(m["type"])) != "message" {
			normalized = append(normalized, item)
			continue
		}
		parts, ok := m["content"].([]any)
		if !ok {
			normalized = append(normalized, item)
			continue
		}

		var newItem map[string]any
		var newParts []any
		ensureItemCopy := func() {
			if newItem != nil {
				return
			}
			newItem = make(map[string]any, len(m))
			for key, value := range m {
				newItem[key] = value
			}
			newParts = make([]any, len(parts))
			copy(newParts, parts)
		}

		for i, rawPart := range parts {
			part, ok := rawPart.(map[string]any)
			if !ok {
				continue
			}
			text, hasText := part["text"]
			if !hasText {
				continue
			}
			if _, ok := text.(string); ok {
				continue
			}

			ensureItemCopy()
			newPart := make(map[string]any, len(part))
			for key, value := range part {
				newPart[key] = value
			}
			newPart["text"] = stringifyCodexContentText(text)
			newParts[i] = newPart
			modified = true
		}

		if newItem != nil {
			newItem["content"] = newParts
			normalized = append(normalized, newItem)
			continue
		}
		normalized = append(normalized, item)
	}
	if !modified {
		return input, false
	}
	return normalized, true
}

func stringifyCodexContentText(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprint(v)
	}
}

func normalizeCodexModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return openai.DefaultTestModel
	}
	if isOpenAIImageGenerationModel(model) {
		return model
	}

	modelID := model
	if strings.Contains(modelID, "/") {
		parts := strings.Split(modelID, "/")
		modelID = parts[len(parts)-1]
	}

	if mapped := getNormalizedCodexModel(modelID); mapped != "" {
		return mapped
	}

	normalized := strings.ToLower(modelID)

	if strings.Contains(normalized, "gpt-5.5") || strings.Contains(normalized, "gpt 5.5") {
		return "gpt-5.5"
	}
	if strings.Contains(normalized, "gpt-5.4-mini") || strings.Contains(normalized, "gpt 5.4 mini") {
		return "gpt-5.4-mini"
	}
	if strings.Contains(normalized, "gpt-5.4-nano") || strings.Contains(normalized, "gpt 5.4 nano") {
		return "gpt-5.4-nano"
	}
	if strings.Contains(normalized, "gpt-5.4") || strings.Contains(normalized, "gpt 5.4") {
		return "gpt-5.4"
	}
	if strings.Contains(normalized, "gpt-5.3-codex-spark") || strings.Contains(normalized, "gpt 5.3 codex spark") {
		return "gpt-5.3-codex"
	}
	if strings.Contains(normalized, "gpt-5.3-codex") || strings.Contains(normalized, "gpt 5.3 codex") {
		return "gpt-5.3-codex"
	}
	if strings.Contains(normalized, "gpt-5.2-codex") || strings.Contains(normalized, "gpt 5.2 codex") {
		return "gpt-5.2-codex"
	}
	if strings.Contains(normalized, "gpt-5.2") || strings.Contains(normalized, "gpt 5.2") {
		return "gpt-5.2"
	}
	if strings.Contains(normalized, "gpt-5.1-codex-max") || strings.Contains(normalized, "gpt 5.1 codex max") {
		return "gpt-5.1-codex-max"
	}
	if strings.Contains(normalized, "gpt-5.1-codex-mini") || strings.Contains(normalized, "gpt 5.1 codex mini") {
		return "gpt-5.1-codex-mini"
	}
	if strings.Contains(normalized, "gpt-5-codex-mini") || strings.Contains(normalized, "gpt 5 codex mini") {
		return "gpt-5-codex-mini"
	}
	if strings.Contains(normalized, "gpt-5.1-codex") || strings.Contains(normalized, "gpt 5.1 codex") {
		return "gpt-5.1-codex"
	}
	if strings.Contains(normalized, "gpt-5-codex") || strings.Contains(normalized, "gpt 5 codex") {
		return "gpt-5-codex"
	}
	if strings.Contains(normalized, "gpt-5.1") || strings.Contains(normalized, "gpt 5.1") {
		return "gpt-5.1"
	}
	if normalized == "gpt-5" || normalized == "gpt 5" {
		return "gpt-5"
	}

	return normalized
}

func hasOpenAIImageGenerationTool(reqBody map[string]any) bool {
	rawTools, ok := reqBody["tools"]
	if !ok || rawTools == nil {
		return false
	}
	tools, ok := rawTools.([]any)
	if !ok {
		return false
	}
	for _, rawTool := range tools {
		toolMap, ok := rawTool.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(firstNonEmptyString(toolMap["type"])) == "image_generation" {
			return true
		}
	}
	return false
}

func normalizeOpenAIResponsesImageGenerationTools(reqBody map[string]any) bool {
	rawTools, ok := reqBody["tools"]
	if !ok || rawTools == nil {
		return false
	}
	tools, ok := rawTools.([]any)
	if !ok {
		return false
	}

	modified := false
	for _, rawTool := range tools {
		toolMap, ok := rawTool.(map[string]any)
		if !ok || strings.TrimSpace(firstNonEmptyString(toolMap["type"])) != "image_generation" {
			continue
		}
		if _, ok := toolMap["output_format"]; !ok {
			if value := strings.TrimSpace(firstNonEmptyString(toolMap["format"])); value != "" {
				toolMap["output_format"] = value
				modified = true
			}
		}
		if _, ok := toolMap["output_compression"]; !ok {
			if value, exists := toolMap["compression"]; exists && value != nil {
				toolMap["output_compression"] = value
				modified = true
			}
		}
		if _, ok := toolMap["format"]; ok {
			delete(toolMap, "format")
			modified = true
		}
		if _, ok := toolMap["compression"]; ok {
			delete(toolMap, "compression")
			modified = true
		}
	}
	return modified
}

func validateOpenAIResponsesImageModel(reqBody map[string]any, model string) error {
	if !hasOpenAIImageGenerationTool(reqBody) {
		return nil
	}
	model = strings.TrimSpace(model)
	if !isOpenAIImageGenerationModel(model) {
		return nil
	}
	return fmt.Errorf("/v1/responses image_generation requests require a Responses-capable text model; image-only model %q is not allowed", model)
}

func normalizeOpenAIModelForUpstream(account *Account, model string) string {
	if account == nil || account.Type == AccountTypeOAuth {
		return normalizeCodexModel(model)
	}
	if account.IsOpenAIApiKey() {
		return NormalizeOpenAICompatRequestedModel(model)
	}
	return strings.TrimSpace(model)
}

func normalizeChatGPTOAuthModel(model string) (normalized string, remapped bool) {
	normalized = normalizeCodexModel(model)
	if normalized == "" {
		return "", false
	}
	if mapped, ok := chatGPTOAuthLegacyModelMap[normalized]; ok {
		return mapped, true
	}
	return normalized, false
}

func SupportsVerbosity(model string) bool {
	if !strings.HasPrefix(model, "gpt-") {
		return true
	}

	var major, minor int
	n, _ := fmt.Sscanf(model, "gpt-%d.%d", &major, &minor)

	if major > 5 {
		return true
	}
	if major < 5 {
		return false
	}

	// gpt-5
	if n == 1 {
		return true
	}

	return minor >= 3
}

func getNormalizedCodexModel(modelID string) string {
	if modelID == "" {
		return ""
	}
	if mapped, ok := codexModelMap[modelID]; ok {
		return mapped
	}
	lower := strings.ToLower(modelID)
	for key, value := range codexModelMap {
		if strings.ToLower(key) == lower {
			return value
		}
	}
	return ""
}

// extractTextFromContent extracts plain text from a content value that is either
// a Go string or a []any of content-part maps with type:"text"/"input_text".
func extractTextFromContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, part := range v {
			m, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if t, _ := m["type"].(string); t == "text" || t == "input_text" {
				if text, ok := m["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "")
	default:
		return ""
	}
}

// extractSystemMessagesFromInput scans the input array for items with role=="system",
// removes them, and merges their content into reqBody["instructions"].
// If instructions is already non-empty, extracted content is prepended with "\n\n".
// Returns true if any system messages were extracted.
func extractSystemMessagesFromInput(reqBody map[string]any) bool {
	input, ok := reqBody["input"].([]any)
	if !ok || len(input) == 0 {
		return false
	}

	var systemTexts []string
	remaining := make([]any, 0, len(input))

	for _, item := range input {
		m, ok := item.(map[string]any)
		if !ok {
			remaining = append(remaining, item)
			continue
		}
		if role, _ := m["role"].(string); role != "system" {
			remaining = append(remaining, item)
			continue
		}
		if text := extractTextFromContent(m["content"]); text != "" {
			systemTexts = append(systemTexts, text)
		}
	}

	if len(systemTexts) == 0 {
		return false
	}

	extracted := strings.Join(systemTexts, "\n\n")
	if existing, ok := reqBody["instructions"].(string); ok && strings.TrimSpace(existing) != "" {
		reqBody["instructions"] = extracted + "\n\n" + existing
	} else {
		reqBody["instructions"] = extracted
	}
	reqBody["input"] = remaining
	return true
}

// applyInstructions 处理 instructions 字段：仅在 instructions 为空时填充默认值。
func applyInstructions(reqBody map[string]any, isCodexCLI bool) bool {
	if !isInstructionsEmpty(reqBody) {
		return false
	}
	reqBody["instructions"] = openAIResponsesDefaultInstructions
	return true
}

// isInstructionsEmpty 检查 instructions 字段是否为空
// 处理以下情况：字段不存在、nil、空字符串、纯空白字符串
func isInstructionsEmpty(reqBody map[string]any) bool {
	val, exists := reqBody["instructions"]
	if !exists {
		return true
	}
	if val == nil {
		return true
	}
	str, ok := val.(string)
	if !ok {
		return true
	}
	return strings.TrimSpace(str) == ""
}

// filterCodexInput 按需过滤 item_reference 与 id。
// preserveReferences 为 true 时保持必要的工具续链引用与 id；
// storeDisabled + dropStoreFalseNativeItemReferences 为 true 时丢弃依赖服务端持久化的 native item_reference（如 rs_*）。
func filterCodexInput(
	input []any,
	preserveReferences bool,
	storeDisabled bool,
	dropStoreFalseNativeItemReferences bool,
) ([]any, int) {
	filtered := make([]any, 0, len(input))
	droppedNativeItemReferenceCount := 0
	for _, item := range input {
		m, ok := item.(map[string]any)
		if !ok {
			filtered = append(filtered, item)
			continue
		}
		typ, _ := m["type"].(string)

		// 仅修正真正的 tool/function call 标识，避免误改普通 message/reasoning id；
		// 若 item_reference 指向 legacy call_* 标识，则仅修正该引用本身。
		fixCallIDPrefix := func(id string) string {
			if id == "" || strings.HasPrefix(id, "fc") {
				return id
			}
			if strings.HasPrefix(id, "call_") {
				return "fc" + strings.TrimPrefix(id, "call_")
			}
			return "fc_" + id
		}
		normalizeLegacyMessageID := func(id string) string {
			if len(id) < len("item_") || !strings.EqualFold(id[:len("item_")], "item_") {
				return id
			}
			return "msg_" + id[len("item_"):]
		}
		isToolReferenceID := func(id string) bool {
			id = strings.TrimSpace(id)
			return strings.HasPrefix(id, "call_") || strings.HasPrefix(id, "fc")
		}

		if typ == "item_reference" {
			if !preserveReferences {
				continue
			}
			newItem := make(map[string]any, len(m))
			for key, value := range m {
				newItem[key] = value
			}
			if id, ok := newItem["id"].(string); ok {
				if storeDisabled && dropStoreFalseNativeItemReferences && !isToolReferenceID(id) {
					droppedNativeItemReferenceCount++
					continue
				}
				if isToolReferenceID(id) {
					newItem["id"] = fixCallIDPrefix(id)
				}
			}
			filtered = append(filtered, newItem)
			continue
		}

		newItem := m
		copied := false
		// 仅在需要修改字段时创建副本，避免直接改写原始输入。
		ensureCopy := func() {
			if copied {
				return
			}
			newItem = make(map[string]any, len(m))
			for key, value := range m {
				newItem[key] = value
			}
			copied = true
		}

		if isCodexToolCallItemType(typ) {
			callID, ok := m["call_id"].(string)
			if !ok || strings.TrimSpace(callID) == "" {
				if id, ok := m["id"].(string); ok && strings.TrimSpace(id) != "" {
					callID = id
					ensureCopy()
					newItem["call_id"] = callID
				}
			}

			if callID != "" {
				fixedCallID := fixCallIDPrefix(callID)
				if fixedCallID != callID {
					ensureCopy()
					newItem["call_id"] = fixedCallID
				}
			}
		}

		if typ == "message" {
			if id, ok := m["id"].(string); ok {
				normalizedID := normalizeLegacyMessageID(id)
				if normalizedID != id {
					ensureCopy()
					newItem["id"] = normalizedID
				}
			}
		}

		if !preserveReferences {
			ensureCopy()
			delete(newItem, "id")
			if !isCodexToolCallItemType(typ) {
				delete(newItem, "call_id")
			}
		}

		if codexInputItemRequiresName(typ) {
			if strings.TrimSpace(firstNonEmptyString(newItem["name"])) == "" {
				name := firstNonEmptyString(newItem["tool_name"])
				if name == "" {
					if function, ok := newItem["function"].(map[string]any); ok {
						name = firstNonEmptyString(function["name"])
					}
				}
				if name == "" {
					name = "tool"
				}
				ensureCopy()
				newItem["name"] = name
			}
		}

		filtered = append(filtered, newItem)
	}
	return filtered, droppedNativeItemReferenceCount
}

func isCodexToolCallItemType(typ string) bool {
	typ = strings.TrimSpace(typ)
	if typ == "" {
		return false
	}
	switch typ {
	case "tool_search_output":
		return true
	default:
		return strings.HasSuffix(typ, "_call") || strings.HasSuffix(typ, "_call_output")
	}
}

func codexInputItemRequiresName(typ string) bool {
	switch strings.TrimSpace(typ) {
	case "function_call", "custom_tool_call", "mcp_tool_call":
		return true
	default:
		return false
	}
}

func normalizeCodexTools(reqBody map[string]any) bool {
	rawTools, ok := reqBody["tools"]
	if !ok || rawTools == nil {
		return false
	}
	tools, ok := rawTools.([]any)
	if !ok {
		return false
	}

	modified := false
	validTools := make([]any, 0, len(tools))

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			// Keep unknown structure as-is to avoid breaking upstream behavior.
			validTools = append(validTools, tool)
			continue
		}

		toolType, _ := toolMap["type"].(string)
		toolType = strings.TrimSpace(toolType)
		if toolType != "function" {
			validTools = append(validTools, toolMap)
			continue
		}

		// ChatCompletions-style tools use {type:"function", function:{...}}.
		functionValue, hasFunction := toolMap["function"]
		function, ok := functionValue.(map[string]any)
		if !hasFunction || functionValue == nil || !ok {
			function = nil
		}

		name := strings.TrimSpace(firstNonEmptyString(toolMap["name"], toolMap["function_name"]))
		if name == "" && function != nil {
			name = strings.TrimSpace(firstNonEmptyString(function["name"]))
		}
		if name == "" {
			// Drop function tools with no callable name.
			modified = true
			continue
		}

		if currentName := strings.TrimSpace(firstNonEmptyString(toolMap["name"])); currentName == "" || currentName != name {
			toolMap["name"] = name
			modified = true
		}

		if _, ok := toolMap["function_name"]; ok {
			delete(toolMap, "function_name")
			modified = true
		}

		if _, ok := toolMap["description"]; !ok && function != nil {
			if desc, ok := function["description"].(string); ok && strings.TrimSpace(desc) != "" {
				toolMap["description"] = desc
				modified = true
			}
		}
		if params, ok := toolMap["parameters"]; ok {
			normalizedParams, changed := normalizeCodexToolParameters(params)
			if changed {
				toolMap["parameters"] = normalizedParams
				modified = true
			}
		} else if function != nil {
			if params, ok := function["parameters"]; ok {
				normalizedParams, _ := normalizeCodexToolParameters(params)
				toolMap["parameters"] = normalizedParams
				modified = true
			} else {
				toolMap["parameters"] = defaultCodexToolParameters()
				modified = true
			}
		} else {
			toolMap["parameters"] = defaultCodexToolParameters()
			modified = true
		}
		if _, ok := toolMap["strict"]; !ok && function != nil {
			if strict, ok := function["strict"]; ok {
				toolMap["strict"] = strict
				modified = true
			}
		}

		validTools = append(validTools, toolMap)
	}

	if modified {
		reqBody["tools"] = validTools
	}

	return modified
}

func normalizeCodexToolParameters(value any) (any, bool) {
	if value == nil {
		return defaultCodexToolParameters(), true
	}

	params, ok := value.(map[string]any)
	if !ok {
		return defaultCodexToolParameters(), true
	}

	typ := strings.TrimSpace(firstNonEmptyString(params["type"]))
	if typ == "" {
		normalized := copyStringAnyMap(params)
		normalized["type"] = "object"
		if _, ok := normalized["properties"]; !ok {
			normalized["properties"] = map[string]any{}
		}
		return normalized, true
	}

	if typ == "object" {
		if properties, ok := params["properties"]; !ok || properties == nil {
			normalized := copyStringAnyMap(params)
			normalized["properties"] = map[string]any{}
			return normalized, true
		}
	}

	return value, false
}

func defaultCodexToolParameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func copyStringAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+2)
	for key, value := range in {
		out[key] = value
	}
	return out
}
