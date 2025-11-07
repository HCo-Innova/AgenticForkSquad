export type WSEvent = { type: string; payload?: Record<string, any> };

function defaultWsUrl(): string {
  const envUrl = import.meta.env.VITE_WS_URL as string | undefined
  if (envUrl && envUrl.trim() !== "") {
    // Allow relative "/ws" or absolute
    if (envUrl.startsWith("/")) {
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:"
      return `${proto}//${window.location.host}${envUrl}`
    }
    return envUrl
  }
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:"
  return `${proto}//${window.location.host}/ws`
}

export function useWebSocket(url = defaultWsUrl(), subscribeTo?: string[]) {
  let socket: WebSocket | null = null;
  const listeners = new Set<(ev: WSEvent) => void>();
  let connectTimer: number | null = null;
  let shouldCloseWhileConnecting = false;

  function connect() {
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return;
    socket = new WebSocket(url);
    socket.onopen = () => {
      // optional ping
      socket?.send(JSON.stringify({ type: "ping" }));
      // optional subscribe
      if (subscribeTo && subscribeTo.length > 0) {
        socket?.send(JSON.stringify({ type: "subscribe", payload: { events: subscribeTo } }));
      }
      if (shouldCloseWhileConnecting) {
        // defer actual close until open to avoid browser error log
        socket?.close();
        shouldCloseWhileConnecting = false;
      }
    };
    socket.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data as string) as WSEvent;
        listeners.forEach((l) => l(data));
      } catch (_) {}
    };
  }

  function onMessage(cb: (ev: WSEvent) => void) {
    listeners.add(cb);
    return () => {
      listeners.delete(cb);
    }
  }

  function send(ev: WSEvent) {
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(JSON.stringify(ev));
  }

  function close() {
    if (connectTimer !== null) {
      clearTimeout(connectTimer);
      connectTimer = null;
    }
    if (socket) {
      if (socket.readyState === WebSocket.CONNECTING) {
        shouldCloseWhileConnecting = true;
        return;
      }
      try { socket.close(); } catch {}
      socket = null;
    }
    listeners.clear();
  }

  return { connect, onMessage, send, close };
}
