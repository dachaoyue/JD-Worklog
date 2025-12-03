package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

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

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model     string                  `json:"model"`
	Messages  []ChatCompletionMessage `json:"messages"`
	Stream    bool                    `json:"stream"`
	ExtraBody map[string]interface{}  `json:"extra_body,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
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

	apiKey := h.Cfg.DeepseekAPIKey

	chatReq := ChatCompletionRequest{
		Model: "deepseek-chat",
		Messages: []ChatCompletionMessage{
			{Role: "system", Content: workReportTemplate},
			{Role: "user", Content: prompt},
		},
		Stream: false,
		ExtraBody: map[string]interface{}{
			"thinking": map[string]string{
				"type": "enabled",
			},
		},
	}

	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request body"})
		return
	}

	httpReq, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create HTTP request"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call DeepSeek API"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read DeepSeek API response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "DeepSeek API returned an error", "details": string(body)})
		return
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse DeepSeek API response"})
		return
	}

	if len(chatResp.Choices) > 0 {
		c.JSON(http.StatusOK, gin.H{"report": chatResp.Choices[0].Message.Content})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No response from DeepSeek API"})
	}
}
