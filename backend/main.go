package main

import (
	"expchange-backend/config"
	"expchange-backend/database"
	"expchange-backend/handlers"
	"expchange-backend/kline"
	"expchange-backend/matching"
	"expchange-backend/middleware"
	"expchange-backend/services"
	"expchange-backend/simulator"
	"expchange-backend/websocket"
	"log"
	"os"

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

	// 初始化充值验证服务
	depositVerifier, err := services.NewDepositVerifier()
	if err != nil {
		log.Printf("⚠️  充值验证服务初始化失败: %v", err)
		log.Println("充值功能将不可用")
	} else {
		go depositVerifier.Start()
		log.Println("✅ 充值验证队列已启动")
	}

	// 初始化提现处理服务
	// 注意：需要配置私钥环境变量 PLATFORM_PRIVATE_KEY
	privateKey := os.Getenv("PLATFORM_PRIVATE_KEY")
	if privateKey != "" {
		withdrawProcessor, err := services.NewWithdrawProcessor(privateKey)
		if err != nil {
			log.Printf("⚠️  提现处理服务初始化失败: %v", err)
			log.Println("提现功能将不可用")
		} else {
			go withdrawProcessor.Start()
			log.Println("✅ 提现处理队列已启动")
		}
	} else {
		log.Println("⚠️  未配置 PLATFORM_PRIVATE_KEY，提现功能将不可用")
	}

	// 初始化市场模拟器（可选，用于演示）
	// 如果设置了环境变量 ENABLE_SIMULATOR=true，则启动模拟器
	if os.Getenv("ENABLE_SIMULATOR") == "true" {
		log.Println("🎮 启用市场模拟器")

		// 启动价格模拟器
		marketSim := simulator.NewTrendSimulator(wsHub)
		marketSim.Start()

		// 启动订单簿模拟器
		orderbookSim := simulator.NewOrderBookSimulator(matchingManager)
		orderbookSim.Start()
	}

	// 初始化Gin
	r := gin.Default()

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
			admin.GET("/stats", adminHandler.GetStats)
			admin.POST("/pairs", adminHandler.CreateTradingPair)
			admin.PUT("/pairs/:id/status", adminHandler.UpdateTradingPairStatus)
			admin.POST("/klines/generate", klineHandler.GenerateHistoricalKlines)
			admin.GET("/fees", feeHandler.GetAllFeeRecords)
			admin.GET("/fees/configs", feeHandler.GetFeeConfigs)
			admin.PUT("/users/:id/level", feeHandler.UpdateUserLevel)
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
