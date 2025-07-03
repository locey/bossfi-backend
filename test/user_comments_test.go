package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bossfi-backend/api/dto"
	"bossfi-backend/db/database"
	"bossfi-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestGetUserComments 测试获取用户评论功能
func TestGetUserComments(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 初始化数据库连接
	err := database.InitDB()
	assert.NoError(t, err)

	// 清理测试数据
	cleanupTestData(t)

	// 创建测试用户
	user1 := createTestUser(t, "testuser1", "0x1234567890123456789012345678901234567890")
	user2 := createTestUser(t, "testuser2", "0x1234567890123456789012345678901234567891")

	// 创建测试分类
	category := createTestCategory(t, "技术", "技术相关文章", "#FF5733")

	// 创建测试文章
	article1 := createTestArticle(t, user1.ID, category.ID, "测试文章1", "这是测试文章1的内容")
	article2 := createTestArticle(t, user2.ID, category.ID, "测试文章2", "这是测试文章2的内容")

	// 创建测试评论
	comment1 := createTestComment(t, user1.ID, article1.ID, nil, "这是第一条评论")
	comment2 := createTestComment(t, user1.ID, article2.ID, nil, "这是第二条评论")
	comment3 := createTestComment(t, user1.ID, article1.ID, &comment1.ID, "这是对第一条评论的回复")

	// 创建测试服务器
	router := setupTestRouter()

	// 获取用户token
	token := getTestToken(t, router, user1.Address)

	// 测试用例1: 获取用户评论列表（第一页）
	t.Run("GetUserComments_FirstPage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/comments?page=1&page_size=2", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserCommentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, int64(3), response.Total)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 2, response.PageSize)
		assert.Len(t, response.Comments, 2)

		// 验证评论按创建时间倒序排列
		assert.Equal(t, comment3.ID, response.Comments[0].ID)
		assert.Equal(t, comment2.ID, response.Comments[1].ID)
	})

	// 测试用例2: 获取用户评论列表（第二页）
	t.Run("GetUserComments_SecondPage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/comments?page=2&page_size=2", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserCommentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, int64(3), response.Total)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 2, response.PageSize)
		assert.Len(t, response.Comments, 1)

		assert.Equal(t, comment1.ID, response.Comments[0].ID)
	})

	// 测试用例3: 验证评论包含文章信息
	t.Run("GetUserComments_WithArticleInfo", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/comments?page=1&page_size=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserCommentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Len(t, response.Comments, 1)
		comment := response.Comments[0]

		// 验证文章信息
		assert.Equal(t, article1.ID, comment.Article.ID)
		assert.Equal(t, article1.Title, comment.Article.Title)
		assert.Contains(t, comment.Article.Content, "这是测试文章1的内容")
		assert.Equal(t, category.ID, *comment.Article.CategoryID)
		assert.Equal(t, category.Name, comment.Article.Category.Name)
		assert.Equal(t, category.Color, comment.Article.Category.Color)

		// 验证文章作者信息
		assert.Equal(t, user1.ID, comment.Article.User.ID)
		assert.Equal(t, user1.Username, comment.Article.User.Username)
	})

	// 测试用例4: 验证回复评论包含父评论信息
	t.Run("GetUserComments_WithParentInfo", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/comments?page=1&page_size=3", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserCommentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 找到回复评论
		var replyComment *dto.UserCommentResponse
		for i := range response.Comments {
			if response.Comments[i].ParentID != nil {
				replyComment = &response.Comments[i]
				break
			}
		}

		assert.NotNil(t, replyComment)
		assert.Equal(t, comment1.ID, *replyComment.ParentID)
		assert.NotNil(t, replyComment.Parent)

		// 验证父评论信息
		assert.Equal(t, comment1.ID, replyComment.Parent.ID)
		assert.Contains(t, replyComment.Parent.Content, "这是第一条评论")
		assert.Equal(t, user1.ID, replyComment.Parent.User.ID)
		assert.Equal(t, user1.Username, replyComment.Parent.User.Username)
	})

	// 测试用例5: 未认证访问
	t.Run("GetUserComments_Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/comments?page=1&page_size=10", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// 测试用例6: 无效的分页参数
	t.Run("GetUserComments_InvalidPagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/comments?page=0&page_size=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// 清理测试数据
	cleanupTestData(t)
}

// createTestComment 创建测试评论
func createTestComment(t *testing.T, userID, articleID uint, parentID *uint, content string) models.ArticleComment {
	comment := models.ArticleComment{
		UserID:    userID,
		ArticleID: articleID,
		ParentID:  parentID,
		Content:   content,
	}

	err := database.DB.Create(&comment).Error
	assert.NoError(t, err)

	return comment
}

// TestUserCommentsIntegration 集成测试：完整的用户评论流程
func TestUserCommentsIntegration(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 初始化数据库连接
	err := database.InitDB()
	assert.NoError(t, err)

	// 清理测试数据
	cleanupTestData(t)

	// 创建测试用户
	user1 := createTestUser(t, "testuser1", "0x1234567890123456789012345678901234567890")
	user2 := createTestUser(t, "testuser2", "0x1234567890123456789012345678901234567891")

	// 创建测试分类
	category := createTestCategory(t, "技术", "技术相关文章", "#FF5733")

	// 创建测试文章
	article1 := createTestArticle(t, user1.ID, category.ID, "测试文章1", "这是测试文章1的内容，包含了很多技术细节和实现方案")
	article2 := createTestArticle(t, user2.ID, category.ID, "测试文章2", "这是测试文章2的内容，主要讨论架构设计")

	// 创建测试服务器
	router := setupTestRouter()

	// 获取用户token
	token1 := getTestToken(t, router, user1.Address)
	token2 := getTestToken(t, router, user2.Address)

	// 测试完整流程
	t.Run("UserComments_CompleteFlow", func(t *testing.T) {
		// 1. 用户1在文章1上发表评论
		comment1Req := dto.CreateCommentRequest{
			ArticleID: article1.ID,
			Content:   "这是一条很长的评论内容，包含了很多想法和建议，希望能够对作者有所帮助",
		}
		comment1Data, _ := json.Marshal(comment1Req)
		req1, _ := http.NewRequest("POST", "/api/v1/comments", bytes.NewBuffer(comment1Data))
		req1.Header.Set("Authorization", "Bearer "+token1)
		req1.Header.Set("Content-Type", "application/json")

		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		var comment1Response dto.CommentResponse
		err := json.Unmarshal(w1.Body.Bytes(), &comment1Response)
		assert.NoError(t, err)

		// 2. 用户2回复用户1的评论
		comment2Req := dto.CreateCommentRequest{
			ArticleID: article1.ID,
			ParentID:  &comment1Response.ID,
			Content:   "回复用户1的评论，表示赞同并补充一些观点",
		}
		comment2Data, _ := json.Marshal(comment2Req)
		req2, _ := http.NewRequest("POST", "/api/v1/comments", bytes.NewBuffer(comment2Data))
		req2.Header.Set("Authorization", "Bearer "+token2)
		req2.Header.Set("Content-Type", "application/json")

		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// 3. 用户1在文章2上发表评论
		comment3Req := dto.CreateCommentRequest{
			ArticleID: article2.ID,
			Content:   "对文章2的评论，讨论架构设计的优缺点",
		}
		comment3Data, _ := json.Marshal(comment3Req)
		req3, _ := http.NewRequest("POST", "/api/v1/comments", bytes.NewBuffer(comment3Data))
		req3.Header.Set("Authorization", "Bearer "+token1)
		req3.Header.Set("Content-Type", "application/json")

		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusOK, w3.Code)

		// 4. 获取用户1的所有评论
		req4, _ := http.NewRequest("GET", "/api/v1/user/comments?page=1&page_size=10", nil)
		req4.Header.Set("Authorization", "Bearer "+token1)

		w4 := httptest.NewRecorder()
		router.ServeHTTP(w4, req4)
		assert.Equal(t, http.StatusOK, w4.Code)

		var userCommentsResponse dto.UserCommentListResponse
		err = json.Unmarshal(w4.Body.Bytes(), &userCommentsResponse)
		assert.NoError(t, err)

		// 验证用户1有2条评论
		assert.Equal(t, int64(2), userCommentsResponse.Total)
		assert.Len(t, userCommentsResponse.Comments, 2)

		// 验证评论按时间倒序排列
		assert.Equal(t, comment3Response.ID, userCommentsResponse.Comments[0].ID)
		assert.Equal(t, comment1Response.ID, userCommentsResponse.Comments[1].ID)

		// 验证文章信息
		for _, comment := range userCommentsResponse.Comments {
			assert.NotNil(t, comment.Article)
			assert.NotEmpty(t, comment.Article.Title)
			assert.NotEmpty(t, comment.Article.Content)
			assert.NotNil(t, comment.Article.Category)
			assert.Equal(t, category.Name, comment.Article.Category.Name)
		}

		// 5. 获取用户2的所有评论
		req5, _ := http.NewRequest("GET", "/api/v1/user/comments?page=1&page_size=10", nil)
		req5.Header.Set("Authorization", "Bearer "+token2)

		w5 := httptest.NewRecorder()
		router.ServeHTTP(w5, req5)
		assert.Equal(t, http.StatusOK, w5.Code)

		var user2CommentsResponse dto.UserCommentListResponse
		err = json.Unmarshal(w5.Body.Bytes(), &user2CommentsResponse)
		assert.NoError(t, err)

		// 验证用户2有1条评论
		assert.Equal(t, int64(1), user2CommentsResponse.Total)
		assert.Len(t, user2CommentsResponse.Comments, 1)

		// 验证回复评论包含父评论信息
		replyComment := user2CommentsResponse.Comments[0]
		assert.NotNil(t, replyComment.ParentID)
		assert.NotNil(t, replyComment.Parent)
		assert.Equal(t, comment1Response.ID, *replyComment.ParentID)
		assert.Equal(t, comment1Response.ID, replyComment.Parent.ID)
		assert.Contains(t, replyComment.Parent.Content, "这是一条很长的评论内容")
	})

	// 清理测试数据
	cleanupTestData(t)
}
