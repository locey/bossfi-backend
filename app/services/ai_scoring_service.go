package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bossfi-backend/config"
	"bossfi-backend/db/database"
	"bossfi-backend/models"

	"github.com/sirupsen/logrus"
)

// 评分状态常量
const (
	ScoreStatusPending = 0  // 待评分
	ScoreStatusScoring = 1  // 评分中
	ScoreStatusSuccess = 2  // 评分成功
	ScoreStatusFailed  = -1 // 评分失败
)

type AIScoringService struct{}

func NewAIScoringService() *AIScoringService {
	return &AIScoringService{}
}

// AiHubMixRequest AiHubMix API请求结构
type AiHubMixRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AiHubMixResponse AiHubMix API响应结构
type AiHubMixResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ScoreArticle 对文章进行AI评分
func (s *AIScoringService) ScoreArticle(articleID uint) error {
	// 获取文章信息
	var article models.Article
	if err := database.DB.Where("id = ? AND is_deleted = ?", articleID, false).First(&article).Error; err != nil {
		return fmt.Errorf("article not found: %v", err)
	}

	// 检查评分状态，避免重复评分
	if article.ScoreStatus == ScoreStatusScoring {
		return fmt.Errorf("article %d is already being scored", articleID)
	}

	// 更新状态为评分中
	if err := database.DB.Model(&article).Update("score_status", ScoreStatusScoring).Error; err != nil {
		logrus.Errorf("Failed to update scoring status for article %d: %v", articleID, err)
		return err
	}

	// 构建评分提示
	prompt := s.buildScoringPrompt(article.Title, article.Content)

	// 调用AiHubMix API
	score, reason, err := s.callAiHubMixAPI(prompt)
	if err != nil {
		logrus.Errorf("Failed to call AiHubMix API for article %d: %v", articleID, err)
		// 更新状态为评分失败
		database.DB.Model(&article).Update("score_status", ScoreStatusFailed)
		return err
	}

	// 更新文章评分和状态
	now := time.Now()
	updateData := map[string]interface{}{
		"score":        score,
		"score_time":   now,
		"score_reason": reason,
		"score_status": ScoreStatusSuccess,
	}

	if err := database.DB.Model(&article).Updates(updateData).Error; err != nil {
		logrus.Errorf("Failed to update article score: %v", err)
		// 更新状态为评分失败
		database.DB.Model(&article).Update("score_status", ScoreStatusFailed)
		return err
	}

	logrus.Infof("Successfully scored article %d with score %.2f", articleID, score)
	return nil
}

// buildScoringPrompt 构建评分提示
func (s *AIScoringService) buildScoringPrompt(title, content string) string {
	return fmt.Sprintf(`请对以下文章进行评分和分析。评分标准如下：

1. 内容质量 (40%): 信息准确性、深度、原创性
2. 表达清晰度 (30%): 语言流畅性、逻辑结构、可读性
3. 价值贡献 (20%): 对读者的实用价值、启发性
4. 创新性 (10%): 观点新颖性、独特见解

请给出：
1. 总体评分 (0-10分，保留两位小数)
2. 详细评分理由 (200字以内)

文章标题：%s
文章内容：%s

请以JSON格式回复，格式如下：
{
  "score": 8.5,
  "reason": "详细评分理由..."
}`, title, content)
}

// callAiHubMixAPI 调用AiHubMix API
func (s *AIScoringService) callAiHubMixAPI(prompt string) (float64, string, error) {
	requestBody := AiHubMixRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", config.AppConfig.AiHubMix.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AppConfig.AiHubMix.APIKey)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// 解析响应
	var response AiHubMixResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(response.Choices) == 0 {
		return 0, "", fmt.Errorf("no choices in response")
	}

	content := response.Choices[0].Message.Content

	// 解析JSON响应
	var result struct {
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// 如果JSON解析失败，尝试提取数字作为评分
		logrus.Warnf("Failed to parse JSON response, attempting to extract score from text: %s", content)
		return 5.0, "AI评分解析失败，使用默认评分", nil
	}

	// 验证评分范围
	if result.Score < 0 || result.Score > 10 {
		result.Score = 5.0
		result.Reason = "评分超出范围，使用默认评分"
	}

	return result.Score, result.Reason, nil
}

// GetArticleScore 获取文章评分
func (s *AIScoringService) GetArticleScore(articleID uint) (*models.Article, error) {
	var article models.Article
	err := database.DB.Where("id = ? AND is_deleted = ?", articleID, false).First(&article).Error
	if err != nil {
		return nil, fmt.Errorf("article not found: %v", err)
	}

	return &article, nil
}

// ScoreMultipleArticles 批量评分文章
func (s *AIScoringService) ScoreMultipleArticles(articleIDs []uint) error {
	for _, articleID := range articleIDs {
		if err := s.ScoreArticle(articleID); err != nil {
			logrus.Errorf("Failed to score article %d: %v", articleID, err)
			// 继续处理其他文章，不中断整个流程
			continue
		}
		// 添加延迟避免API限制
		time.Sleep(1 * time.Second)
	}
	return nil
}

// GetArticlesWithoutScore 获取未评分的文章
func (s *AIScoringService) GetArticlesWithoutScore(limit int) ([]models.Article, error) {
	var articles []models.Article
	err := database.DB.Where("is_deleted = ? AND (score IS NULL OR score_status = ?)", false, ScoreStatusPending).
		Order("created_at desc").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// GetArticlesByScoreStatus 根据评分状态获取文章
func (s *AIScoringService) GetArticlesByScoreStatus(status int, limit int) ([]models.Article, error) {
	var articles []models.Article
	err := database.DB.Where("is_deleted = ? AND score_status = ?", false, status).
		Order("created_at desc").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// ResetScoreStatus 重置评分状态
func (s *AIScoringService) ResetScoreStatus(articleID uint) error {
	return database.DB.Model(&models.Article{}).
		Where("id = ? AND is_deleted = ?", articleID, false).
		Update("score_status", ScoreStatusPending).Error
}

// RetryFailedScoring 重试失败的评分
func (s *AIScoringService) RetryFailedScoring(limit int) (int, int, error) {
	// 获取评分失败的文章
	failedArticles, err := s.GetArticlesByScoreStatus(ScoreStatusFailed, limit)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get failed articles: %v", err)
	}

	successCount := 0
	failCount := 0

	for _, article := range failedArticles {
		// 重置状态为待评分
		if err := s.ResetScoreStatus(article.ID); err != nil {
			logrus.Errorf("Failed to reset score status for article %d: %v", article.ID, err)
			failCount++
			continue
		}

		// 重新评分
		if err := s.ScoreArticle(article.ID); err != nil {
			logrus.Errorf("Failed to retry scoring article %d: %v", article.ID, err)
			failCount++
		} else {
			successCount++
		}

		// 添加延迟避免API限制
		time.Sleep(2 * time.Second)
	}

	return successCount, failCount, nil
}

// RetryPendingScoring 重试待评分的文章
func (s *AIScoringService) RetryPendingScoring(limit int) (int, int, error) {
	// 获取待评分的文章
	pendingArticles, err := s.GetArticlesByScoreStatus(ScoreStatusPending, limit)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get pending articles: %v", err)
	}

	successCount := 0
	failCount := 0

	for _, article := range pendingArticles {
		// 直接评分
		if err := s.ScoreArticle(article.ID); err != nil {
			logrus.Errorf("Failed to score pending article %d: %v", article.ID, err)
			failCount++
		} else {
			successCount++
		}

		// 添加延迟避免API限制
		time.Sleep(2 * time.Second)
	}

	return successCount, failCount, nil
}
