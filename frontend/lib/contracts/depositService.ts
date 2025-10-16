import { ethers } from 'ethers';
import { USDT_ABI } from '../chains/config';
import type { ChainConfig } from '../services/api';

/**
 * 充值服务（支持多链）
 */
export class DepositService {
  /**
   * 执行 USDT 充值转账
   * @param provider Web3 Provider
   * @param amount 充值金额（USDT）
   * @param chainConfig 链配置（从后端获取）
   * @returns 交易 hash
   */
  static async depositUSDT(
    provider: any,
    amount: string,
    chainConfig: ChainConfig
  ): Promise<string> {
    try {
      if (!provider) {
        throw new Error('Please connect wallet first');
      }

      if (!chainConfig) {
        throw new Error('Chain config is required');
      }

      // 创建以太坊 provider
      const ethersProvider = new ethers.BrowserProvider(provider);
      const signer = await ethersProvider.getSigner();
      const userAddress = await signer.getAddress();

      // 获取网络信息
      const network = await ethersProvider.getNetwork();
      const currentChainId = Number(network.chainId);
      
      // 验证用户在正确的链上
      if (currentChainId !== chainConfig.chain_id) {
        throw new Error(`Please switch to ${chainConfig.chain_name} (Chain ID: ${chainConfig.chain_id})`);
      }

      console.log('💰 开始充值流程');
      console.log('链:', chainConfig.chain_name);
      console.log('用户地址:', userAddress);
      console.log('充值金额:', amount, 'USDT');
      console.log('平台地址:', chainConfig.platform_deposit_address);

      // 创建 USDT 合约实例
      const usdtContract = new ethers.Contract(
        chainConfig.usdt_contract_address,
        USDT_ABI,
        signer
      );

      // 检查用户 USDT 余额
      const balance = await usdtContract.balanceOf(userAddress);
      const balanceInUSDT = ethers.formatUnits(balance, chainConfig.usdt_decimals);
      console.log('当前 USDT 余额:', balanceInUSDT);

      // 转换金额为 wei
      const amountInWei = ethers.parseUnits(amount, chainConfig.usdt_decimals);

      // 检查余额是否足够
      if (balance < amountInWei) {
        throw new Error(`Insufficient USDT balance. You have ${balanceInUSDT} USDT`);
      }

      // 执行转账
      console.log('🔄 发送转账交易...');
      const tx = await usdtContract.transfer(
        chainConfig.platform_deposit_address,
        amountInWei
      );

      console.log('✅ 交易已发送，hash:', tx.hash);
      console.log('⏳ 等待交易确认...');

      // 等待交易确认
      const receipt = await tx.wait();
      console.log('✅ 交易已确认！区块号:', receipt.blockNumber);

      return tx.hash;
    } catch (error: any) {
      console.error('❌ 充值失败:', error);
      
      // 处理用户拒绝交易
      if (error.code === 'ACTION_REJECTED' || error.code === 4001) {
        throw new Error('Transaction cancelled by user');
      }
      
      // 处理其他错误
      throw new Error(error.message || 'Deposit failed');
    }
  }

  /**
   * 查询 USDT 余额
   * @param provider Web3 Provider
   * @param address 钱包地址
   * @param chainConfig 链配置
   * @returns USDT 余额
   */
  static async getUSDTBalance(
    provider: any,
    address: string,
    chainConfig?: ChainConfig
  ): Promise<string> {
    try {
      if (!chainConfig) {
        return '0';
      }

      const ethersProvider = new ethers.BrowserProvider(provider);

      const usdtContract = new ethers.Contract(
        chainConfig.usdt_contract_address,
        USDT_ABI,
        ethersProvider
      );

      const balance = await usdtContract.balanceOf(address);
      return ethers.formatUnits(balance, chainConfig.usdt_decimals);
    } catch (error) {
      console.error('获取余额失败:', error);
      return '0';
    }
  }

  /**
   * 验证交易状态
   * @param provider Web3 Provider
   * @param txHash 交易 hash
   * @returns 交易是否成功
   */
  static async verifyTransaction(
    provider: any,
    txHash: string
  ): Promise<boolean> {
    try {
      const ethersProvider = new ethers.BrowserProvider(provider);
      const receipt = await ethersProvider.getTransactionReceipt(txHash);
      
      if (!receipt) {
        return false;
      }

      return receipt.status === 1;
    } catch (error) {
      console.error('验证交易失败:', error);
      return false;
    }
  }
}

