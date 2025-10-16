package main

import (
	"expchange-backend/config"
	"expchange-backend/database"
	"expchange-backend/handlers"
	"expchange-backend/kline"
	"expchange-backend/matching"
	"expchange-backend/middleware"
	"expchange-backend/queue"
	"expchange-backend/simulator"
	"expchange-backend/websocket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库
	if err := database.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 自动初始化数据（首次启动时）
	database.AutoSeed()

	// 初始化Redis（可选，目前未使用）
	if err := database.InitRedis(cfg); err != nil {
		log.Printf("Warning: Redis not available: %v", err)
		log.Println("Continuing without Redis (caching features will be disabled)")
	}

	// 初始化撮合引擎
	matchingManager := matching.NewManager()

	// 初始化WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 初始化K线生成器
	klineGenerator := kline.NewGenerator()
	klineGenerator.Start()

	// 初始化系统配置管理器
	sysConfig := database.GetSystemConfigManager()

	// 初始化任务队列（统一管理所有后台任务）
	// 任务队列内部会初始化：
	// - 数据生成workers（多个并发worker）
	// - 充值验证worker（独立进程）
	// - 提现处理worker（独立进程 + Nonce管理器）
	queue.GetQueue()
	workers := sysConfig.GetInt("task.queue.workers", 10)
	log.Printf("✅ 任务队列已初始化")
	log.Printf("   - 数据生成: %d个并发worker", workers)
	log.Printf("   - 充值验证: 1个独立worker")
	log.Printf("   - 提现处理: 1个独立worker（线程安全）")

	// 启动动态订单簿模拟器（根据数据库配置自动为启用的交易对生成订单）
	dynamicSim := simulator.NewDynamicOrderBookSimulator(matchingManager)
	dynamicSim.Start()
	log.Println("✅ 动态订单簿模拟器已启动")

	// 初始化Gin (设置为Release模式降低日志输出)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 只添加Recovery中间件，不添加Logger中间件以减少日志输出
	r.Use(gin.Recovery())

	// 中间件
	r.Use(middleware.CORSMiddleware(cfg))

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(cfg)
	marketHandler := handlers.NewMarketHandler(matchingManager)
	orderHandler := handlers.NewOrderHandler(matchingManager)
	balanceHandler := handlers.NewBalanceHandler()
	wsHandler := handlers.NewWebSocketHandler(wsHub)
	adminHandler := handlers.NewAdminHandler()
	klineHandler := handlers.NewKlineHandler(klineGenerator)
	feeHandler := handlers.NewFeeHandler()
	chainHandler := handlers.NewChainHandler()

	// API路由
	api := r.Group("/api")
	{
		// 公开路由
		auth := api.Group("/auth")
		{
			auth.POST("/nonce", authHandler.GetNonce)
			auth.POST("/login", authHandler.Login)
		}

		// 市场数据（公开）
		market := api.Group("/market")
		{
			market.GET("/pairs", marketHandler.GetTradingPairs)
			market.GET("/ticker/:symbol", marketHandler.GetTicker)
			market.GET("/tickers", marketHandler.GetAllTickers)
			market.GET("/orderbook/:symbol", marketHandler.GetOrderBook)
			market.GET("/trades/:symbol", marketHandler.GetRecentTrades)
			market.GET("/klines/:symbol", klineHandler.GetKlines)
			market.GET("/klines/:symbol/tv", klineHandler.GetKlinesForTradingView)
		}

		// 链配置（公开，只返回启用的链）
		api.GET("/chains", chainHandler.GetEnabledChains)

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(cfg))
		{
			// 用户信息
			authenticated.GET("/profile", authHandler.GetProfile)

			// 订单
			orders := authenticated.Group("/orders")
			{
				orders.POST("", orderHandler.CreateOrder)
				orders.GET("", orderHandler.GetOrders)
				orders.GET("/:id", orderHandler.GetOrder)
				orders.DELETE("/:id", orderHandler.CancelOrder)
			}

			// 余额
			balances := authenticated.Group("/balances")
			{
				balances.GET("", balanceHandler.GetBalances)
				balances.GET("/:asset", balanceHandler.GetBalance)
				balances.POST("/deposit", balanceHandler.Deposit)
				balances.POST("/withdraw", balanceHandler.Withdraw)
				balances.GET("/deposits", balanceHandler.GetDepositRecords)
				balances.GET("/withdraws", balanceHandler.GetWithdrawRecords)
			}

			// 手续费
			fees := authenticated.Group("/fees")
			{
				fees.GET("/stats", feeHandler.GetUserFeeStats)
				fees.GET("/records", feeHandler.GetUserFeeRecords)
			}
		}

		// 管理后台路由
		admin := api.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware(cfg))
		{
			admin.GET("/users", adminHandler.GetUsers)
			admin.GET("/orders", adminHandler.GetAllOrders)
			admin.GET("/trades", adminHandler.GetAllTrades)
			admin.GET("/deposits", adminHandler.GetAllDeposits)
			admin.GET("/withdrawals", adminHandler.GetAllWithdrawals)
			admin.GET("/stats", adminHandler.GetStats)

		// 交易对管理
		admin.GET("/pairs", adminHandler.GetTradingPairs)
		admin.POST("/pairs", adminHandler.CreateTradingPair)
		admin.PUT("/pairs/:id", adminHandler.UpdateTradingPair)
		admin.PUT("/pairs/:id/status", adminHandler.UpdateTradingPairStatus)
		admin.PUT("/pairs/:id/simulator", adminHandler.UpdateTradingPairSimulator)

		// 数据生成任务
		admin.POST("/pairs/generate-trades", adminHandler.GenerateTradeDataForPair)
		admin.POST("/pairs/generate-klines", adminHandler.GenerateKlineDataForPair)

			// 任务管理
			admin.GET("/tasks", adminHandler.GetAllTasks)
			admin.GET("/tasks/:id", adminHandler.GetTaskStatus)
			admin.GET("/tasks/:id/logs", adminHandler.GetTaskLogs)
			admin.POST("/tasks/:id/retry", adminHandler.RetryTask)
			admin.GET("/tasks/running", adminHandler.GetRunningTask)

			// K线管理
			admin.POST("/klines/generate", klineHandler.GenerateHistoricalKlines)

			// 手续费管理
			admin.GET("/fees", feeHandler.GetAllFeeRecords)
			admin.GET("/fees/configs", feeHandler.GetFeeConfigs)
			admin.PUT("/users/:id/level", feeHandler.UpdateUserLevel)

			// 系统配置管理
			admin.GET("/configs", adminHandler.GetSystemConfigs)
			admin.GET("/configs/:id", adminHandler.GetSystemConfig)
			admin.PUT("/configs/:id", adminHandler.UpdateSystemConfig)
			admin.POST("/configs/reload", adminHandler.ReloadSystemConfigs)

			// 链配置管理
			admin.GET("/chains", chainHandler.GetChains)
			admin.GET("/chains/:id", chainHandler.GetChain)
			admin.POST("/chains", chainHandler.CreateChain)
			admin.PUT("/chains/:id", chainHandler.UpdateChain)
			admin.PUT("/chains/:id/status", chainHandler.UpdateChainStatus)
			admin.DELETE("/chains/:id", chainHandler.DeleteChain)
		}
	}

	// WebSocket路由
	r.GET("/ws", wsHandler.HandleWebSocket)

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
