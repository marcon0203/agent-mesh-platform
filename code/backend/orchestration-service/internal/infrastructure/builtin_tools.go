package infrastructure

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// builtinTools 是内置 Tool 的进程内实现，key 是 marketplace-service 里
// builtin_capabilities.go 注册的 Capability.ID（约定为固定数值 "1"/"2"）。
// 内置能力免走 MCP，直接在本进程内被 buildToolsConfig 解析成 Eino BaseTool，
// 对应技术规格文档 §6.2「内置能力走进程内调用」的方案。
var builtinTools = map[string]tool.BaseTool{
	"1": mustInferTool(webFetchTool()),
	"2": mustInferTool(calendarParseTool()),
}

func mustInferTool(t tool.InvokableTool, err error) tool.InvokableTool {
	if err != nil {
		panic("failed to build builtin tool: " + err.Error())
	}
	return t
}

type webFetchInput struct {
	URL string `json:"url" jsonschema:"required" jsonschema_description:"要抓取的网页 URL，必须是完整地址（含 http/https）"`
}

type webFetchOutput struct {
	Text string `json:"text"`
}

// webFetchTool 对应市场里的"网页检索"内置能力：抓取一个 URL，返回截断后的正文文本。
// 不做 HTML 结构化解析（不引入额外依赖），只做基础的标签剥离。
func webFetchTool() (tool.InvokableTool, error) {
	return utils.InferTool[webFetchInput, webFetchOutput]("web_fetch", "抓取指定 URL 的网页正文文本", func(ctx context.Context, in webFetchInput) (webFetchOutput, error) {
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, in.URL, nil)
		if err != nil {
			return webFetchOutput{}, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return webFetchOutput{}, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return webFetchOutput{}, err
		}

		text := stripHTMLTags(string(body))
		const maxLen = 4000
		if len(text) > maxLen {
			text = text[:maxLen] + "...(截断)"
		}
		return webFetchOutput{Text: text}, nil
	})
}

var htmlTagPattern = regexp.MustCompile(`(?s)<[^>]*>`)
var whitespacePattern = regexp.MustCompile(`\s+`)

func stripHTMLTags(html string) string {
	text := htmlTagPattern.ReplaceAllString(html, " ")
	return strings.TrimSpace(whitespacePattern.ReplaceAllString(text, " "))
}

type calendarParseInput struct {
	Text string `json:"text" jsonschema:"required" jsonschema_description:"包含日期时间的自然语言，例如“明天下午3点”“周五 15:00”"`
}

type calendarParseOutput struct {
	ISO8601 string `json:"iso8601"`
	Matched bool   `json:"matched"`
}

// calendarParseTool 对应市场里的"日历解析"内置能力：把常见中文相对日期/时间表达
// 解析成 ISO8601 时间戳。覆盖有限的常见模式，不是完整 NLP 实现。
func calendarParseTool() (tool.InvokableTool, error) {
	return utils.InferTool[calendarParseInput, calendarParseOutput]("calendar_parse", "把中文自然语言日期时间表达解析为 ISO8601", func(ctx context.Context, in calendarParseInput) (calendarParseOutput, error) {
		t, ok := parseChineseDateTime(in.Text, time.Now())
		if !ok {
			return calendarParseOutput{Matched: false}, nil
		}
		return calendarParseOutput{ISO8601: t.Format(time.RFC3339), Matched: true}, nil
	})
}

var (
	dayOffsetPattern  = regexp.MustCompile(`今天|明天|后天|大后天`)
	hourMinutePattern = regexp.MustCompile(`([0-2]?[0-9])[:：]([0-5][0-9])`)
	hourOnlyPattern   = regexp.MustCompile(`([0-2]?[0-9])\s*点`)
)

func parseChineseDateTime(text string, now time.Time) (time.Time, bool) {
	dayOffsets := map[string]int{"今天": 0, "明天": 1, "后天": 2, "大后天": 3}
	offset, matchedDay := 0, false
	if m := dayOffsetPattern.FindString(text); m != "" {
		offset = dayOffsets[m]
		matchedDay = true
	}

	hour, minute, matchedTime := 0, 0, false
	if m := hourMinutePattern.FindStringSubmatch(text); len(m) == 3 {
		hour = atoiSafe(m[1])
		minute = atoiSafe(m[2])
		matchedTime = true
	} else if m := hourOnlyPattern.FindStringSubmatch(text); len(m) == 2 {
		hour = atoiSafe(m[1])
		matchedTime = true
	}
	if matchedTime && strings.Contains(text, "下午") && hour < 12 {
		hour += 12
	}
	if matchedTime && strings.Contains(text, "晚上") && hour < 12 {
		hour += 12
	}

	if !matchedDay && !matchedTime {
		return time.Time{}, false
	}

	base := now.AddDate(0, 0, offset)
	return time.Date(base.Year(), base.Month(), base.Day(), hour, minute, 0, 0, base.Location()), true
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}
