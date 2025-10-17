# 钱包地址大小写问题修复

## 🐛 问题描述

**症状**：两个设备登录，另一个设备看不到数据

**原因**：钱包地址大小写不一致导致被识别为不同用户

## 🔍 问题根源

### Ethereum 地址格式

以太坊地址有两种格式：

1. **小写格式** (lowercase):
   ```
   0xabcdef1234567890abcdef1234567890abcdef12
   ```

2. **校验和格式** (checksummed - EIP-55):
   ```
   0xAbCdEf1234567890AbCdEf1234567890AbCdEf12
   ```

**问题**：
- 设备 A：MetaMask 返回 `0xAbCd...`（大小写混合）
- 设备 B：Trust Wallet 返回 `0xabcd...`（全小写）
- 数据库查询时大小写敏感 → 被当作两个不同的用户！

## ✅ 解决方案

### 1. 后端强制转小写（多重保护）

#### A. 模型钩子（最底层保护）

**文件**：`backend/models/models.go`

```go
import "strings"

func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = utils.GenerateObjectID()
    }
    // 🔒 强制转换为小写
    u.WalletAddress = strings.ToLower(u.WalletAddress)
    return nil
}

func (u *User) BeforeSave(tx *gorm.DB) error {
    // 🔒 强制转换为小写
    u.WalletAddress = strings.ToLower(u.WalletAddress)
    return nil
}
```

**作用**：无论从哪里创建/更新用户，都会自动转小写。

#### B. API 层转换（第二层保护）

**文件**：`backend/handlers/auth.go`

```go
// GetNonce
walletAddress := strings.ToLower(req.WalletAddress)  // ✅ 已实现

// Login
walletAddress := strings.ToLower(req.WalletAddress)  // ✅ 已实现
```

**作用**：API 接收时就转小写，确保查询正确。

### 2. 前端转小写

**文件**：`frontend/components/Header.tsx`

```typescript
const handleAutoLogin = async (walletAddress: string) => {
  const lowerAddress = walletAddress.toLowerCase();  // ✅ 已实现
  
  await getNonce(lowerAddress);
  await login({ walletAddress: lowerAddress, signature });
}
```

**作用**：确保发送到后端的地址是小写。

### 3. 修复现有数据

**文件**：`backend/fix_wallet_address_case.sql`

执行 SQL 修复已存在的大写地址：

```sql
-- 将所有钱包地址转为小写
UPDATE users 
SET wallet_address = LOWER(wallet_address)
WHERE wallet_address != LOWER(wallet_address);
```

**执行方法**：

```bash
# MySQL
mysql -u referral_user -p referral_system < backend/fix_wallet_address_case.sql

# 或者在 MySQL 命令行中
USE expchange;
UPDATE users SET wallet_address = LOWER(wallet_address);
```

## 🔧 完整修复步骤

### 步骤 1：清理数据库

```bash
# SSH 到服务器
ssh root@e11e

# 连接到 MySQL
mysql -u referral_user -p

# 切换数据库
USE expchange;

# 修复地址
UPDATE users SET wallet_address = LOWER(wallet_address);

# 验证
SELECT wallet_address FROM users;

# 退出
EXIT;
```

### 步骤 2：重启后端

```bash
cd /root/go
./star.sh
```

新代码会应用 `BeforeCreate` 和 `BeforeSave` 钩子。

### 步骤 3：清理前端缓存

在两个设备上都执行：
```javascript
// 打开浏览器控制台
localStorage.clear();
// 刷新页面
location.reload();
```

### 步骤 4：重新登录

两个设备重新连接钱包登录，现在应该能看到相同的数据了！

## 🧪 验证修复

### 检查数据库

```sql
-- 检查是否还有大写地址
SELECT 
    wallet_address,
    CASE 
        WHEN wallet_address = LOWER(wallet_address) THEN '✅'
        ELSE '❌'
    END as is_lowercase
FROM users;

-- 应该全部显示 ✅
```

### 检查日志

```bash
# 查看后端日志
tail -f /root/go/log/exchange_access.log

# 登录时应该看到
🔐 开始登录流程，地址: 0xabcdef...（全小写）
```

## 📋 已修改的文件

1. ✅ `backend/models/models.go`
   - 添加 `BeforeCreate` 钩子转小写
   - 添加 `BeforeSave` 钩子转小写
   - 添加 `strings` 导入

2. ✅ `backend/handlers/auth.go`
   - `GetNonce` 中转小写（已存在）
   - `Login` 中转小写（已存在）

3. ✅ `frontend/components/Header.tsx`
   - `handleAutoLogin` 中转小写（已存在）

4. ✅ `backend/fix_wallet_address_case.sql`
   - SQL 脚本修复现有数据

## 🛡️ 防护层级

现在有 **4 层保护** 确保地址始终是小写：

```
第1层：前端转换
  └─> walletAddress.toLowerCase()

第2层：API 层转换
  └─> strings.ToLower(req.WalletAddress)

第3层：模型创建钩子
  └─> BeforeCreate() { u.WalletAddress = strings.ToLower(...) }

第4层：模型保存钩子
  └─> BeforeSave() { u.WalletAddress = strings.ToLower(...) }
```

## ✨ 总结

**问题**：钱包返回地址格式不一致（checksummed vs lowercase）

**解决**：
1. ✅ 后端模型钩子强制转小写（万无一失）
2. ✅ API 层转小写（已实现）
3. ✅ 前端登录转小写（已实现）
4. ✅ SQL 脚本修复现有数据

**下一步**：
1. 在服务器上执行 SQL 修复脚本
2. 重启后端应用
3. 清理两个设备的浏览器缓存
4. 重新登录测试

现在应该能解决跨设备数据不一致的问题了！🎉

