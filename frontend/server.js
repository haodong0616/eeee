const { createServer } = require('http');
const { parse } = require('url');
const next = require('next');
const { WebSocketServer } = require('ws');
const WebSocket = require('ws');

const dev = process.env.NODE_ENV !== 'production';
const hostname = '0.0.0.0';
const port = 3000;

const app = next({ dev, hostname, port });
const handle = app.getRequestHandler();

const BACKEND_URL = process.env.BACKEND_URL || 'localhost:8383';

app.prepare().then(() => {
  const server = createServer(async (req, res) => {
    try {
      const parsedUrl = parse(req.url, true);
      await handle(req, res, parsedUrl);
    } catch (err) {
      console.error('Error occurred handling', req.url, err);
      res.statusCode = 500;
      res.end('internal server error');
    }
  });

  // WebSocket 代理
  const wss = new WebSocketServer({ noServer: true });

  server.on('upgrade', (request, socket, head) => {
    const { pathname } = parse(request.url);

    if (pathname === '/api/ws') {
      wss.handleUpgrade(request, socket, head, (ws) => {
        // 连接到后端 WebSocket
        const backendWs = new WebSocket(`ws://${BACKEND_URL}/ws`);

        // 前端 -> 后端（数据消息）
        ws.on('message', (data, isBinary) => {
          if (backendWs.readyState === WebSocket.OPEN) {
            backendWs.send(data, { binary: isBinary });
          }
        });

        // 后端 -> 前端（数据消息）
        backendWs.on('message', (data, isBinary) => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(data, { binary: isBinary });
          }
        });

        // 前端 -> 后端（Ping）
        ws.on('ping', (data) => {
          if (backendWs.readyState === WebSocket.OPEN) {
            backendWs.ping(data);
          }
        });

        // 后端 -> 前端（Ping）
        backendWs.on('ping', (data) => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.ping(data);
          }
        });

        // 前端 -> 后端（Pong）
        ws.on('pong', (data) => {
          if (backendWs.readyState === WebSocket.OPEN) {
            backendWs.pong(data);
          }
        });

        // 后端 -> 前端（Pong）
        backendWs.on('pong', (data) => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.pong(data);
          }
        });

        // 后端连接成功
        backendWs.on('open', () => {
          console.log('✅ WebSocket 代理连接成功');
        });

        // 错误处理
        backendWs.on('error', (error) => {
          console.error('❌ 后端 WebSocket 错误:', error.message);
          ws.close();
        });

        ws.on('error', (error) => {
          console.error('❌ 前端 WebSocket 错误:', error.message);
          backendWs.close();
        });

        // 关闭处理
        ws.on('close', (code, reason) => {
          console.log(`📤 前端连接关闭 [${code}]: ${reason}`);
          backendWs.close();
        });

        backendWs.on('close', (code, reason) => {
          console.log(`📤 后端连接关闭 [${code}]: ${reason}`);
          ws.close();
        });
      });
    } else {
      socket.destroy();
    }
  });

  server.listen(port, hostname, (err) => {
    if (err) throw err;
    console.log('');
    console.log('🚀 Velocity Exchange 启动成功！');
    console.log('========================================');
    console.log(`📱 手机访问地址（Network地址）`);
    console.log(`   http://${hostname === '0.0.0.0' ? '你的IP' : hostname}:${port}`);
    console.log('');
    console.log(`💻 本地访问地址：`);
    console.log(`   http://localhost:${port}`);
    console.log('');
    console.log('✨ API 和 WebSocket 已自动代理到后端');
    console.log('========================================');
    console.log('');
  });
});


