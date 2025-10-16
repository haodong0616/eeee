import { BrowserProvider } from 'ethers';

declare global {
  interface Window {
    ethereum?: any;
  }
}

export async function connectWallet(): Promise<string | null> {
  if (typeof window.ethereum === 'undefined') {
    alert('请安装 MetaMask 钱包!');
    return null;
  }

  try {
    const provider = new BrowserProvider(window.ethereum);
    const accounts = await provider.send('eth_requestAccounts', []);
    return accounts[0].toLowerCase();
  } catch (error) {
    console.error('连接钱包失败:', error);
    return null;
  }
}

export async function signMessage(message: string): Promise<string | null> {
  if (typeof window.ethereum === 'undefined') {
    return null;
  }

  try {
    const provider = new BrowserProvider(window.ethereum);
    const signer = await provider.getSigner();
    const signature = await signer.signMessage(message);
    return signature;
  } catch (error) {
    console.error('签名失败:', error);
    return null;
  }
}

export async function getConnectedWallet(): Promise<string | null> {
  if (typeof window.ethereum === 'undefined') {
    return null;
  }

  try {
    const provider = new BrowserProvider(window.ethereum);
    const accounts = await provider.send('eth_accounts', []);
    return accounts.length > 0 ? accounts[0].toLowerCase() : null;
  } catch (error) {
    console.error('获取钱包地址失败:', error);
    return null;
  }
}

