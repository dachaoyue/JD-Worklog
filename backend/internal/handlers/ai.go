package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/example/worklog-system/internal/config"
)

const workReportTemplate = `### 今日工作总结：

**1、【[项目/模块名称]】**
工作耗时：[X]小时
工作内容：[详细描述做了哪些工作，如：开发了XX模块、修复了XX bug (ITR号)]
遗留问题：[今日未完成的任务、遇到的阻碍或依赖]
推进计划：[针对遗留问题的下一步行动、需要协调的资源]

**2、【[项目/模块名称]】**
工作耗时：[X]小时
工作内容：[详细描述]
遗留问题：[详细描述]
推进计划：[详细描述]

*(...根据需要可罗列N个项目)*

### 明日工作计划：

**1、【[项目/模块名称]】**
* [简述明日计划的任务1]
* [简述明日计划的任务2]

**2、【[项目/模块名称]】**
* [简述明日计划的任务1]`

type AIHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

type GenerateReportRequest struct {
	Date         string   `json:"date"`
	ExtraContent string   `json:"extra_content"`
	Worklogs     []string `json:"worklogs"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (h *AIHandler) GenerateReport(c *gin.Context) {
	var req GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Worklogs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Worklogs cannot be empty."})
		return
	}

	prompt := "基于以下工作内容生成一份工作日报：\n"
	for _, log := range req.Worklogs {
		prompt += "- " + log + "\n"
	}
	if req.ExtraContent != "" {
		prompt += "附加信息：" + req.ExtraContent + "\n"
	}

	apiKey := strings.TrimSpace(h.Cfg.DeepseekAPIKey)
	if apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "未配置 DEEPSEEK_API_KEY，请在环境变量或 run-dev.bat 中设置 API Key",
		})
		return
	}

	model := h.Cfg.DeepseekModel
	if model == "" {
		model = "glm-5.1"
	}

	report, status, errMsg, details := h.callAnthropicMessages(apiKey, model, workReportTemplate, prompt)
	if errMsg != "" {
		c.JSON(status, gin.H{"error": errMsg, "details": details})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

func (h *AIHandler) callAnthropicMessages(apiKey, model, system, userPrompt string) (report string, httpStatus int, errMsg, details string) {
	chatReq := anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		System:    system,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		return "", http.StatusInternalServerError, "Failed to create request body", ""
	}

	apiURL := anthropicMessagesURL(h.Cfg.DeepseekAPIBase)
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", http.StatusInternalServerError, "Failed to create HTTP request", ""
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", http.StatusInternalServerError, "Failed to call AI gateway", err.Error()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", http.StatusInternalServerError, "Failed to read AI gateway response", ""
	}

	if resp.StatusCode != http.StatusOK {
		msg := parseAIError(body)
		if resp.StatusCode == http.StatusUnauthorized {
			return "", http.StatusBadGateway, "AI 网关认证失败，请检查 DEEPSEEK_API_KEY", msg
		}
		return "", http.StatusBadGateway, "AI 网关调用失败", msg
	}

	var chatResp anthropicResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", http.StatusInternalServerError, "Failed to parse AI gateway response", string(body)
	}

	for _, block := range chatResp.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text, 0, "", ""
		}
	}

	return "", http.StatusInternalServerError, "AI 网关未返回有效内容", string(body)
}

func anthropicMessagesURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(base, "/v1/messages") {
		return base
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/messages"
	}
	return base + "/v1/messages"
}

func parseAIError(body []byte) string {
	var wrapped struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Error.Message != "" {
		return wrapped.Error.Message
	}
	if len(body) > 500 {
		return string(body[:500])
	}
	return string(body)
}
