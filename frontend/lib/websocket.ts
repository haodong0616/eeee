// WebSocket 配置 - 自动检测环境
const getWsUrl = () => {
  // 服务器端直接返回默认值
  if (typeof window === 'undefined') {
    return 'ws://localhost:8383';
  }
  
  if (process.env.NEXT_PUBLIC_WS_URL) {
    return process.env.NEXT_PUBLIC_WS_URL;
  }
  
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const hostname = window.location.hostname;
  const port = window.location.port;
  
  // ⚠️ 生产环境（通过Nginx，标准端口）：wss://velocity.0v1.xyz
  // ⚠️ 开发环境（Next.js服务器，非标准端口）：ws://localhost:3000
  if (port) {
    return `${protocol}//${hostname}:${port}`;
  } else {
    return `${protocol}//${hostname}`;
  }
};

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectInterval: number = 5000;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private handlers: Map<string, ((data: any) => void)[]> = new Map();

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    const WS_URL = getWsUrl();
    this.ws = new WebSocket(`${WS_URL}/ws`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        
        // 调试日志：显示收到的消息
        console.log(`📩 WebSocket收到消息 [${message.type}]:`, message.data);
        
        const handlers = this.handlers.get(message.type);
        if (handlers) {
          console.log(`✅ 执行${handlers.length}个处理器`);
          handlers.forEach((handler) => handler(message.data));
        } else {
          console.log(`⚠️ 没有${message.type}类型的处理器`);
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.reconnect();
    };
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private reconnect() {
    if (this.reconnectTimer) {
      return;
    }

    this.reconnectTimer = setTimeout(() => {
      console.log('Reconnecting WebSocket...');
      this.connect();
    }, this.reconnectInterval);
  }

  on(type: string, handler: (data: any) => void) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, []);
    }
    this.handlers.get(type)!.push(handler);
  }

  off(type: string, handler: (data: any) => void) {
    const handlers = this.handlers.get(type);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index !== -1) {
        handlers.splice(index, 1);
      }
    }
  }

  send(message: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }
}

// 延迟创建 WebSocket 客户端实例，避免服务器端渲染时出错
let _wsClient: WebSocketClient | null = null;

// 导出一个getter，完全避免在模块加载时创建实例
export const wsClient = new Proxy({} as WebSocketClient, {
  get(target, prop) {
    if (typeof window === 'undefined') {
      // 服务器端返回空函数
      return () => {};
    }
    
    if (!_wsClient) {
      _wsClient = new WebSocketClient();
    }
    
    const value = (_wsClient as any)[prop];
    return typeof value === 'function' ? value.bind(_wsClient) : value;
  }
});

