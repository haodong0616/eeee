#!/bin/bash
# =========================================
# 自动化部署脚本：安装 Docker + 拉取前端代码 + Docker 运行
# 仓库: git@github.com/web3-mutual-aid/front-end.git
# 端口: 3000
# =========================================



APP_NAME="web3-expchange-frontend"
GIT_REPO="git@github.com:haodong0616/eeee.git"
APP_DIR="/opt/${APP_NAME}"
FRONTEND_DIR="frontend"  # 前端代码所在子目录
PORT=3005

echo "2️⃣ 检查并安装 Docker..."
if ! command -v docker &> /dev/null; then
  curl -fsSL https://get.docker.com | sh
  sudo systemctl enable docker
  sudo systemctl start docker
else
  echo "✅ Docker 已安装"
fi

echo "3️⃣ 拉取项目代码..."
if [ -d "${APP_DIR}" ]; then
  cd "${APP_DIR}" && git reset --hard && git pull
else
  sudo git clone "${GIT_REPO}" "${APP_DIR}"
  sudo chown -R $USER:$USER "${APP_DIR}"
fi

echo "4️⃣ 进入前端目录..."
cd "${APP_DIR}/${FRONTEND_DIR}"

echo "5️⃣ 创建 Dockerfile..."
cat > Dockerfile <<EOF
FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN yarn install --frozen-lockfile || npm install

COPY . .

RUN yarn build || npm run build

EXPOSE ${PORT}

CMD ["yarn", "start"]
EOF

echo "6️⃣ 停止并删除旧容器（如果有）..."
sudo docker rm -f ${APP_NAME} 2>/dev/null || true

echo "7️⃣ 删除旧镜像（如果有）..."
OLD_IMAGE_ID=$(sudo docker images -q ${APP_NAME}:latest)
if [ -n "$OLD_IMAGE_ID" ]; then
  echo "🗑️  删除旧镜像 ${APP_NAME}:latest ..."
  sudo docker rmi -f "$OLD_IMAGE_ID"
fi

echo "8️⃣ 构建 Docker 镜像..."
sudo docker build -t ${APP_NAME}:latest .

echo "9️⃣ 运行新容器..."
sudo docker run -d \
  --name ${APP_NAME} \
  -p ${PORT}:3000 \
  ${APP_NAME}:latest

echo "✅ 部署完成，访问: http://服务器IP:${PORT}"
