package noncemanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

// WalletNonce 钱包nonce记录
type WalletNonce struct {
	ID        uint      `gorm:"primaryKey"`
	Address   string    `gorm:"uniqueIndex:idx_address_chain;size:42;not null"`
	ChainID   int       `gorm:"uniqueIndex:idx_address_chain;not null"`
	Nonce     uint64    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// NonceManager Nonce 管理器（线程安全）
type NonceManager struct {
	db         *gorm.DB
	locks      map[string]*sync.Mutex // address_chainid -> mutex
	locksMutex sync.RWMutex           // 保护 locks map
	nonceCache map[string]uint64      // address_chainid -> nonce (内存缓存)
	cacheMutex sync.RWMutex           // 保护 cache
}

// NewNonceManager 创建 Nonce 管理器
func NewNonceManager(db *gorm.DB) *NonceManager {
	// 确保表存在
	db.AutoMigrate(&WalletNonce{})

	return &NonceManager{
		db:         db,
		locks:      make(map[string]*sync.Mutex),
		nonceCache: make(map[string]uint64),
	}
}

// getOrCreateLock 获取或创建钱包的互斥锁
func (nm *NonceManager) getOrCreateLock(address string, chainID int) *sync.Mutex {
	key := fmt.Sprintf("%s_%d", address, chainID)

	nm.locksMutex.RLock()
	lock, exists := nm.locks[key]
	nm.locksMutex.RUnlock()

	if exists {
		return lock
	}

	// 需要创建新锁
	nm.locksMutex.Lock()
	defer nm.locksMutex.Unlock()

	// 双重检查
	if lock, exists := nm.locks[key]; exists {
		return lock
	}

	lock = &sync.Mutex{}
	nm.locks[key] = lock
	return lock
}

// AcquireNonce 获取下一个可用的 nonce（线程安全）
// 返回：nonce, release函数, error
func (nm *NonceManager) AcquireNonce(rpcURL, address string, chainID int) (uint64, func(), error) {
	// 1. 获取该钱包的锁
	lock := nm.getOrCreateLock(address, chainID)
	lock.Lock()

	// 2. 从缓存或数据库获取当前 nonce
	cacheKey := fmt.Sprintf("%s_%d", address, chainID)

	nm.cacheMutex.RLock()
	cachedNonce, hasCached := nm.nonceCache[cacheKey]
	nm.cacheMutex.RUnlock()

	var currentNonce uint64

	if hasCached {
		// 使用缓存的 nonce
		currentNonce = cachedNonce
	} else {
		// 从数据库加载
		var record WalletNonce
		err := nm.db.Where("address = ? AND chain_id = ?", address, chainID).First(&record).Error

		if err == gorm.ErrRecordNotFound {
			// 第一次使用，从链上查询
			onChainNonce, err := nm.getOnChainNonce(rpcURL, address)
			if err != nil {
				lock.Unlock()
				return 0, nil, fmt.Errorf("failed to get on-chain nonce: %w", err)
			}
			currentNonce = onChainNonce

			// 保存到数据库
			record = WalletNonce{
				Address: address,
				ChainID: chainID,
				Nonce:   currentNonce,
			}
			nm.db.Create(&record)
		} else if err != nil {
			lock.Unlock()
			return 0, nil, fmt.Errorf("failed to query nonce: %w", err)
		} else {
			currentNonce = record.Nonce
		}

		// 更新缓存
		nm.cacheMutex.Lock()
		nm.nonceCache[cacheKey] = currentNonce
		nm.cacheMutex.Unlock()
	}

	// 3. 递增 nonce（内存）
	nextNonce := currentNonce
	nm.cacheMutex.Lock()
	nm.nonceCache[cacheKey] = currentNonce + 1
	nm.cacheMutex.Unlock()

	// 4. 创建 release 函数
	released := false
	release := func() {
		if !released {
			released = true

			// 更新数据库（异步，避免阻塞）
			go func() {
				nm.db.Model(&WalletNonce{}).
					Where("address = ? AND chain_id = ?", address, chainID).
					Update("nonce", currentNonce+1)
			}()

			lock.Unlock()
		}
	}

	return nextNonce, release, nil
}

// getOnChainNonce 从链上查询实际的 nonce
func (nm *NonceManager) getOnChainNonce(rpcURL, address string) (uint64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(address))
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

// SyncFromChain 从链上同步 nonce（用于恢复或重新校准）
func (nm *NonceManager) SyncFromChain(rpcURL, address string, chainID int) error {
	onChainNonce, err := nm.getOnChainNonce(rpcURL, address)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s_%d", address, chainID)

	// 获取锁
	lock := nm.getOrCreateLock(address, chainID)
	lock.Lock()
	defer lock.Unlock()

	// 更新缓存
	nm.cacheMutex.Lock()
	nm.nonceCache[cacheKey] = onChainNonce
	nm.cacheMutex.Unlock()

	// 更新数据库
	return nm.db.Model(&WalletNonce{}).
		Where("address = ? AND chain_id = ?", address, chainID).
		Updates(map[string]interface{}{
			"nonce": onChainNonce,
		}).Error
}

// ResetNonce 重置 nonce（危险操作，仅用于恢复）
func (nm *NonceManager) ResetNonce(address string, chainID int, newNonce uint64) error {
	cacheKey := fmt.Sprintf("%s_%d", address, chainID)

	// 获取锁
	lock := nm.getOrCreateLock(address, chainID)
	lock.Lock()
	defer lock.Unlock()

	// 更新缓存
	nm.cacheMutex.Lock()
	nm.nonceCache[cacheKey] = newNonce
	nm.cacheMutex.Unlock()

	// 更新数据库
	return nm.db.Model(&WalletNonce{}).
		Where("address = ? AND chain_id = ?", address, chainID).
		Updates(map[string]interface{}{
			"nonce": newNonce,
		}).Error
}

