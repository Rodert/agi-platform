#!/bin/bash

# AGI Platform Backend API 测试脚本
# 使用方法: bash test_api.sh

API_URL="http://localhost:8080"
TOKEN=""
USER_EMAIL="test_$(date +%s)@example.com"
PASSWORD="password123"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 AGI Platform Backend API 测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

info() {
    echo -e "${YELLOW}→${NC} $1"
}

# 1. 健康检查
echo "1️⃣  健康检查"
response=$(curl -s "$API_URL/health")
if [[ $response == *"\"status\":\"ok\""* ]]; then
    success "服务运行正常"
else
    error "服务未运行"
    exit 1
fi
echo ""

# 2. 发送验证码
echo "2️⃣  发送验证码"
info "邮箱: $USER_EMAIL"
response=$(curl -s -X POST "$API_URL/api/v1/auth/send-code" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"type\":\"register\"}")

if [[ $response == *"\"success\":true"* ]]; then
    success "验证码发送成功"
else
    error "验证码发送失败: $response"
fi
echo ""

# 3. 用户注册（使用固定验证码123456用于测试）
echo "3️⃣  用户注册"
read -p "请输入收到的验证码（测试可输入任意6位数）: " code
if [ -z "$code" ]; then
    code="123456"
fi

response=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\":\"$USER_EMAIL\",
    \"code\":\"$code\",
    \"password\":\"$PASSWORD\",
    \"confirm_password\":\"$PASSWORD\"
  }")

if [[ $response == *"\"success\":true"* ]]; then
    TOKEN=$(echo $response | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    success "注册成功"
    info "Token: ${TOKEN:0:20}..."
else
    error "注册失败: $response"

    # 尝试登录（如果已注册）
    echo ""
    info "尝试登录..."
    response=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$PASSWORD\",\"type\":\"password\"}")

    if [[ $response == *"\"success\":true"* ]]; then
        TOKEN=$(echo $response | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        success "登录成功"
        info "Token: ${TOKEN:0:20}..."
    else
        error "登录失败"
        exit 1
    fi
fi
echo ""

# 4. 获取用户资料
echo "4️⃣  获取用户资料"
response=$(curl -s "$API_URL/api/v1/users/profile" \
  -H "Authorization: Bearer $TOKEN")

if [[ $response == *"\"success\":true"* ]]; then
    success "获取资料成功"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
else
    error "获取资料失败"
fi
echo ""

# 5. 获取 AI 模型列表
echo "5️⃣  获取 AI 模型列表"
response=$(curl -s "$API_URL/api/v1/generation/models?type=image")

if [[ $response == *"\"success\":true"* ]]; then
    success "获取模型列表成功"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
else
    error "获取模型列表失败"
fi
echo ""

# 6. 创建图片生成任务
echo "6️⃣  创建图片生成任务"
response=$(curl -s -X POST "$API_URL/api/v1/generation/image" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的橘猫，坐在窗台上看风景",
    "model_name": "GPT Image2",
    "params": {
      "ratio": "1:1",
      "resolution": "1K"
    }
  }')

if [[ $response == *"\"success\":true"* ]]; then
    TASK_ID=$(echo $response | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    success "任务创建成功，ID: $TASK_ID"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
else
    error "任务创建失败: $response"
fi
echo ""

# 7. 获取任务详情
if [ ! -z "$TASK_ID" ]; then
    echo "7️⃣  获取任务详情"
    response=$(curl -s "$API_URL/api/v1/tasks/$TASK_ID" \
      -H "Authorization: Bearer $TOKEN")

    if [[ $response == *"\"success\":true"* ]]; then
        success "获取任务详情成功"
        echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    else
        error "获取任务详情失败"
    fi
    echo ""
fi

# 8. 获取任务列表
echo "8️⃣  获取任务列表"
response=$(curl -s "$API_URL/api/v1/tasks?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")

if [[ $response == *"\"success\":true"* ]]; then
    success "获取任务列表成功"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
else
    error "获取任务列表失败"
fi
echo ""

# 9. 测试并发限制（创建多个任务）
echo "9️⃣  测试并发限制（创建3个任务）"
for i in {1..3}; do
    response=$(curl -s -X POST "$API_URL/api/v1/generation/image" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"prompt\": \"测试任务 $i\",
        \"model_name\": \"GPT Image2\",
        \"params\": {\"ratio\": \"1:1\"}
      }")

    if [[ $response == *"\"success\":true"* ]]; then
        success "任务 $i 创建成功"
    else
        if [[ $response == *"超过最大并发"* ]]; then
            info "触发并发限制（符合预期）"
        else
            error "任务 $i 创建失败: $response"
        fi
    fi
    sleep 0.5
done
echo ""

# 总结
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 测试完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 测试账号信息："
echo "  邮箱: $USER_EMAIL"
echo "  密码: $PASSWORD"
echo "  Token: ${TOKEN:0:30}..."
echo ""
echo "💡 提示："
echo "  - 查看数据库确认数据是否正确"
echo "  - 检查积分是否正确扣除"
echo "  - 任务状态应该是 'queued'"
echo ""
