import { useEffect, useRef, useState, useCallback } from 'react';

interface UseWebSocketOptions {
  url: string;
  onMessage: (data: any) => void;
  onError?: (error: Event) => void;
  onOpen?: () => void;
  onClose?: () => void;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

interface UseWebSocketReturn {
  isConnected: boolean;
  reconnectAttempts: number;
  sendMessage: (message: any) => void;
  close: () => void;
}

/**
 * Custom hook for WebSocket connections with automatic reconnection
 *
 * @param options - WebSocket configuration options
 * @returns WebSocket connection state and controls
 */
export function useWebSocket({
  url,
  onMessage,
  onError,
  onOpen,
  onClose,
  reconnectInterval = 3000, // Not used with custom backoff
  maxReconnectAttempts = 10,
}: UseWebSocketOptions): UseWebSocketReturn {
  const [isConnected, setIsConnected] = useState(false);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const shouldReconnectRef = useRef(true);
  const reconnectAttemptsRef = useRef(0);

  // Custom backoff pattern: 30s, 15s, 15s, then 60s for all subsequent attempts
  const getReconnectDelay = (attemptNumber: number): number => {
    if (attemptNumber === 0) return 30000; // 30 seconds for first retry
    if (attemptNumber === 1) return 15000; // 15 seconds for second retry
    if (attemptNumber === 2) return 15000; // 15 seconds for third retry
    return 60000; // 60 seconds for all subsequent retries
  };

  // Store callbacks in refs to avoid reconnection when they change
  const onMessageRef = useRef(onMessage);
  const onErrorRef = useRef(onError);
  const onOpenRef = useRef(onOpen);
  const onCloseRef = useRef(onClose);

  // Update refs when callbacks change
  useEffect(() => {
    onMessageRef.current = onMessage;
  }, [onMessage]);

  useEffect(() => {
    onErrorRef.current = onError;
  }, [onError]);

  useEffect(() => {
    onOpenRef.current = onOpen;
  }, [onOpen]);

  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  const connect = useCallback(() => {
    // Don't connect if URL is empty (used to disable connection)
    if (!url) {
      return;
    }

    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        // console.log(`WebSocket connected: ${url}`);
        setIsConnected(true);
        setReconnectAttempts(0);
        reconnectAttemptsRef.current = 0;
        onOpenRef.current?.();
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          onMessageRef.current(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        onErrorRef.current?.(error);
      };

      ws.onclose = () => {
        // console.log(`WebSocket closed: ${url}`);
        setIsConnected(false);
        onCloseRef.current?.();

        // Attempt reconnection with custom backoff pattern
        const currentAttempts = reconnectAttemptsRef.current;
        if (shouldReconnectRef.current && currentAttempts < maxReconnectAttempts && url) {
          const delay = getReconnectDelay(currentAttempts);
          console.log(`Reconnecting in ${delay / 1000}s (attempt ${currentAttempts + 1}/${maxReconnectAttempts})`);

          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptsRef.current += 1;
            setReconnectAttempts((prev) => prev + 1);
            connect();
          }, delay);
        } else if (currentAttempts >= maxReconnectAttempts) {
          console.error(`Max reconnection attempts (${maxReconnectAttempts}) reached for ${url}`);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
    }
  }, [url, maxReconnectAttempts]); // Removed reconnectInterval since we use getReconnectDelay

  const sendMessage = useCallback((message: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }, []);

  const close = useCallback(() => {
    shouldReconnectRef.current = false;

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  useEffect(() => {
    shouldReconnectRef.current = true;
    connect();

    return () => {
      close();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run on mount/unmount, not when connect/close change

  return {
    isConnected,
    reconnectAttempts,
    sendMessage,
    close,
  };
}

/**
 * Hook for subscribing to session updates via WebSocket
 * Only connects when a valid authentication token is available
 */
export function useSessionsWebSocket(onUpdate: (sessions: any[]) => void) {
  const [isEnabled, setIsEnabled] = useState(false);

  // Check for token on mount and when storage changes
  useEffect(() => {
    const token = localStorage.getItem('token');
    setIsEnabled(!!token);

    // Listen for storage changes (login/logout in other tabs)
    const handleStorageChange = () => {
      const newToken = localStorage.getItem('token');
      setIsEnabled(!!newToken);
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  // Use window.location to connect through Vite proxy in dev, or directly in production
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const token = localStorage.getItem('token');
  const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/sessions${token ? `?token=${encodeURIComponent(token)}` : ''}`;

  // Only initialize WebSocket if we have a token
  const ws = useWebSocket({
    url: isEnabled ? wsUrl : '', // Empty URL prevents connection
    onMessage: (data) => {
      if (data.type === 'sessions_update' && data.sessions) {
        onUpdate(data.sessions);
      }
    },
    onError: (error) => {
      // Token might be expired - check if we should reconnect
      const currentToken = localStorage.getItem('token');
      if (!currentToken) {
        setIsEnabled(false); // Stop trying to reconnect without token
      }
    },
    // onOpen: () => console.log('Sessions WebSocket connected'),
    // onClose: () => console.log('Sessions WebSocket disconnected'),
  });

  return ws;
}

/**
 * Hook for subscribing to cluster metrics via WebSocket
 * Only connects when a valid authentication token is available
 */
export function useMetricsWebSocket(onUpdate: (metrics: any) => void) {
  // Use window.location to connect through Vite proxy in dev, or directly in production
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const token = localStorage.getItem('token');

  // Don't connect without a token
  const wsUrl = token
    ? `${protocol}//${window.location.host}/api/v1/ws/cluster?token=${encodeURIComponent(token)}`
    : '';

  return useWebSocket({
    url: wsUrl,
    onMessage: (data) => {
      if (data.type === 'metrics_update' && data.metrics) {
        onUpdate(data.metrics);
      }
    },
    // onOpen: () => console.log('Metrics WebSocket connected'),
    // onClose: () => console.log('Metrics WebSocket disconnected'),
  });
}

/**
 * Hook for subscribing to pod logs via WebSocket
 * Only connects when a valid authentication token is available
 */
export function useLogsWebSocket(namespace: string, podName: string, onLog: (log: string) => void) {
  // Use window.location to connect through Vite proxy in dev, or directly in production
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const token = localStorage.getItem('token');

  // Don't connect without a token
  const wsUrl = token
    ? `${protocol}//${window.location.host}/api/v1/ws/logs/${namespace}/${podName}?token=${encodeURIComponent(token)}`
    : '';

  return useWebSocket({
    url: wsUrl,
    onMessage: (data) => {
      if (typeof data === 'string') {
        onLog(data);
      }
    },
    // onOpen: () => console.log(`Logs WebSocket connected: ${namespace}/${podName}`),
    // onClose: () => console.log(`Logs WebSocket disconnected: ${namespace}/${podName}`),
  });
}
