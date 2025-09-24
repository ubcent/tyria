/**
 * Toast notification component for displaying error messages
 * This integrates with our toast system in lib/toast.ts
 */
'use client';

import React, { useState, useEffect } from 'react';
import { toastStore, ToastMessage } from '../lib/toast';

const ToastNotification: React.FC<{ toast: ToastMessage; onClose: (id: string) => void }> = ({ 
  toast, 
  onClose 
}) => {
  const getIcon = () => {
    switch (toast.type) {
      case 'success': return '✅';
      case 'error': return '❌';
      case 'warning': return '⚠️';
      case 'info': return 'ℹ️';
      default: return '📢';
    }
  };

  const getColorClasses = () => {
    switch (toast.type) {
      case 'success': return 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800 dark:text-green-200';
      case 'error': return 'bg-red-50 border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-800 dark:text-red-200';
      case 'warning': return 'bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-900/20 dark:border-yellow-800 dark:text-yellow-200';
      case 'info': return 'bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900/20 dark:border-blue-800 dark:text-blue-200';
      default: return 'bg-gray-50 border-gray-200 text-gray-800 dark:bg-gray-900/20 dark:border-gray-800 dark:text-gray-200';
    }
  };

  return (
    <div className={`rounded-lg border p-4 shadow-lg ${getColorClasses()}`}>
      <div className="flex items-start">
        <div className="flex-shrink-0">
          <span className="text-lg">{getIcon()}</span>
        </div>
        <div className="ml-3 w-0 flex-1 pt-0.5">
          <p className="text-sm font-medium">{toast.title}</p>
          {toast.message && (
            <p className="mt-1 text-sm opacity-75">{toast.message}</p>
          )}
        </div>
        <div className="ml-4 flex-shrink-0 flex">
          <button
            className="rounded-md inline-flex text-sm font-medium hover:opacity-75 focus:outline-none focus:ring-2 focus:ring-offset-2"
            onClick={() => onClose(toast.id)}
          >
            <span className="sr-only">Close</span>
            ✕
          </button>
        </div>
      </div>
    </div>
  );
};

export const ToastContainer: React.FC = () => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  useEffect(() => {
    const unsubscribe = toastStore.subscribe(setToasts);
    return unsubscribe;
  }, []);

  const handleClose = (id: string) => {
    toastStore.remove(id);
  };

  if (toasts.length === 0) {
    return null;
  }

  return (
    <div className="fixed top-4 right-4 z-50 space-y-4 max-w-sm w-full">
      {toasts.map((toast) => (
        <ToastNotification key={toast.id} toast={toast} onClose={handleClose} />
      ))}
    </div>
  );
};

export default ToastContainer;