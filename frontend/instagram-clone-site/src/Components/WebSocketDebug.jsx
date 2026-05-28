import React, { useState, useEffect, useRef } from "react";
import { wsUrl } from "../Services/apiConfig";

/**
 * WebSocketDebug component - shows real-time WebSocket connection status and incoming messages
 * Usage: <WebSocketDebug token={jwtToken} />
 */
export const WebSocketDebug = ({ token }) => {
  console.log(
    "🟢 WebSocketDebug component mounted, token:",
    token ? "YES" : "NO",
  );
  const [messages, setMessages] = useState([]);
  const [connected, setConnected] = useState(false);
  const messagesEndRef = useRef(null);

  const latestMessage = messages[messages.length - 1];

  useEffect(() => {
    if (!token) {
      console.log("No token provided to WebSocketDebug");
      return;
    }

    const socketUrl = `${wsUrl("/api/ws")}?token=${encodeURIComponent(token)}`;

    const ws = new WebSocket(socketUrl);

    ws.onopen = () => {
      console.log("✅ WebSocket connected");
      setConnected(true);
      setMessages((prev) => [
        ...prev,
        {
          timestamp: new Date().toLocaleTimeString(),
          type: "system",
          content: "✅ Connected",
        },
      ]);
    };

    ws.onmessage = (event) => {
      console.log("🔴 RAW MESSAGE RECEIVED:", event.data);
      try {
        const data = JSON.parse(event.data);
        console.log("✅ PARSED DATA:", data);
        console.log("📊 Data keys:", Object.keys(data));
        console.log("  sender_id:", data.sender_id);
        console.log("  receiver_id:", data.receiver_id);
        console.log("  reciever_id:", data.reciever_id);
        console.log("  type:", data.type);
        console.log("  content:", data.content);
        console.log("  status:", data.status);

        setMessages((prev) => [
          ...prev,
          {
            timestamp: new Date().toLocaleTimeString(),
            type: data.type || data.status || "unknown",
            sender_id: data.sender_id,
            receiver_id: data.receiver_id || data.reciever_id,
            user_id: data.user_id,
            content: data.content || data.status,
            raw: JSON.stringify(data, null, 2),
          },
        ]);
      } catch (e) {
        console.error("❌ Failed to parse message:", e);
        console.error("Raw data was:", event.data);
        setMessages((prev) => [
          ...prev,
          {
            timestamp: new Date().toLocaleTimeString(),
            type: "error",
            content: `Parse error: ${e.message} | Raw: ${event.data}`,
          },
        ]);
      }
    };

    ws.onerror = (error) => {
      console.error("❌ WebSocket error:", error);
      setMessages((prev) => [
        ...prev,
        {
          timestamp: new Date().toLocaleTimeString(),
          type: "error",
          content: `❌ Error: ${error.message || "unknown"}`,
        },
      ]);
    };

    ws.onclose = () => {
      console.log("❌ WebSocket closed");
      setConnected(false);
      setMessages((prev) => [
        ...prev,
        {
          timestamp: new Date().toLocaleTimeString(),
          type: "system",
          content: "❌ Disconnected",
        },
      ]);
    };

    return () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, [token]);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  return (
    <div
      className="WebSocketDebug"
      data-websocket-debug="true"
      style={{
        position: "fixed",
        top: 12,
        right: 20,
        minWidth: 180,
        maxWidth: 260,
        backgroundColor: "rgba(30, 30, 30, 0.92)",
        color: "#d4d4d4",
        border: `1px solid ${connected ? "#4ec9b0" : "#f48771"}`,
        borderRadius: 999,
        padding: "8px 12px",
        fontFamily: "monospace",
        fontSize: 11,
        zIndex: 9999,
        boxShadow: "0 4px 12px rgba(0, 0, 0, 0.35)",
        display: "flex",
        alignItems: "center",
        gap: 8,
        backdropFilter: "blur(10px)",
      }}
    >
      <span
        style={{
          alignItems: "center",
          whiteSpace: "nowrap",
          fontWeight: 700,
          color: connected ? "#4ec9b0" : "#f48771",
        }}
      >
        {connected ? "🟢 WS Connected" : "🔴 WS Disconnected"}
      </span>

      <span
        style={{
          color: "#858585",
          whiteSpace: "nowrap",
          overflow: "hidden",
          textOverflow: "ellipsis",
          maxWidth: 120,
        }}
      >
        {latestMessage?.content
          ? `Last: ${latestMessage.content}`
          : "Waiting..."}
        <div ref={messagesEndRef} />
      </span>
    </div>
  );
};

export default WebSocketDebug;
