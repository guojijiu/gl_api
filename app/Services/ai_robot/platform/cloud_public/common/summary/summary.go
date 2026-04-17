package summary

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud-platform-api/app/Services/ai_gateway/llm"
)

// maxAPIJSONForPrompt 限制注入给大模型的云平台原始 JSON 长度，
// 避免 prompt 过长导致成本/延迟飙升、甚至因为 token 上限或模型解析失败而变差。
// 说明：这里按字节长度截断即可（Go string 的 len 也是字节数），目的是控制 token 规模。
const maxAPIJSONForPrompt = 8000

func truncateAPIJSONForPrompt(s string) string {
	if len(s) <= maxAPIJSONForPrompt {
		return s
	}
	return s[:maxAPIJSONForPrompt] + "\n...[truncated]..."
}

func SummarizeWithData(ctx context.Context, c llm.ChatCompletionClient, userQuestion string, contextSummary string, apiJSON []byte) (string, error) {
	compressedJSON := compressAPIJSONForPrompt(apiJSON)
	prompt := buildSummaryPrompt(userQuestion, contextSummary, compressedJSON)
	return c.ChatCompletion(ctx, prompt)
}

func SummarizeProjectData(ctx context.Context, c llm.ChatCompletionClient, userQuestion string, contextSummary string, apiJSON []byte) (string, error) {
	compressedJSON := compressProjectJSONForPrompt(apiJSON)
	return c.ChatCompletion(ctx, buildSummaryPrompt(userQuestion, contextSummary, compressedJSON))
}

func SummarizeTaskData(ctx context.Context, c llm.ChatCompletionClient, userQuestion string, contextSummary string, apiJSON []byte) (string, error) {
	compressedJSON := compressTaskJSONForPrompt(apiJSON)
	return c.ChatCompletion(ctx, buildSummaryPrompt(userQuestion, contextSummary, compressedJSON))
}

func SummarizeProjectArticleData(ctx context.Context, c llm.ChatCompletionClient, userQuestion string, contextSummary string, apiJSON []byte) (string, error) {
	compressedJSON := compressProjectArticleJSONForPrompt(apiJSON)
	return c.ChatCompletion(ctx, buildSummaryPrompt(userQuestion, contextSummary, compressedJSON))
}

func buildSummaryPrompt(userQuestion string, contextSummary string, compressedJSON string) string {
	return fmt.Sprintf(`你是云平台客服助手。请严格按以下规则回复，且只输出最终给客户的答复，不要任何额外信息。

用户问题：
%s

会话上下文摘要（可能为空；仅在有帮助时参考，不要编造未确认的上下文）：
%s

云平台接口返回 JSON（可能包含 code、showMsg、content 等字段）：
%s

回复规则（按优先级）：
1）若 code=2（账号登录问题）：
   - 若 showMsg 表示“账号登录已失效”，按照showMsg，重新组装语言和客户说明并且告知下一步应该怎么操作（如重新登录）。
   - 其他登录问题：根据 showMsg 翻译成客户易识别中文，直接告知客户下一步处理（如重新登录），不要解释系统细节。
2）若接口结果正常且存在有效数据：用简短、清晰的中文直接回答用户问题。
3）若接口结果正常但没有有效数据（如列表为空、content 为空等）：只回复“未查询到相关数据”。
4）若接口错误（如 code 不为 1 或存在明确错误信息）：根据 showMsg 翻译成客户易识别的中文；若 showMsg 缺失，则只回复“未查询到相关数据”。
5）若压缩 JSON 中出现列表总条数 count（或 total）大于 5，或 omitted_count 大于 0：说明结果未全部展示。请在答复末尾另起一行，用一句话提示用户可回复「查看更多」或「下一页」，或在会话消息列表页面分页查看；不要编造未出现在样本中的具体条目。

格式规则：
- 输出 Markdown 正文，不要代码块。
- 仅输出客户可见内容，不要标题“结果/说明”等包装词。
- 有多条结果时可用 Markdown 列表；单条结果时用一行文本。

禁止项：
- 不要礼貌寒暄
- 不要解释规则
- 不要补充与当前问题无关的建议（但允许按上一条规则给出分页/查看更多引导）
- 不要编造任何接口中不存在的信息
- 不要出现“根据接口返回/接口显示/code/showMsg/系统层面/客户无需关注细节”等描述
`, userQuestion, contextSummary, compressedJSON)
}

func SummarizeUnsupported(ctx context.Context, c llm.ChatCompletionClient, userQuestion string, contextSummary string, reason string) (string, error) {
	prompt := fmt.Sprintf(`你是云平台客服助手。请严格按以下规则回复，且只输出最终给客户的答复，不要任何额外信息。

用户问题：%s
会话上下文摘要：%s
系统判断信息：%s

回复规则（按优先级）：
1）若系统判断信息表明 code=2（账号登录问题）：
   - 若 showMsg 表示“账号登录已失效”，按照showMsg，重新组装语言和客户说明并且告知下一步应该怎么操作（如重新登录）。
   - 其他登录问题：根据信息翻译成客户易识别中文，直接告知客户下一步处理（如重新登录），不要解释系统细节。
2）若系统判断信息中有明确可用结果：用简短、清晰的中文直接回答。
3）若无结果：只回复“未查询到相关数据”。
4）若属于错误或不支持：根据系统判断信息翻译成客户易识别的中文。

格式规则：
- 输出 Markdown 正文，不要代码块。
- 仅输出客户可见内容，不要标题“结果/说明”等包装词。
- 有多条结果时可用 Markdown 列表；单条结果时用一行文本。

禁止项：
- 不要礼貌寒暄
- 不要解释规则
- 不要补充建议
- 不要承诺具体时间
- 不要添加任何额外信息
- 不要出现“根据接口返回/接口显示/code/showMsg/系统层面/客户无需关注细节”等描述
`, userQuestion, contextSummary, reason)
	return c.ChatCompletion(ctx, prompt)
}

func mustJSON(v interface{}, fallback []byte) string {
	b, err := json.Marshal(v)
	if err != nil {
		return compressAPIJSONForPrompt(fallback)
	}
	return truncateAPIJSONForPrompt(string(b))
}
