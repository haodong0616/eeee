'use client';

import { Provider } from 'react-redux';
import { store } from '@/lib/store/store';
import { useEffect } from 'react';
import { useAppDispatch } from '@/lib/store/hooks';
import { setToken } from '@/lib/store/slices/authSlice';

function AuthInitializer({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      dispatch(setToken(token));
    }
  }, [dispatch]);

  return <>{children}</>;
}

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <Provider store={store}>
      <AuthInitializer>{children}</AuthInitializer>
    </Provider>
  );
}

