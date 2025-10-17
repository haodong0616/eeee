-- =========================================
-- 修复钱包地址大小写不一致问题
-- =========================================
-- 
-- 问题：不同设备登录时，钱包可能返回不同格式的地址
--   设备 A: 0xAbCdEf... (checksummed)
--   设备 B: 0xabcdef... (lowercase)
-- 
-- 解决：统一转换为小写
-- =========================================

-- 1. 查看当前有哪些地址（检查是否有大写）
SELECT 
    id,
    wallet_address,
    LOWER(wallet_address) as lowercase_address,
    CASE 
        WHEN wallet_address = LOWER(wallet_address) THEN '✅ 已是小写'
        ELSE '❌ 包含大写'
    END as status
FROM users
ORDER BY created_at DESC;

-- 2. 修复：将所有钱包地址转为小写
UPDATE users 
SET wallet_address = LOWER(wallet_address)
WHERE wallet_address != LOWER(wallet_address);

-- 3. 验证修复结果
SELECT 
    COUNT(*) as total_users,
    SUM(CASE WHEN wallet_address = LOWER(wallet_address) THEN 1 ELSE 0 END) as lowercase_count
FROM users;

-- 4. 显示所有用户（确认）
SELECT 
    id,
    wallet_address,
    user_level,
    created_at
FROM users
ORDER BY created_at DESC
LIMIT 10;

-- =========================================
-- 注意事项
-- =========================================
-- 
-- ✅ 已在代码中添加保护：
--    - User.BeforeCreate(): 创建前强制转小写
--    - User.BeforeSave(): 保存前强制转小写
--    - auth.GetNonce(): 查询前转小写
--    - auth.Login(): 查询前转小写
-- 
-- ✅ 前端登录时也已转小写：
--    - Header.tsx: handleAutoLogin() 中使用 toLowerCase()
-- 
-- ⚠️ 如果仍有问题，检查：
--    1. 数据库中是否有重复的地址（大小写不同）
--    2. 前端缓存的 token 是否对应错误的用户
--    3. localStorage 中的数据是否需要清理
-- 
-- =========================================

