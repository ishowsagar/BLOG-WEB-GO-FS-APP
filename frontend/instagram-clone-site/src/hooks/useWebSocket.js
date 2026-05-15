import { useEffect, useRef, useCallback } from "react";

export const useWebSocket = (token, endpointPath = "/api/ws") => {
  const wsRef = useRef(null);
  const messageHandlersRef = useRef([]);
  const reconnectTimeoutRef = useRef(null);

  useEffect(() => {
    if (!token) {
      console.log("No token, WebSocket not connecting");
      return;
    }

    const connectWebSocket = () => {
      try {
        // Connect to WebSocket endpoint with token as query parameter
        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";

        // since we are sending token, we won't send to client as it is but an encoded
        const wsUrl = `${protocol}//localhost:8080${endpointPath}?token=${encodeURIComponent(token)}`; //sending token through qp,it does not supports header

        // #1 Opening new instance of webSocket connection
        const ws = new WebSocket(wsUrl); // have to provide the url where the handler is listening for ws request where -> conn is migrated into ws conn and readers waiting for incoming data & writer sending response with ws writeJson method

        // #2 On opening notify client
        ws.onopen = () => {
          console.log("✅ WebSocket connected");
          // Token already sent in URL query param
        };

        //  data written to ws url, checking .onmessage passing func which recieves the wrote data

        //#3 request-response method for ws conn on frontend
        ws.onmessage = (event) => {
          // event is type of data being passed between ws, which is being expected by backend
          try {
            const notification = JSON.parse(event.data);
            console.log("📩 Notification received:", notification);
            // Call all registered handlers
            messageHandlersRef.current.forEach((handler) => {
              try {
                handler(notification);
              } catch (e) {
                console.error("Handler error:", e);
              }
            });
          } catch (e) {
            console.error("Failed to parse notification:", e);
          }
        };

        ws.onerror = (error) => {
          console.error("❌ WebSocket error:", error);
        };

        ws.onclose = () => {
          console.log("🔌 WebSocket closed");
          // Attempt to reconnect after 3 seconds
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log("🔄 Attempting to reconnect...");
            connectWebSocket();
          }, 3000);
        };

        wsRef.current = ws;
      } catch (error) {
        console.error("Failed to create WebSocket:", error);
      }
    };

    connectWebSocket();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [token, endpointPath]);

  // Register a handler to be called on messages
  const subscribe = useCallback((handler) => {
    messageHandlersRef.current.push(handler);
    return () => {
      messageHandlersRef.current = messageHandlersRef.current.filter(
        (h) => h !== handler,
      );
    };
  }, []);

  // Send a message through WebSocket
  const send = useCallback((message) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket not ready");
    }
  }, []);

  return {
    subscribe,
    send,
    isConnected: wsRef.current?.readyState === WebSocket.OPEN,
  };
};
