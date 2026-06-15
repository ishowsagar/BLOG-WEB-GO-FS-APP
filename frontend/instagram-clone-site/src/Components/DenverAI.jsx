import React, { useState, useRef, useEffect } from "react";
import { useOutletContext } from "react-router-dom";
import { apiUrl } from "../Services/apiConfig";
import "./DenverAI.css";

// Sample suggestions for interactive mock usage
const SUGGESTED_PROMPTS = [
  {
    icon: "✍️",
    label: "Draft a travel post",
    text: "Draft an engaging Denvergram post about a weekend escape to the mountains.",
  },
  {
    icon: "💡",
    label: "Generate bio ideas",
    text: "Generate 3 professional yet creative Denvergram bios for a web developer.",
  },
  {
    icon: "🚀",
    label: "Explain Go channels",
    text: "Give a simple, visual explanation of how channels work in Go.",
  },
  {
    icon: "🎨",
    label: "Caption for photography",
    text: "Create a poetic, minimalist caption for a sunset ocean photo.",
  },
];

// Mock answers database for realistic prompt interactions
const MOCK_ANSWERS = {
  "Draft an engaging Denvergram post about a weekend escape to the mountains.":
    "🌲 **Mountain Air & Quiet Paths** 🏔️\n\nSometimes, the best way to move forward is to unplug and head high. Spent the weekend escaping the city noise and trading it for whispering pines and crisp mountain peaks. \n\nThere's something incredibly centering about being surrounded by heights that have stood for thousands of years. It reminds you how small our daily worries are in the grand scheme.\n\n*📍 Alpine Peaks*\n\n🎒 **Pack list essentials:**\n- Good boots 🥾\n- Hot coffee in a flask ☕\n- Zero cellular connection 📵\n\n#MountainEscape #AlpsTravel #Unplugged #NatureVibes #Wanderlust #HikingLife",

  "Generate 3 professional yet creative Denvergram bios for a web developer.":
    "Here are 3 unique bio options for your profile:\n\n**Option 1: Tech + Creative**\n💻 Web Developer | Transforming caffeine into clean, interactive code ☕\n✨ Building pixel-perfect, accessible web experiences\n📂 Check out my latest project below!\n👇 [github.com/developer]\n\n**Option 2: Minimalist**\n🚀 Code. Deploy. Repeat.\n🛠️ Crafting the modern web with React & Go\n🌱 Lifelong learner & open-source contributor\n✉️ DM for collaborations!\n\n**Option 3: Storyteller**\n🌐 Designing digital spaces where code meets human interaction\n🎨 Full Stack Developer by day, tinkerer by night\n⛰️ Outdoors enthusiast when the screen is closed\n🔗 Portfolio: [yourdomain.com]",

  "Give a simple, visual explanation of how channels work in Go.":
    "Think of **Go Channels** like a **pneumatic tube mail system** in a modern office building: \n\n1. **The Goroutines (Workers):** Imagine you have two workers at different desks. Desk A (Goroutine A) is generating reports, and Desk B (Goroutine B) is signing them.\n2. **The Channel (The Tube):** A channel is the physical pipe connecting Desk A to Desk B. It has a specific size (it can be empty, buffered to hold 3 reports, or unbuffered).\n3. **Sending (`ch <- value`):** Desk A rolls up a report, puts it in a cartridge, and drops it into the tube. \n4. **Blocking / Waiting:** If the tube is unbuffered, Desk A *must wait* until Desk B actually takes the cartridge out of the other end before Desk A can go back to work. This is Go's built-in synchronization! No mutexes needed.\n5. **Receiving (`<-ch`):** Desk B listens for the cartridge, pops it open, and begins signing.\n\nIt is a safe, elegant way to share memory by communicating, rather than communicating by sharing memory! 🚀",

  "Create a poetic, minimalist caption for a sunset ocean photo.":
    'Here are a few poetic caption options:\n\n🌅 *Option 1:*\n"Where the sky dips its toes in gold, and the world goes quiet."\n\n✨ *Option 2:*\n"Daylight’s softest, final sigh. 🌊"\n\n🍊 *Option 3:*\n"Sunsets are proof that endings can be beautiful, too."\n\n🍂 *Option 4:*\n"Trading gold for starlight."',
};

export default function DenverAI() {
  const [messages, setMessages] = useState([]);
  const [inputText, setInputText] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef(null);


  const outletContext = useOutletContext();
  const subscribeNotifications = outletContext?.subscribeNotifications;

  // Listen for WebSocket messages of type "denverai_prompt"
  useEffect(() => {
    if (!subscribeNotifications) return;

    console.log("DenverAI subscribing to WebSocket notification feed...");
    // for working with recieved notificatins & validating payload of notification
    const unsubscribe = subscribeNotifications((notification) => {
      if (notification && notification.type === "denverai_prompt") {
        console.log("DenverAI component received backend response via WS:", notification);
        const aiMsg = {
          id: Date.now(),
          sender: "ai",
          text: notification.content,
          timestamp: new Date().toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          }),
        };
        setMessages((prev) => [...prev, aiMsg]);
        setIsTyping(false);
      }
    });

    return () => {
      if (unsubscribe) {
        console.log("DenverAI unsubscribing from WebSocket notification feed...");
        unsubscribe();
      }
    };
  }, [subscribeNotifications]);

  // Scroll to bottom on new messages
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isTyping]);

  // Handle sending a message
  const handleSend = async (textToSend) => {
    const text = textToSend || inputText.trim();
    if (!text) return;

    // Add user message to local chat state
    const userMsg = {
      id: Date.now(),
      sender: "user",
      text: text,
      timestamp: new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    };

    setMessages((prev) => [...prev, userMsg]);
    if (!textToSend) setInputText("");
    setIsTyping(true);

    const token = localStorage.getItem("token");

    try {
      const response = await fetch(apiUrl("/api/denverai/new"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": token,
        },
        body: JSON.stringify({ prompt: text }),
      });

      if (!response.ok) {
        throw new Error("Backend response error status: " + response.status);
      }

      console.log("DenverAI prompt successfully dispatched to backend.");
      // We do not add the response here since it is re-routed via WebSocket.
    } catch (err) {
      console.warn("DenverAI backend post failed, falling back to mock replies:", err);

      // Fallback: If backend is offline or fails, wait 1.2s and use mock answers database
      setTimeout(() => {
        let replyText = "";
        if (MOCK_ANSWERS[text]) {
          replyText = MOCK_ANSWERS[text];
        } else {
          replyText = `Backend connection error. \n\n*This is a fallback response. Try starting the server or checking that RabbitMQ, PNS, and your websocket connections are fully operational!*`;
        }

        const aiMsg = {
          id: Date.now() + 1,
          sender: "ai",
          text: replyText,
          timestamp: new Date().toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          }),
        };

        setMessages((prev) => [...prev, aiMsg]);
        setIsTyping(false);
      }, 1200);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const clearChat = () => {
    setMessages([]);
  };

  // func that send prompt to the handler for retreving response
  const askAI = () => {
    console.log("user has sent request to ask ai.");
  };

  return (
    <div className="denver-ai-container">
      {/* Sidebar - Quick Prompts & Stats */}
      <div className="denver-sidebar denver-glass">
        <div className="denver-brand-section">
          <div className="denver-ai-logo-outer">
            <svg
              className="denver-prompt-icon"
              style={{ width: "32px", height: "32px", fill: "white" }}
              viewBox="0 0 24 24"
            >
              <path d="M12 2L14.7 8.3L21 9.6L16.2 14.3L17.7 21L12 17.3L6.3 21L7.8 14.3L3 9.6L9.3 8.3L12 2Z" />
            </svg>
            <div className="denver-ai-logo-glow" />
          </div>
          <h2 className="denver-brand-title">Denver AI</h2>
          <p className="denver-brand-tagline">
            Your glassmorphic assistant for content creation, coding queries,
            and profile design.
          </p>
        </div>

        <div className="denver-sidebar-section">
          <span className="denver-section-title">💡 Suggested Prompts</span>
          <div className="denver-suggested-prompts">
            {SUGGESTED_PROMPTS.map((prompt, index) => (
              <button
                key={index}
                className="denver-prompt-pill"
                onClick={() => handleSend(prompt.text)}
                disabled={isTyping}
              >
                <span className="denver-prompt-icon">{prompt.icon}</span>
                <span>{prompt.label}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="denver-status-panel">
          <div className="denver-status-row">
            <span>Model:</span>
            <span className="denver-status-value">gemini-2.5-flash</span>
          </div>
          <div className="denver-status-row">
            <span>Status:</span>
            <span className="denver-status-value">
              <span className="denver-indicator" /> Online
            </span>
          </div>
          <div className="denver-status-row">
            <span>Connection:</span>
            <span className="denver-status-value" style={{ color: "#3b82f6" }}>
              Secure WebSocket
            </span>
          </div>
        </div>
      </div>

      {/* Main Chat Area */}
      <div className="denver-chat-panel denver-glass">
        {/* Chat Header */}
        <div className="denver-chat-header">
          <div className="denver-chat-header-info">
            <div className="denver-chat-header-avatar">
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="currentColor"
              >
                <path d="M19,2H5A3,3 0 0,0 2,5V19A3,3 0 0,0 5,22H19A3,3 0 0,0 22,19V5A3,3 0 0,0 19,2M12,4A4,4 0 0,1 16,8A4,4 0 0,1 12,12A4,4 0 0,1 8,8A4,4 0 0,1 12,4M19,19H5V18C5,15.79 8.58,14 12,14C15.42,14 19,15.79 19,18V19Z" />
              </svg>
            </div>
            <div className="denver-chat-header-details">
              <span className="denver-chat-header-title">
                Denver Intelligent Core
              </span>
              <span className="denver-chat-header-status">
                <span className="denver-indicator" /> Ready to assist
              </span>
            </div>
          </div>
          <div className="denver-chat-header-actions">
            {messages.length > 0 && (
              <button
                className="denver-action-btn"
                onClick={clearChat}
                title="Clear conversation"
              >
                <svg
                  width="18"
                  height="18"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            )}
          </div>
        </div>

        {/* Chat Messages */}
        <div className="denver-messages-container">
          {messages.length === 0 ? (
            <div className="denver-welcome-container">
              <div className="denver-welcome-glow-icon">🚀</div>
              <h3 className="denver-welcome-title">Ask Denver AI anything</h3>
              <p className="denver-welcome-desc">
                Welcome to Denver AI Core. Generate creative blog drafts, refine
                your bio content, or ask questions instantly. Pick a quick
                prompt below or start typing.
              </p>

              <div className="denver-capabilities-grid">
                <div className="denver-cap-card">
                  <span className="denver-cap-icon">✍️</span>
                  <span className="denver-cap-label">Draft Content</span>
                  <span className="denver-cap-text">
                    Generate ready-to-post drafts with hashtags and tags.
                  </span>
                </div>
                <div className="denver-cap-card">
                  <span className="denver-cap-icon">💡</span>
                  <span className="denver-cap-label">Optimizations</span>
                  <span className="denver-cap-text">
                    Refine your profile descriptions and bios effortlessly.
                  </span>
                </div>
                <div className="denver-cap-card">
                  <span className="denver-cap-icon">⚡</span>
                  <span className="denver-cap-label">Instant Replies</span>
                  <span className="denver-cap-text">
                    Powered by the lightning-fast Gemini 2.5 Flash model.
                  </span>
                </div>
                <div className="denver-cap-card">
                  <span className="denver-cap-icon">🧩</span>
                  <span className="denver-cap-label">Explain Concepts</span>
                  <span className="denver-cap-text">
                    Ask complex programming or configuration questions.
                  </span>
                </div>
              </div>
            </div>
          ) : (
            messages.map((msg) => (
              <div
                key={msg.id}
                className={`denver-msg-row ${msg.sender === "user" ? "is-user" : "is-ai"}`}
              >
                <div className="denver-msg-bubble">
                  {/* Handle linebreaks in messages nicely */}
                  <div style={{ whiteSpace: "pre-line" }}>{msg.text}</div>
                  <div className="denver-msg-info">
                    <span>{msg.timestamp}</span>
                    {msg.sender === "ai" && <span>• gemini</span>}
                  </div>
                </div>
              </div>
            ))
          )}

          {isTyping && (
            <div className="denver-msg-row is-ai">
              <div className="denver-msg-bubble">
                <div className="denver-typing-indicator">
                  <span className="denver-typing-dot" />
                  <span className="denver-typing-dot" />
                  <span className="denver-typing-dot" />
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input Composer */}
        <div className="denver-composer">
          <div className="denver-input-wrapper">
            <textarea
              className="denver-input"
              placeholder="Ask Denver AI a question..."
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              onKeyDown={handleKeyDown}
              rows="1"
              disabled={isTyping}
            />
            <button
              className="denver-send-btn"
              onClick={() => handleSend()}
              disabled={!inputText.trim() || isTyping}
            >
              <svg className="denver-send-icon" viewBox="0 0 24 24">
                <path d="M2,21L23,12L2,3V10L17,12L2,14V21Z" />
              </svg>
            </button>
          </div>
          <div className="denver-disclaimer">
            Denver AI can make mistakes. Verify important information.
          </div>
        </div>
      </div>
    </div>
  );
}
