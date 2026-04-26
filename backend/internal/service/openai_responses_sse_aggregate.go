package service

import (
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
)

type responsesStreamAccumulator struct {
	responseID    string
	model         string
	object        string
	status        string
	usage         *apicompat.ResponsesUsage
	incomplete    *apicompat.ResponsesIncompleteDetails
	responseErr   *apicompat.ResponsesError
	outputByIndex map[int]*apicompat.ResponsesOutput

	terminalEventType string
	terminalResponse  *apicompat.ResponsesResponse
}

func newResponsesStreamAccumulator() *responsesStreamAccumulator {
	return &responsesStreamAccumulator{
		outputByIndex: make(map[int]*apicompat.ResponsesOutput),
	}
}

func (a *responsesStreamAccumulator) ApplyEvent(evt *apicompat.ResponsesStreamEvent) {
	if evt == nil {
		return
	}

	a.captureResponseMeta(evt.Type, evt.Response)

	switch evt.Type {
	case "response.output_item.added", "response.output_item.done":
		if evt.Item != nil {
			a.mergeOutputItem(evt.OutputIndex, evt.Item)
		}
	case "response.output_text.delta", "response.output_text.done":
		a.appendOutputText(evt.OutputIndex, evt.ContentIndex, firstNonEmpty(evt.Delta, evt.Text))
	case "response.function_call_arguments.delta", "response.function_call_arguments.done":
		a.appendFunctionCallArguments(evt.OutputIndex, evt.CallID, evt.Name, firstNonEmpty(evt.Delta, evt.Arguments))
	case "response.reasoning_summary_text.delta", "response.reasoning_summary_text.done":
		a.appendReasoningSummary(evt.OutputIndex, evt.SummaryIndex, firstNonEmpty(evt.Delta, evt.Text))
	case "response.completed", "response.incomplete", "response.failed", "response.done":
		a.terminalEventType = evt.Type
		if evt.Response != nil {
			a.terminalResponse = cloneResponsesResponse(evt.Response)
		}
	}
}

func (a *responsesStreamAccumulator) FinalResponse() (*apicompat.ResponsesResponse, bool) {
	if a.terminalResponse == nil {
		return nil, false
	}

	resp := cloneResponsesResponse(a.terminalResponse)
	if resp.ID == "" {
		resp.ID = a.responseID
	}
	if resp.Object == "" {
		resp.Object = a.object
	}
	if resp.Object == "" {
		resp.Object = "response"
	}
	if resp.Model == "" {
		resp.Model = a.model
	}
	if resp.Status == "" {
		resp.Status = firstNonEmpty(a.status, statusFromResponsesEventType(a.terminalEventType))
	}
	if resp.Usage == nil && a.usage != nil {
		resp.Usage = cloneResponsesUsage(a.usage)
	}
	if resp.IncompleteDetails == nil && a.incomplete != nil {
		resp.IncompleteDetails = cloneResponsesIncompleteDetails(a.incomplete)
	}
	if resp.Error == nil && a.responseErr != nil {
		resp.Error = cloneResponsesError(a.responseErr)
	}
	resp.Output = mergeResponsesOutputSlices(resp.Output, a.aggregatedOutput())

	return resp, true
}

func (a *responsesStreamAccumulator) TerminalEventType() string {
	return a.terminalEventType
}

func (a *responsesStreamAccumulator) captureResponseMeta(eventType string, resp *apicompat.ResponsesResponse) {
	if resp == nil {
		return
	}
	if resp.ID != "" {
		a.responseID = resp.ID
	}
	if resp.Model != "" {
		a.model = resp.Model
	}
	if resp.Object != "" {
		a.object = resp.Object
	}
	if resp.Status != "" {
		a.status = resp.Status
	} else if status := statusFromResponsesEventType(eventType); status != "" {
		a.status = status
	}
	if resp.Usage != nil {
		a.usage = cloneResponsesUsage(resp.Usage)
	}
	if resp.IncompleteDetails != nil {
		a.incomplete = cloneResponsesIncompleteDetails(resp.IncompleteDetails)
	}
	if resp.Error != nil {
		a.responseErr = cloneResponsesError(resp.Error)
	}
}

func (a *responsesStreamAccumulator) mergeOutputItem(outputIndex int, item *apicompat.ResponsesOutput) {
	if item == nil {
		return
	}
	existing := a.ensureOutputItem(outputIndex, item.Type)
	merged := mergeResponsesOutput(*existing, *item)
	if item.Status != "" {
		merged.Status = item.Status
	}
	a.outputByIndex[outputIndex] = &merged
}

func (a *responsesStreamAccumulator) appendOutputText(outputIndex, contentIndex int, text string) {
	if text == "" {
		return
	}
	item := a.ensureOutputItem(outputIndex, "message")
	part := ensureResponsesContentPart(item, contentIndex)
	if part.Type == "" {
		part.Type = "output_text"
	}
	part.Text = mergeAccumulatedResponsesText(part.Text, text)
}

func (a *responsesStreamAccumulator) appendFunctionCallArguments(outputIndex int, callID, name, arguments string) {
	if arguments == "" && callID == "" && name == "" {
		return
	}
	item := a.ensureOutputItem(outputIndex, "function_call")
	if item.CallID == "" {
		item.CallID = callID
	}
	if item.Name == "" {
		item.Name = name
	}
	item.Arguments += arguments
}

func (a *responsesStreamAccumulator) appendReasoningSummary(outputIndex, summaryIndex int, text string) {
	if text == "" {
		return
	}
	item := a.ensureOutputItem(outputIndex, "reasoning")
	summary := ensureResponsesSummary(item, summaryIndex)
	if summary.Type == "" {
		summary.Type = "summary_text"
	}
	summary.Text += text
}

func (a *responsesStreamAccumulator) ensureOutputItem(outputIndex int, typ string) *apicompat.ResponsesOutput {
	item := a.outputByIndex[outputIndex]
	if item == nil {
		item = &apicompat.ResponsesOutput{}
		a.outputByIndex[outputIndex] = item
	}
	if item.Type == "" {
		item.Type = typ
	}
	if item.Type == "message" && item.Role == "" {
		item.Role = "assistant"
	}
	return item
}

func (a *responsesStreamAccumulator) aggregatedOutput() []apicompat.ResponsesOutput {
	if len(a.outputByIndex) == 0 {
		return nil
	}
	keys := make([]int, 0, len(a.outputByIndex))
	for idx, item := range a.outputByIndex {
		if item == nil || item.Type == "" {
			continue
		}
		keys = append(keys, idx)
	}
	sort.Ints(keys)

	out := make([]apicompat.ResponsesOutput, 0, len(keys))
	for _, idx := range keys {
		out = append(out, *cloneResponsesOutput(a.outputByIndex[idx]))
	}
	return out
}

func ensureResponsesContentPart(item *apicompat.ResponsesOutput, contentIndex int) *apicompat.ResponsesContentPart {
	for len(item.Content) <= contentIndex {
		item.Content = append(item.Content, apicompat.ResponsesContentPart{})
	}
	return &item.Content[contentIndex]
}

func ensureResponsesSummary(item *apicompat.ResponsesOutput, summaryIndex int) *apicompat.ResponsesSummary {
	for len(item.Summary) <= summaryIndex {
		item.Summary = append(item.Summary, apicompat.ResponsesSummary{})
	}
	return &item.Summary[summaryIndex]
}

func mergeResponsesOutputSlices(base, overlay []apicompat.ResponsesOutput) []apicompat.ResponsesOutput {
	if len(base) == 0 {
		return cloneResponsesOutputs(overlay)
	}
	if len(overlay) == 0 {
		return cloneResponsesOutputs(base)
	}

	out := cloneResponsesOutputs(base)
	for i, item := range overlay {
		if i < len(out) {
			out[i] = mergeResponsesOutput(out[i], item)
			continue
		}
		out = append(out, *cloneResponsesOutput(&item))
	}
	return out
}

func mergeResponsesOutput(base, overlay apicompat.ResponsesOutput) apicompat.ResponsesOutput {
	out := *cloneResponsesOutput(&base)
	if out.Type == "" {
		out.Type = overlay.Type
	}
	if out.ID == "" {
		out.ID = overlay.ID
	}
	if out.Role == "" {
		out.Role = overlay.Role
	}
	if out.Status == "" {
		out.Status = overlay.Status
	}
	if out.EncryptedContent == "" {
		out.EncryptedContent = overlay.EncryptedContent
	}
	if out.CallID == "" {
		out.CallID = overlay.CallID
	}
	if out.Name == "" {
		out.Name = overlay.Name
	}
	if out.Arguments == "" {
		out.Arguments = overlay.Arguments
	}
	if out.Action == nil && overlay.Action != nil {
		out.Action = cloneWebSearchAction(overlay.Action)
	}
	out.Content = mergeResponsesContentParts(out.Content, overlay.Content)
	out.Summary = mergeResponsesSummaries(out.Summary, overlay.Summary)
	return out
}

func mergeResponsesContentParts(base, overlay []apicompat.ResponsesContentPart) []apicompat.ResponsesContentPart {
	if len(base) == 0 {
		return cloneResponsesContentParts(overlay)
	}
	if len(overlay) == 0 {
		return cloneResponsesContentParts(base)
	}

	out := cloneResponsesContentParts(base)
	for i, part := range overlay {
		if i < len(out) {
			if out[i].Type == "" {
				out[i].Type = part.Type
			}
			if out[i].Text == "" {
				out[i].Text = part.Text
			}
			if out[i].ImageURL == "" {
				out[i].ImageURL = part.ImageURL
			}
			continue
		}
		out = append(out, part)
	}
	return out
}

func mergeResponsesSummaries(base, overlay []apicompat.ResponsesSummary) []apicompat.ResponsesSummary {
	if len(base) == 0 {
		return cloneResponsesSummaries(overlay)
	}
	if len(overlay) == 0 {
		return cloneResponsesSummaries(base)
	}

	out := cloneResponsesSummaries(base)
	for i, summary := range overlay {
		if i < len(out) {
			if out[i].Type == "" {
				out[i].Type = summary.Type
			}
			if out[i].Text == "" {
				out[i].Text = summary.Text
			}
			continue
		}
		out = append(out, summary)
	}
	return out
}

func cloneResponsesResponse(resp *apicompat.ResponsesResponse) *apicompat.ResponsesResponse {
	if resp == nil {
		return nil
	}
	return &apicompat.ResponsesResponse{
		ID:                resp.ID,
		Object:            resp.Object,
		Model:             resp.Model,
		Status:            resp.Status,
		Output:            cloneResponsesOutputs(resp.Output),
		Usage:             cloneResponsesUsage(resp.Usage),
		IncompleteDetails: cloneResponsesIncompleteDetails(resp.IncompleteDetails),
		Error:             cloneResponsesError(resp.Error),
	}
}

func cloneResponsesOutputs(in []apicompat.ResponsesOutput) []apicompat.ResponsesOutput {
	if len(in) == 0 {
		return nil
	}
	out := make([]apicompat.ResponsesOutput, len(in))
	for i := range in {
		out[i] = *cloneResponsesOutput(&in[i])
	}
	return out
}

func cloneResponsesOutput(in *apicompat.ResponsesOutput) *apicompat.ResponsesOutput {
	if in == nil {
		return nil
	}
	return &apicompat.ResponsesOutput{
		Type:             in.Type,
		ID:               in.ID,
		Role:             in.Role,
		Content:          cloneResponsesContentParts(in.Content),
		Status:           in.Status,
		EncryptedContent: in.EncryptedContent,
		Summary:          cloneResponsesSummaries(in.Summary),
		CallID:           in.CallID,
		Name:             in.Name,
		Arguments:        in.Arguments,
		Action:           cloneWebSearchAction(in.Action),
	}
}

func cloneResponsesContentParts(in []apicompat.ResponsesContentPart) []apicompat.ResponsesContentPart {
	if len(in) == 0 {
		return nil
	}
	out := make([]apicompat.ResponsesContentPart, len(in))
	copy(out, in)
	return out
}

func cloneResponsesSummaries(in []apicompat.ResponsesSummary) []apicompat.ResponsesSummary {
	if len(in) == 0 {
		return nil
	}
	out := make([]apicompat.ResponsesSummary, len(in))
	copy(out, in)
	return out
}

func cloneResponsesUsage(in *apicompat.ResponsesUsage) *apicompat.ResponsesUsage {
	if in == nil {
		return nil
	}
	return &apicompat.ResponsesUsage{
		InputTokens:         in.InputTokens,
		OutputTokens:        in.OutputTokens,
		TotalTokens:         in.TotalTokens,
		InputTokensDetails:  cloneResponsesInputTokensDetails(in.InputTokensDetails),
		OutputTokensDetails: cloneResponsesOutputTokensDetails(in.OutputTokensDetails),
	}
}

func cloneResponsesInputTokensDetails(in *apicompat.ResponsesInputTokensDetails) *apicompat.ResponsesInputTokensDetails {
	if in == nil {
		return nil
	}
	return &apicompat.ResponsesInputTokensDetails{
		CachedTokens: in.CachedTokens,
	}
}

func cloneResponsesOutputTokensDetails(in *apicompat.ResponsesOutputTokensDetails) *apicompat.ResponsesOutputTokensDetails {
	if in == nil {
		return nil
	}
	return &apicompat.ResponsesOutputTokensDetails{
		ReasoningTokens: in.ReasoningTokens,
	}
}

func cloneResponsesIncompleteDetails(in *apicompat.ResponsesIncompleteDetails) *apicompat.ResponsesIncompleteDetails {
	if in == nil {
		return nil
	}
	return &apicompat.ResponsesIncompleteDetails{
		Reason: in.Reason,
	}
}

func cloneResponsesError(in *apicompat.ResponsesError) *apicompat.ResponsesError {
	if in == nil {
		return nil
	}
	return &apicompat.ResponsesError{
		Code:    in.Code,
		Message: in.Message,
	}
}

func cloneWebSearchAction(in *apicompat.WebSearchAction) *apicompat.WebSearchAction {
	if in == nil {
		return nil
	}
	return &apicompat.WebSearchAction{
		Type:  in.Type,
		Query: in.Query,
	}
}

func statusFromResponsesEventType(eventType string) string {
	switch eventType {
	case "response.completed", "response.done":
		return "completed"
	case "response.incomplete":
		return "incomplete"
	case "response.failed":
		return "failed"
	default:
		return ""
	}
}

func isResponsesTerminalUsageEvent(eventType string) bool {
	switch eventType {
	case "response.completed", "response.done", "response.incomplete", "response.failed":
		return true
	default:
		return false
	}
}

func openAIUsageFromResponsesResponse(resp *apicompat.ResponsesResponse) (OpenAIUsage, bool) {
	if resp == nil || resp.Usage == nil {
		return OpenAIUsage{}, false
	}

	usage := OpenAIUsage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
	}
	if resp.Usage.InputTokensDetails != nil {
		usage.CacheReadInputTokens = resp.Usage.InputTokensDetails.CachedTokens
	}
	return usage, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func mergeAccumulatedResponsesText(existing, incoming string) string {
	if existing == "" {
		return incoming
	}
	if incoming == "" {
		return existing
	}
	if strings.HasPrefix(incoming, existing) {
		return incoming
	}
	if strings.HasPrefix(existing, incoming) {
		return existing
	}

	maxOverlap := len(existing)
	if len(incoming) < maxOverlap {
		maxOverlap = len(incoming)
	}
	for overlap := maxOverlap; overlap > 0; overlap-- {
		if strings.HasSuffix(existing, incoming[:overlap]) {
			return existing + incoming[overlap:]
		}
	}

	return existing + incoming
}
