const WebSocket = global.WebSocket;

if (!WebSocket) {
  console.error(
    "❌ This Node version does not provide a built-in WebSocket implementation.",
  );
  console.error("Use Node.js 22+ or install a WebSocket client package.");
  process.exit(1);
}

// For testing: extract token from localStorage or use jwt token
// You need to provide a valid JWT token for a user
const token = process.argv[2];
const userId = process.argv[3] || "16";

if (!token) {
  console.error("❌ Usage: node main.js <token> [user_id]");
  console.error('Example: node main.js "eyJhbGc..." 16');
  process.exit(1);
}

const wsUrl = `ws://localhost:8080/api/ws?token=${encodeURIComponent(token)}`;
console.log(`🔗 Connecting to ${wsUrl}`);
console.log(`📍 Testing for user: ${userId}`);

//& opens ws connection instance on this url
const ws = new WebSocket(wsUrl);
let heartbeatTimer = null;

ws.onopen = () => {
  console.log("✅ WebSocket connected!");
  console.log("Waiting for messages...\n");
  heartbeatTimer = setInterval(() => {
    console.log(`⏳ still connected, readyState=${ws.readyState}`);
  }, 5000);
};

// & when wb socket path connection client recieves any data
ws.onmessage = (event) => {
  console.log("📨 Message received:");
  try {
    const parsed = JSON.parse(event.data);
    console.log(JSON.stringify(parsed, null, 2));
  } catch (e) {
    console.log("Raw:", event.data);
  }
  console.log("---");
};

ws.addEventListener("message", (event) => {
  console.log("📩 message event listener fired");
  console.log(
    "payload length:",
    typeof event.data === "string" ? event.data.length : "non-string",
  );
});

ws.onerror = (error) => {
  console.error("❌ WebSocket error:", error.message || error);
};

ws.onclose = () => {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer);
  }
  console.log("❌ WebSocket closed");
  process.exit(0);
};

// Keep the process running
console.log("(Press Ctrl+C to exit)");
