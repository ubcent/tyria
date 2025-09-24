/**
 * Simple toast notification system for error handling
 * This provides a minimal implementation that can be enhanced later
 */

export interface ToastMessage {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message?: string;
  duration?: number;
}

// Global toast store for managing notifications
class ToastStore {
  private listeners: Set<(toasts: ToastMessage[]) => void> = new Set();
  private toasts: ToastMessage[] = [];

  subscribe(listener: (toasts: ToastMessage[]) => void) {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  private notify() {
    this.listeners.forEach(listener => listener([...this.toasts]));
  }

  show(toast: Omit<ToastMessage, 'id'>) {
    const id = Math.random().toString(36).substring(2);
    const newToast: ToastMessage = {
      id,
      duration: 5000,
      ...toast,
    };

    this.toasts.push(newToast);
    this.notify();

    // Auto-remove after duration
    if (newToast.duration && newToast.duration > 0) {
      setTimeout(() => this.remove(id), newToast.duration);
    }

    return id;
  }

  remove(id: string) {
    this.toasts = this.toasts.filter(toast => toast.id !== id);
    this.notify();
  }

  clear() {
    this.toasts = [];
    this.notify();
  }
}

export const toastStore = new ToastStore();

// Convenience methods for common toast types
export const toast = {
  success: (title: string, message?: string) =>
    toastStore.show({ type: 'success', title, message }),

  error: (title: string, message?: string) =>
    toastStore.show({ type: 'error', title, message }),

  warning: (title: string, message?: string) =>
    toastStore.show({ type: 'warning', title, message }),

  info: (title: string, message?: string) =>
    toastStore.show({ type: 'info', title, message }),
};