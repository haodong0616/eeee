import toast from 'react-hot-toast';

export const useToast = () => {
  return {
    success: (message: string, options?: { duration?: number }) => {
      return toast.success(message, {
        duration: options?.duration || 3000,
      });
    },

    error: (message: string, options?: { duration?: number }) => {
      return toast.error(message, {
        duration: options?.duration || 4000,
      });
    },

    loading: (message: string) => {
      return toast.loading(message);
    },

    dismiss: (toastId?: string) => {
      toast.dismiss(toastId);
    },

    promise: <T,>(
      promise: Promise<T>,
      messages: {
        loading: string;
        success: string | ((data: T) => string);
        error: string | ((error: any) => string);
      }
    ) => {
      return toast.promise(promise, messages, {
        success: {
          duration: 4000,
        },
        error: {
          duration: 5000,
        },
      });
    },

    info: (message: string, options?: { duration?: number }) => {
      return toast(message, {
        icon: 'ℹ️',
        duration: options?.duration || 3000,
      });
    },

    warning: (message: string, options?: { duration?: number }) => {
      return toast(message, {
        icon: '⚠️',
        duration: options?.duration || 4000,
        style: {
          background: '#1a1f35',
          color: '#fff',
          border: '1px solid #f59e0b',
        },
      });
    },

    custom: (message: string, options?: any) => {
      return toast(message, options);
    },
  };
};

// 导出便捷方法（无需hook）
export const showToast = {
  success: (message: string, options?: { duration?: number }) => {
    return toast.success(message, {
      duration: options?.duration || 3000,
    });
  },

  error: (message: string, options?: { duration?: number }) => {
    return toast.error(message, {
      duration: options?.duration || 4000,
    });
  },

  loading: (message: string) => {
    return toast.loading(message);
  },

  dismiss: (toastId?: string) => {
    toast.dismiss(toastId);
  },

  promise: <T,>(
    promise: Promise<T>,
    messages: {
      loading: string;
      success: string | ((data: T) => string);
      error: string | ((error: any) => string);
    }
  ) => {
    return toast.promise(promise, messages, {
      success: {
        duration: 4000,
      },
      error: {
        duration: 5000,
      },
    });
  },

  info: (message: string, options?: { duration?: number }) => {
    return toast(message, {
      icon: 'ℹ️',
      duration: options?.duration || 3000,
    });
  },

  warning: (message: string, options?: { duration?: number }) => {
    return toast(message, {
      icon: '⚠️',
      duration: options?.duration || 4000,
      style: {
        background: '#1a1f35',
        color: '#fff',
        border: '1px solid #f59e0b',
      },
    });
  },
};


