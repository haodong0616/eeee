# 管理后台过滤模拟器数据

## 🎯 功能说明

管理后台的订单和交易列表现在会**自动过滤掉系统模拟器**产生的数据，只显示真实用户的订单和交易。

## 🤖 模拟器识别

### 虚拟用户地址

系统模拟器使用固定的虚拟用户：

```
钱包地址: 0x0000000000000000000000000000000000000000
用途: 为订单簿提供流动性（挂买卖单）
```

## 📝 修改内容

### 1. 订单列表过滤

**文件**: `backend/handlers/admin.go`

**修改前**:
```go
func (h *AdminHandler) GetAllOrders(c *gin.Context) {
    var orders []models.Order
    database.DB.Preload("User").Order("created_at DESC").Limit(500).Find(&orders)
    c.JSON(http.StatusOK, orders)
}
```

**修改后**:
```go
func (h *AdminHandler) GetAllOrders(c *gin.Context) {
    var orders []models.Order
    
    // 排除虚拟用户的订单
    database.DB.
        Preload("User").
        Where("user_id NOT IN (?)", 
            database.DB.Table("users").
                Select("id").
                Where("wallet_address = ?", "0x0000000000000000000000000000000000000000")).
        Order("created_at DESC").
        Limit(500).
        Find(&orders)
    
    c.JSON(http.StatusOK, orders)
}
```

**SQL 示例**:
```sql
SELECT * FROM orders 
WHERE user_id NOT IN (
    SELECT id FROM users 
    WHERE wallet_address = '0x0000000000000000000000000000000000000000'
)
ORDER BY created_at DESC 
LIMIT 500;
```

### 2. 交易列表过滤

**修改前**:
```go
func (h *AdminHandler) GetAllTrades(c *gin.Context) {
    var trades []models.Trade
    database.DB.Order("created_at DESC").Limit(500).Find(&trades)
    c.JSON(http.StatusOK, trades)
}
```

**修改后**:
```go
func (h *AdminHandler) GetAllTrades(c *gin.Context) {
    var trades []models.Trade
    
    // 排除虚拟用户参与的交易（买单或卖单）
    database.DB.
        Where("buy_order_id NOT IN (?) AND sell_order_id NOT IN (?)",
            database.DB.Table("orders").
                Joins("JOIN users ON users.id = orders.user_id").
                Where("users.wallet_address = ?", "0x0000000000000000000000000000000000000000").
                Select("orders.id"),
            database.DB.Table("orders").
                Joins("JOIN users ON users.id = orders.user_id").
                Where("users.wallet_address = ?", "0x0000000000000000000000000000000000000000").
                Select("orders.id")).
        Order("created_at DESC").
        Limit(500).
        Find(&trades)
    
    c.JSON(http.StatusOK, trades)
}
```

**SQL 示例**:
```sql
SELECT * FROM trades 
WHERE buy_order_id NOT IN (
    SELECT orders.id FROM orders
    JOIN users ON users.id = orders.user_id
    WHERE users.wallet_address = '0x0000000000000000000000000000000000000000'
)
AND sell_order_id NOT IN (
    SELECT orders.id FROM orders
    JOIN users ON users.id = orders.user_id
    WHERE users.wallet_address = '0x0000000000000000000000000000000000000000'
)
ORDER BY created_at DESC 
LIMIT 500;
```

## 🎯 过滤规则

### 订单过滤
- ✅ 排除：user_id 是虚拟用户的订单
- ✅ 显示：所有真实用户的订单

### 交易过滤
- ✅ 排除：买方或卖方任一是虚拟用户的交易
- ✅ 显示：买卖双方都是真实用户的交易

## 📊 数据统计

### 统计信息不受影响

**文件**: `backend/handlers/admin.go - GetStats()`

统计数据（如总订单数、总交易量）**仍包含模拟器数据**，因为：
- 模拟器订单提供市场流动性
- 模拟器交易反映真实的市场撮合
- 统计应该反映整个系统的活跃度

如果需要单独统计真实用户数据，可以添加：

```go
// 真实用户订单数
realOrderCount := 0
database.DB.Model(&models.Order{}).
    Where("user_id NOT IN (?)", 
        database.DB.Table("users").Select("id").Where("wallet_address = ?", "0x0000...")).
    Count(&realOrderCount)

// 模拟器订单数
simOrderCount := 0
database.DB.Model(&models.Order{}).
    Where("user_id IN (?)", 
        database.DB.Table("users").Select("id").Where("wallet_address = ?", "0x0000...")).
    Count(&simOrderCount)
```

## ✅ 验证修复

### 1. 重启后端

```bash
cd /root/go
./star.sh
```

### 2. 访问管理后台

```
http://localhost:3001/dashboard/orders
http://localhost:3001/dashboard/trades
```

### 3. 确认效果

**订单页面**:
- ✅ 只显示真实用户的订单
- ✅ 不显示虚拟用户的挂单

**交易页面**:
- ✅ 只显示真实用户之间的交易
- ✅ 不显示涉及虚拟用户的交易

## 💡 其他可选过滤

如果需要在前端也能查看模拟器数据，可以添加过滤选项：

```typescript
// 添加过滤器
const [showSimulator, setShowSimulator] = useState(false);

// 在页面添加切换按钮
<label>
  <input 
    type="checkbox" 
    checked={showSimulator}
    onChange={e => setShowSimulator(e.target.checked)}
  />
  显示模拟器数据
</label>

// 后端添加查询参数
GET /admin/orders?include_simulator=true
```

## 🔍 调试

如果过滤不生效，检查：

```sql
-- 1. 确认虚拟用户存在
SELECT * FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000';

-- 2. 查看虚拟用户的订单数量
SELECT COUNT(*) FROM orders WHERE user_id = 'VIRTUAL_USER_ID';

-- 3. 查看虚拟用户的交易数量
SELECT COUNT(*) FROM trades 
WHERE buy_order_id IN (SELECT id FROM orders WHERE user_id = 'VIRTUAL_USER_ID')
   OR sell_order_id IN (SELECT id FROM orders WHERE user_id = 'VIRTUAL_USER_ID');
```

## ✨ 总结

✅ **订单列表** - 只显示真实用户订单
✅ **交易列表** - 只显示真实交易
✅ **性能优化** - 使用子查询，效率高
✅ **灵活扩展** - 可轻松添加过滤开关

现在管理后台更清晰，只关注真实用户的交易活动！🎉

