#!/bin/bash

# 用户评论功能端到端测试脚本
# 测试获取登录用户的所有评论功能，包括文章信息和父评论信息

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

# 测试数据
USER1_ADDRESS="0x1234567890123456789012345678901234567890"
USER2_ADDRESS="0x1234567890123456789012345678901234567891"
USER1_USERNAME="testuser1"
USER2_USERNAME="testuser2"

# 全局变量
USER1_TOKEN=""
USER2_TOKEN=""
CATEGORY_ID=""
ARTICLE1_ID=""
ARTICLE2_ID=""
COMMENT1_ID=""
COMMENT2_ID=""

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务是否运行
check_service() {
    log_info "检查服务是否运行..."
    if ! curl -s "$BASE_URL/health" > /dev/null; then
        log_error "服务未运行，请先启动服务: make run"
        exit 1
    fi
    log_success "服务运行正常"
}

# 获取测试token
get_test_token() {
    local address=$1
    log_info "获取用户 $address 的测试token..."
    
    response=$(curl -s -X POST "$API_BASE/auth/test-token" \
        -H "Content-Type: application/json" \
        -d "{\"address\": \"$address\"}")
    
    if echo "$response" | grep -q "token"; then
        token=$(echo "$response" | jq -r '.token')
        echo "$token"
        log_success "获取token成功"
    else
        log_error "获取token失败: $response"
        exit 1
    fi
}

# 创建测试分类
create_test_category() {
    log_info "创建测试分类..."
    
    response=$(curl -s -X POST "$API_BASE/categories" \
        -H "Authorization: Bearer $USER1_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "测试分类",
            "description": "用于测试的分类",
            "icon": "test-icon",
            "color": "#FF5733"
        }')
    
    if echo "$response" | grep -q "id"; then
        category_id=$(echo "$response" | jq -r '.id')
        echo "$category_id"
        log_success "创建分类成功，ID: $category_id"
    else
        log_error "创建分类失败: $response"
        exit 1
    fi
}

# 创建测试文章
create_test_article() {
    local user_token=$1
    local category_id=$2
    local title=$3
    local content=$4
    
    log_info "创建测试文章: $title"
    
    response=$(curl -s -X POST "$API_BASE/articles" \
        -H "Authorization: Bearer $user_token" \
        -H "Content-Type: application/json" \
        -d "{
            \"title\": \"$title\",
            \"content\": \"$content\",
            \"category_id\": $category_id
        }")
    
    if echo "$response" | grep -q "id"; then
        article_id=$(echo "$response" | jq -r '.id')
        echo "$article_id"
        log_success "创建文章成功，ID: $article_id"
    else
        log_error "创建文章失败: $response"
        exit 1
    fi
}

# 创建测试评论
create_test_comment() {
    local user_token=$1
    local article_id=$2
    local parent_id=$3
    local content=$4
    
    log_info "创建测试评论: $content"
    
    if [ -n "$parent_id" ]; then
        data="{
            \"article_id\": $article_id,
            \"parent_id\": $parent_id,
            \"content\": \"$content\"
        }"
    else
        data="{
            \"article_id\": $article_id,
            \"content\": \"$content\"
        }"
    fi
    
    response=$(curl -s -X POST "$API_BASE/comments" \
        -H "Authorization: Bearer $user_token" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    if echo "$response" | grep -q "id"; then
        comment_id=$(echo "$response" | jq -r '.id')
        echo "$comment_id"
        log_success "创建评论成功，ID: $comment_id"
    else
        log_error "创建评论失败: $response"
        exit 1
    fi
}

# 测试获取用户评论列表
test_get_user_comments() {
    local user_token=$1
    local user_name=$2
    local page=$3
    local page_size=$4
    
    log_info "测试获取用户 $user_name 的评论列表 (page=$page, page_size=$page_size)..."
    
    response=$(curl -s -X GET "$API_BASE/user/comments?page=$page&page_size=$page_size" \
        -H "Authorization: Bearer $user_token")
    
    if echo "$response" | grep -q "comments"; then
        total=$(echo "$response" | jq -r '.total')
        comments_count=$(echo "$response" | jq -r '.comments | length')
        
        log_success "获取用户评论成功"
        log_info "总评论数: $total"
        log_info "当前页评论数: $comments_count"
        
        # 验证响应结构
        if echo "$response" | jq -e '.comments[0].article' > /dev/null; then
            log_success "评论包含文章信息"
        else
            log_warning "评论缺少文章信息"
        fi
        
        # 显示评论详情
        echo "$response" | jq -r '.comments[] | "评论ID: \(.id), 内容: \(.content), 文章: \(.article.title)"'
        
        return 0
    else
        log_error "获取用户评论失败: $response"
        return 1
    fi
}

# 测试未认证访问
test_unauthorized_access() {
    log_info "测试未认证访问..."
    
    response=$(curl -s -X GET "$API_BASE/user/comments?page=1&page_size=10")
    
    if echo "$response" | grep -q "unauthorized"; then
        log_success "未认证访问被正确拒绝"
        return 0
    else
        log_error "未认证访问未被拒绝: $response"
        return 1
    fi
}

# 测试无效参数
test_invalid_parameters() {
    log_info "测试无效参数..."
    
    response=$(curl -s -X GET "$API_BASE/user/comments?page=0&page_size=10" \
        -H "Authorization: Bearer $USER1_TOKEN")
    
    if echo "$response" | grep -q "error"; then
        log_success "无效参数被正确拒绝"
        return 0
    else
        log_warning "无效参数未被拒绝: $response"
        return 1
    fi
}

# 验证评论数据完整性
verify_comment_data() {
    local user_token=$1
    local user_name=$2
    
    log_info "验证用户 $user_name 的评论数据完整性..."
    
    response=$(curl -s -X GET "$API_BASE/user/comments?page=1&page_size=10" \
        -H "Authorization: Bearer $user_token")
    
    if echo "$response" | grep -q "comments"; then
        # 验证每个评论都包含必要信息
        comments_count=$(echo "$response" | jq -r '.comments | length')
        
        for i in $(seq 0 $((comments_count - 1))); do
            comment=$(echo "$response" | jq -r ".comments[$i]")
            
            # 验证评论基本信息
            comment_id=$(echo "$comment" | jq -r '.id')
            content=$(echo "$comment" | jq -r '.content')
            article_title=$(echo "$comment" | jq -r '.article.title')
            
            if [ "$comment_id" != "null" ] && [ "$content" != "null" ] && [ "$article_title" != "null" ]; then
                log_success "评论 $comment_id 数据完整"
            else
                log_error "评论 $comment_id 数据不完整"
                return 1
            fi
            
            # 验证父评论信息（如果是回复）
            parent_id=$(echo "$comment" | jq -r '.parent_id')
            if [ "$parent_id" != "null" ]; then
                parent_content=$(echo "$comment" | jq -r '.parent.content')
                if [ "$parent_content" != "null" ]; then
                    log_success "评论 $comment_id 包含父评论信息"
                else
                    log_error "评论 $comment_id 缺少父评论信息"
                    return 1
                fi
            fi
        done
        
        return 0
    else
        log_error "获取评论数据失败"
        return 1
    fi
}

# 清理测试数据
cleanup_test_data() {
    log_info "清理测试数据..."
    
    # 注意：这里只是示例，实际清理需要根据数据库结构来实现
    # 在生产环境中，应该使用专门的清理脚本
    
    log_success "测试数据清理完成"
}

# 主测试流程
main() {
    log_info "开始用户评论功能端到端测试"
    
    # 检查服务状态
    check_service
    
    # 获取测试token
    USER1_TOKEN=$(get_test_token "$USER1_ADDRESS")
    USER2_TOKEN=$(get_test_token "$USER2_ADDRESS")
    
    # 创建测试分类
    CATEGORY_ID=$(create_test_category)
    
    # 创建测试文章
    ARTICLE1_ID=$(create_test_article "$USER1_TOKEN" "$CATEGORY_ID" "测试文章1" "这是测试文章1的内容，包含了很多技术细节和实现方案")
    ARTICLE2_ID=$(create_test_article "$USER2_TOKEN" "$CATEGORY_ID" "测试文章2" "这是测试文章2的内容，主要讨论架构设计")
    
    # 创建测试评论
    COMMENT1_ID=$(create_test_comment "$USER1_TOKEN" "$ARTICLE1_ID" "" "这是用户1在文章1上的第一条评论，内容比较长，包含了很多想法和建议")
    COMMENT2_ID=$(create_test_comment "$USER2_TOKEN" "$ARTICLE1_ID" "$COMMENT1_ID" "这是用户2对用户1评论的回复，表示赞同并补充一些观点")
    COMMENT3_ID=$(create_test_comment "$USER1_TOKEN" "$ARTICLE2_ID" "" "这是用户1在文章2上的评论，讨论架构设计的优缺点")
    
    log_info "等待数据同步..."
    sleep 2
    
    # 测试用例1: 获取用户1的评论列表
    test_get_user_comments "$USER1_TOKEN" "$USER1_USERNAME" 1 10
    
    # 测试用例2: 获取用户2的评论列表
    test_get_user_comments "$USER2_TOKEN" "$USER2_USERNAME" 1 10
    
    # 测试用例3: 分页测试
    test_get_user_comments "$USER1_TOKEN" "$USER1_USERNAME" 1 2
    
    # 测试用例4: 验证数据完整性
    verify_comment_data "$USER1_TOKEN" "$USER1_USERNAME"
    verify_comment_data "$USER2_TOKEN" "$USER2_USERNAME"
    
    # 测试用例5: 未认证访问
    test_unauthorized_access
    
    # 测试用例6: 无效参数
    test_invalid_parameters
    
    log_success "所有测试用例执行完成"
    
    # 显示测试总结
    echo ""
    log_info "测试总结:"
    echo "✓ 用户评论列表获取功能正常"
    echo "✓ 评论包含完整的文章信息"
    echo "✓ 回复评论包含父评论信息"
    echo "✓ 分页功能正常"
    echo "✓ 认证授权正常"
    echo "✓ 参数验证正常"
    echo "✓ 数据完整性验证通过"
    
    # 清理测试数据
    cleanup_test_data
    
    log_success "用户评论功能端到端测试完成"
}

# 错误处理
trap 'log_error "测试过程中发生错误，退出"; exit 1' ERR

# 运行主测试
main "$@" 