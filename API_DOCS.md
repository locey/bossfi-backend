# BossFi Backend API 文档

## 获取用户资料接口

### 接口信息
- **接口名称**: 获取用户资料
- **接口描述**: 获取当前登录用户的详细资料信息
- **接口地址**: `GET /api/v1/users/profile`
- **需要认证**: 是（Bearer Token）

### 请求参数

#### Headers
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} 格式的JWT令牌 |
| Content-Type | string | 否 | application/json |

#### Query Parameters
无

#### Request Body
无

### 响应数据

#### 成功响应 (200)
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "wallet_address": "0x1234567890123456789012345678901234567890",
    "username": "boss_hunter",
    "email": "user@example.com",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "区块链开发者，专注于DeFi和NFT项目",
    "boss_balance": "1000.500000000000000000",
    "staked_amount": "500.000000000000000000",
    "reward_balance": "25.750000000000000000",
    "is_profile_complete": true,
    "last_login_at": "2025-06-23T12:30:00Z",
    "created_at": "2025-01-01T10:00:00Z",
    "updated_at": "2025-06-23T12:30:00Z"
  }
}
```

#### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | string | 用户唯一标识符（UUID） |
| wallet_address | string | 用户钱包地址（以太坊地址格式） |
| username | string | 用户名（可为空） |
| email | string | 邮箱地址（可为空） |
| avatar | string | 头像URL（可为空） |
| bio | string | 个人简介（可为空） |
| boss_balance | string | BOSS币余额（decimal格式，18位小数） |
| staked_amount | string | 质押金额（decimal格式，18位小数） |
| reward_balance | string | 奖励余额（decimal格式，18位小数） |
| is_profile_complete | boolean | 资料是否完善 |
| last_login_at | string | 最后登录时间（ISO 8601格式，可为空） |
| created_at | string | 账户创建时间（ISO 8601格式） |
| updated_at | string | 最后更新时间（ISO 8601格式） |

#### 错误响应

##### 401 Unauthorized - 未授权
```json
{
  "code": 401,
  "message": "Unauthorized",
  "error": "missing or invalid token"
}
```

##### 404 Not Found - 用户不存在
```json
{
  "code": 404,
  "message": "User not found",
  "error": "user not found in database"
}
```

##### 500 Internal Server Error - 服务器错误
```json
{
  "code": 500,
  "message": "Internal server error",
  "error": "database connection failed"
}
```

### 测试用例

#### 使用 curl 测试
```bash
# 获取用户资料
curl -X GET "http://localhost:8080/api/v1/users/profile" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

#### 使用 JavaScript 测试
```javascript
// 获取用户资料
const getUserProfile = async () => {
  try {
    const response = await fetch('http://localhost:8080/api/v1/users/profile', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    if (data.code === 0) {
      console.log('用户资料:', data.data);
      return data.data;
    } else {
      console.error('获取失败:', data.message);
      throw new Error(data.message);
    }
  } catch (error) {
    console.error('请求错误:', error);
    throw error;
  }
};

// 使用示例
getUserProfile()
  .then(profile => {
    console.log('用户ID:', profile.id);
    console.log('钱包地址:', profile.wallet_address);
    console.log('BOSS余额:', profile.boss_balance);
  })
  .catch(error => {
    console.error('获取用户资料失败:', error);
  });
```

#### 使用 Python 测试
```python
import requests
import json

def get_user_profile(token):
    """获取用户资料"""
    url = "http://localhost:8080/api/v1/users/profile"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    try:
        response = requests.get(url, headers=headers)
        data = response.json()
        
        if data["code"] == 0:
            print("获取成功:", json.dumps(data["data"], indent=2, ensure_ascii=False))
            return data["data"]
        else:
            print(f"获取失败: {data['message']}")
            return None
            
    except requests.exceptions.RequestException as e:
        print(f"请求错误: {e}")
        return None

# 使用示例
token = "your_jwt_token_here"
profile = get_user_profile(token)

if profile:
    print(f"用户ID: {profile['id']}")
    print(f"钱包地址: {profile['wallet_address']}")
    print(f"BOSS余额: {profile['boss_balance']}")
```

### 前端集成建议

#### 1. 状态管理
```javascript
// 使用 Redux/Zustand 等状态管理
const useUserStore = create((set) => ({
  profile: null,
  loading: false,
  error: null,
  
  fetchProfile: async () => {
    set({ loading: true, error: null });
    try {
      const profile = await getUserProfile();
      set({ profile, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },
  
  clearProfile: () => set({ profile: null, error: null })
}));
```

#### 2. React Hook 示例
```javascript
import { useState, useEffect } from 'react';

const useUserProfile = () => {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    const fetchProfile = async () => {
      try {
        setLoading(true);
        const data = await getUserProfile();
        setProfile(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    
    fetchProfile();
  }, []);
  
  return { profile, loading, error, refetch: fetchProfile };
};
```

#### 3. 错误处理
```javascript
const handleApiError = (error, response) => {
  switch (response?.status) {
    case 401:
      // 清除本地token，跳转登录页
      localStorage.removeItem('token');
      window.location.href = '/login';
      break;
    case 404:
      // 用户不存在，可能需要重新注册
      console.error('用户不存在');
      break;
    case 500:
      // 服务器错误，显示错误提示
      console.error('服务器错误，请稍后重试');
      break;
    default:
      console.error('未知错误:', error);
  }
};
```

### 注意事项

1. **认证要求**: 此接口需要有效的JWT令牌，请确保在请求头中包含正确的Authorization信息
2. **余额格式**: 所有余额字段都是decimal格式的字符串，前端处理时需要注意精度
3. **可选字段**: username、email、avatar、bio、last_login_at 字段可能为null，前端需要做好空值处理
4. **时间格式**: 时间字段采用ISO 8601格式，前端可以使用Date对象或moment.js等库进行处理
5. **缓存策略**: 建议对用户资料进行适当缓存，避免频繁请求

### 相关接口

- `PUT /api/v1/users/profile` - 更新用户资料
- `GET /api/v1/users/stats` - 获取用户统计信息  
- `GET /api/v1/users/balance` - 获取用户余额详情
- `POST /api/v1/auth/login` - 用户登录（获取token） 