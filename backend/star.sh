#!/bin/bash
# =========================================
# startapp 自动守护启动脚本
# =========================================
# 设置环境变量
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3308
export DB_USER=referral_user
export DB_PASSWORD=referral123456
export DB_NAME=referral_system

MYAPP="./exchange_main"
STARTAPP="./exchange_start"
LOG_OUT="./log/exchange_access.log"
LOG_ERR="./log/exchange_error.log"

# ------------------ 1️⃣ 杀掉正在运行的 exchange_start ------------------
if [ -f "$STARTAPP" ]; then
    PID=$(pgrep -f "$STARTAPP")
    if [ -n "$PID" ]; then
        echo "⚠️ exchange_start 正在运行，PID: $PID，正在杀掉旧进程..."
        kill -9 $PID
        sleep 1
    fi
else
    echo "ℹ️ exchange_start 文件不存在，跳过杀进程"
fi

# ------------------ 2️⃣ 如果 exchange_main 存在，替换 exchange_start ------------------
if [ -f "$MYAPP" ]; then
    echo "🔄 exchange_main 文件存在，更新 exchange_start..."

    # 删除旧的 bot_backend_start 文件
    if [ -f "$STARTAPP" ]; then
        echo "🗑️ 删除旧的 exchange_start 文件..."
        rm -f "$STARTAPP"
    fi

    # 重命名 exchange_main 为 exchange_start
    mv "$MYAPP" "$STARTAPP"
else
    echo "ℹ️ exchange_main 文件不存在，只是重新启动 exchange_start"
fi

# ------------------ 3️⃣ 检查 exchange_start 是否存在 ------------------
if [ ! -f "$STARTAPP" ]; then
    echo "❌ 错误: exchange_start 文件不存在，无法启动"
    exit 1
fi

# ------------------ 4️⃣ 启动 exchange_start 并守护 ------------------
echo "🚀 启动 exchange_start 并守护进程..."
nohup "$STARTAPP" > "$LOG_OUT" 2> "$LOG_ERR" &

NEW_PID=$!
echo "✅ exchange_start 已启动，PID: $NEW_PID"
echo "日志输出到: $LOG_OUT"

