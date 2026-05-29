import { useState, useEffect, useContext, useRef } from "react";
import { apiUrl } from "../Services/apiConfig";
import { RealtimeContext } from "../Layout/MainLayout";

export default function Messages() {
  const [activePeer, setActivePeer] = useState(null);
  const [draft, setDraft] = useState("");

  // profiles state (populated from API or fallback placeholders)
  const [profiles, setProfiles] = useState([
    { id: 101, name: "Alice" },
    { id: 102, name: "Bob" },
    { id: 103, name: "Charlie" },
    { id: 104, name: "Dana" },
  ]);
  const [loadingProfiles, setLoadingProfiles] = useState(false);
  const [profilesErr, setProfilesErr] = useState(null);

  // realtime context (ws send for DM)
  const { sendDm, subscribeDm, currentUserId } =
    useContext(RealtimeContext) || {};

  const [messages, setMessages] = useState([]);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [messagesErr, setMessagesErr] = useState(null);
  const messagesListRef = useRef(null);
  const [threadLastSeenAt, setThreadLastSeenAt] = useState({});
  const [nowTick, setNowTick] = useState(Date.now());

  function updateThreadLastSeen(peerId, createdAt) {
    const key = Number(peerId);
    if (!key || !createdAt) return;

    const timestamp = new Date(createdAt).getTime();
    if (Number.isNaN(timestamp)) return;

    setThreadLastSeenAt((prev) => {
      const existing = prev[key] || 0;
      return {
        ...prev,
        [key]: Math.max(existing, timestamp),
      };
    });
  }

  function formatLastSeen(peerId) {
    const timestamp = threadLastSeenAt[Number(peerId)];
    if (!timestamp) return "offline";

    const diffMs = Math.max(0, nowTick - timestamp);
    const diffMins = Math.floor(diffMs / 60000);
    if (diffMins <= 0) return "Last seen just now";
    if (diffMins === 1) return "Last seen 1 min ago";
    return `Last seen ${diffMins} mins ago`;
  }

  useEffect(() => {
    async function loadFollowings() {
      setLoadingProfiles(true);
      setProfilesErr(null);
      try {
        const token = localStorage.getItem("token");
        const res = await fetch(apiUrl("/api/followings"), {
          method: "GET",
          headers: token
            ? {
                Authorization: token,
                "Content-Type": "application/json",
              }
            : { "Content-Type": "application/json" },
        });

        const text = await res.text();
        let data = {};
        try {
          data = JSON.parse(text);
        } catch {
          data = { Status: text };
        }

        if (!res.ok)
          throw new Error(data.Status || "failed to fetch followings");

        // expected response shape: { Ok: true, Data: [...] }
        const items = Array.isArray(data.Data) ? data.Data : [];
        if (items.length === 0) {
          setProfiles([]);
        } else {
          // map API items to { id, name }
          const mapped = items.map((it) => ({
            id: it.ID || it.id || it.user_id || it.UserID || 0,
            name: it.Name || it.name || it.username || it.Username || "Unknown",
            // possible avatar fields from API: avatar, avatar_url, profile_picture, pfp
            pfp:
              it.avatar ||
              it.avatar_url ||
              it.profile_picture ||
              it.pfp ||
              it.profile_pic ||
              it.ProfilePic ||
              it.Avatar ||
              null,
          }));
          setProfiles(mapped.filter((p) => p.id));
        }
      } catch (err) {
        setProfilesErr(String(err));
      } finally {
        setLoadingProfiles(false);
      }
    }

    loadFollowings();
  }, []);

  function avatarUrl(id) {
    return `https://i.pravatar.cc/128?u=${id}`;
  }

  function fallbackAvatar(id) {
    const n = Number(id) || 0;
    const gender = n % 2 === 0 ? "men" : "women";
    return `https://randomuser.me/api/portraits/${gender}/${n % 100}.jpg`;
  }

  function resolveAvatar(p) {
    return p.pfp || avatarUrl(p.id) || fallbackAvatar(p.id);
  }

  // load messages for a peer when activePeer changes
  useEffect(() => {
    let cancelled = false;
    async function loadMessages() {
      if (!activePeer) return;
      setLoadingMessages(true);
      setMessagesErr(null);
      try {
        const token = localStorage.getItem("token");
        const res = await fetch(apiUrl(`/api/messages?peer_id=${activePeer}`), {
          method: "GET",
          headers: token
            ? { Authorization: token, "Content-Type": "application/json" }
            : { "Content-Type": "application/json" },
        });
        const text = await res.text();
        let data = {};
        try {
          data = JSON.parse(text);
        } catch {
          data = { Status: text };
        }
        if (!res.ok) throw new Error(data.Status || "failed to fetch messages");
        const items = Array.isArray(data.Data) ? data.Data : [];
        if (!cancelled) {
          const lastItem = items[items.length - 1];
          if (lastItem) {
            updateThreadLastSeen(
              activePeer,
              lastItem.created_at || lastItem.CreatedAt,
            );
          }
          setMessages(
            items.slice(-20).map((it) => ({
              id: it.id || it.ID,
              sender_id: it.sender_id || it.SenderID,
              reciever_id: it.reciever_id || it.RecieverID,
              content: it.content || it.Content,
              created_at: it.created_at || it.CreatedAt,
            })),
          );
        }
      } catch (err) {
        setMessagesErr(String(err));
      } finally {
        setLoadingMessages(false);
      }
    }
    loadMessages();
    return () => {
      cancelled = true;
    };
  }, [activePeer]);

  // subscribe to incoming dm websocket messages and append if relevant
  useEffect(() => {
    if (!subscribeDm) return undefined;
    const unsub = subscribeDm((incoming) => {
      try {
        if (!incoming) return;
        const t = String(incoming.type || incoming.Type || "").toLowerCase();
        if (t !== "dm" && t !== "dm_msg") return;
        const s = incoming.sender_id || incoming.SenderID;
        const r = incoming.reciever_id || incoming.RecieverID;
        if (!activePeer) return;
        if (
          Number(s) === Number(activePeer) ||
          Number(r) === Number(activePeer)
        ) {
          const msg = {
            id: incoming.id || incoming.ID || undefined,
            sender_id: s,
            reciever_id: r,
            content: incoming.content || incoming.Content || "",
            created_at: incoming.created_at || incoming.CreatedAt,
          };
          updateThreadLastSeen(
            Number(s) === Number(currentUserId) ? Number(r) : Number(s),
            msg.created_at,
          );
          setMessages((prev) => [...prev, msg].slice(-20));
        }
      } catch (e) {
        console.error("Failed to process incoming dm", e);
      }
    });
    return unsub;
  }, [subscribeDm, activePeer]);

  // send handler for messages (uses websocket if available)
  async function handleSend() {
    if (!activePeer || !draft.trim()) return;

    const content = draft.trim();

    const outgoing = {
      sender_id: currentUserId,
      reciever_id: activePeer,
      type: "dm",
      content,
      created_at: new Date().toISOString(),
    };

    // optimistic UI append
    updateThreadLastSeen(activePeer, outgoing.created_at);
    setMessages((prev) => [...prev, outgoing].slice(-20));
    setDraft("");

    try {
      if (sendDm) {
        sendDm(outgoing);
      } else {
        // fallback: POST to messages endpoint
        const token = localStorage.getItem("token");
        await fetch(apiUrl("/api/messages"), {
          method: "POST",
          headers: token
            ? { Authorization: token, "Content-Type": "application/json" }
            : { "Content-Type": "application/json" },
          body: JSON.stringify({ reciever_id: activePeer, content }),
        });
      }
    } catch (e) {
      console.warn("failed to send message", e);
    }
  }

  useEffect(() => {
    if (!messagesListRef.current) return;
    messagesListRef.current.scrollTop = messagesListRef.current.scrollHeight;
  }, [messages, activePeer]);

  useEffect(() => {
    const timer = setInterval(() => setNowTick(Date.now()), 60000);
    return () => clearInterval(timer);
  }, []);

  return (
    <section
      className="messages-page"
      style={{
        padding: "2rem",
        minHeight: "72vh",
        background:
          "radial-gradient(circle at top left, rgba(59, 130, 246, 0.16), transparent 34%), radial-gradient(circle at top right, rgba(244, 114, 182, 0.13), transparent 28%), linear-gradient(180deg, #060b16 0%, #090f1d 55%, #0d1426 100%)",
      }}
    >
      <div
        className="messages-shell"
        style={{
          display: "flex",
          gap: 16,
          alignItems: "flex-start",
          maxWidth: 1200,
          margin: "0 auto",
        }}
      >
        <aside
          className="messages-list-panel"
          style={{
            width: 360,
            overflowY: "auto",
            background:
              "linear-gradient(180deg, rgba(18, 26, 44, 0.98), rgba(12, 18, 34, 0.96))",
            borderRadius: 18,
            padding: "0.5rem",
            boxShadow:
              "0 18px 40px rgba(0,0,0,0.26), inset 0 1px 0 rgba(255,255,255,0.04)",
            paddingBottom: 20,
            border: "none",
          }}
        >
          {profiles.map((p) => (
            <button
              key={p.id}
              className={`messages-peer-button${activePeer === p.id ? " is-active" : ""}`}
              onClick={() => {
                setActivePeer(p.id);
              }}
              style={{
                width: "100%",
                padding: "0.85rem 1rem",
                border: "none",
                borderBottom: "1px solid rgba(148, 163, 184, 0.08)",
                textAlign: "left",
                background:
                  activePeer === p.id
                    ? "linear-gradient(135deg, rgba(59, 130, 246, 0.28), rgba(244, 114, 182, 0.18))"
                    : "rgba(255,255,255,0.015)",
                cursor: "pointer",
                display: "flex",
                gap: 12,
                alignItems: "center",
                boxShadow:
                  activePeer === p.id ? "0 10px 24px rgba(0,0,0,0.22)" : "none",
                color: "#f1f5f9",
              }}
            >
              <img
                src={resolveAvatar(p)}
                alt={p.name}
                onError={(e) => {
                  const currentSrc =
                    e.currentTarget.dataset.fallbackStage || "0";
                  if (currentSrc === "0") {
                    e.currentTarget.dataset.fallbackStage = "1";
                    e.currentTarget.src = avatarUrl(p.id);
                    return;
                  }
                  if (currentSrc === "1") {
                    e.currentTarget.dataset.fallbackStage = "2";
                    e.currentTarget.src = fallbackAvatar(p.id);
                    return;
                  }
                  e.currentTarget.src =
                    "https://ui-avatars.com/api/?name=" +
                    encodeURIComponent(p.name || "User") +
                    "&background=EEF2FF&color=0F172A&size=128";
                }}
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: "50%",
                  objectFit: "cover",
                  border: "2px solid rgba(203, 213, 225, 0.18)",
                  boxShadow: "0 8px 18px rgba(0,0,0,0.18)",
                }}
              />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <div
                    style={{
                      fontWeight: 700,
                      color: "#ffffff",
                      fontSize: "0.98rem",
                    }}
                  >
                    {p.name}
                  </div>
                  <div
                    style={{
                      marginLeft: "auto",
                      fontSize: "0.8rem",
                      color: "#cbd5e1",
                    }}
                  >
                    open
                  </div>
                </div>
              </div>
            </button>
          ))}
        </aside>

        <div
          className="messages-chat-panel"
          style={{
            flex: 1,
            maxWidth: 820,
            borderRadius: 20,
            overflow: "hidden",
            display: "grid",
            gridTemplateRows: "auto 1fr auto",
            background:
              "linear-gradient(180deg, rgba(10, 16, 32, 0.97), rgba(8, 14, 28, 0.95))",
            boxShadow: "0 24px 60px rgba(0,0,0,0.34)",
            border: "none",
            marginLeft: 8,
            alignSelf: "flex-start",
          }}
        >
          <div
            style={{
              padding: "0.9rem 1rem",
              borderBottom: "none",
              boxShadow: "inset 0 -1px 0 rgba(148, 163, 184, 0.08)",
              display: "flex",
              alignItems: "center",
              gap: 12,
              background:
                "linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(8, 15, 31, 0.9))",
            }}
          >
            {activePeer ? (
              <>
                <img
                  src={resolveAvatar(
                    profiles.find((x) => x.id === activePeer) || {
                      id: activePeer,
                      pfp: null,
                    },
                  )}
                  onError={(e) => {
                    const activeProfile = profiles.find(
                      (x) => x.id === activePeer,
                    ) || {
                      id: activePeer,
                      name: "User",
                    };
                    const currentSrc =
                      e.currentTarget.dataset.fallbackStage || "0";
                    if (currentSrc === "0") {
                      e.currentTarget.dataset.fallbackStage = "1";
                      e.currentTarget.src = avatarUrl(activeProfile.id);
                      return;
                    }
                    if (currentSrc === "1") {
                      e.currentTarget.dataset.fallbackStage = "2";
                      e.currentTarget.src = fallbackAvatar(activeProfile.id);
                      return;
                    }
                    e.currentTarget.src =
                      "https://ui-avatars.com/api/?name=" +
                      encodeURIComponent(activeProfile.name || "User") +
                      "&background=EEF2FF&color=0F172A&size=128";
                  }}
                  style={{
                    width: 44,
                    height: 44,
                    borderRadius: "50%",
                    objectFit: "cover",
                    border: "2px solid rgba(148, 163, 184, 0.18)",
                    boxShadow: "0 8px 18px rgba(0,0,0,0.22)",
                  }}
                />
                <div>
                  <div style={{ fontWeight: 800, color: "#f8fafc" }}>
                    {profiles.find((p) => p.id === activePeer)?.name}
                  </div>
                  <div style={{ fontSize: "0.85rem", color: "#94a3b8" }}>
                    {formatLastSeen(activePeer)}
                  </div>
                </div>
              </>
            ) : (
              <div style={{ fontWeight: 700, color: "#f8fafc" }}>
                Select a conversation
              </div>
            )}
          </div>

          <div
            className="messages-thread"
            style={{
              padding: "1rem",
              overflowY: "auto",
              minHeight: 0,
              background:
                "linear-gradient(180deg, rgba(8, 14, 28, 0.32), rgba(8, 14, 28, 0.12))",
            }}
          >
            {activePeer ? (
              <div style={{ display: "grid", gap: 12 }}>
                {loadingMessages ? (
                  <div style={{ color: "#94a3b8" }}>Loading messages...</div>
                ) : messagesErr ? (
                  <div style={{ color: "#ef4444" }}>{messagesErr}</div>
                ) : messages.length === 0 ? (
                  <div style={{ color: "#94a3b8" }}>
                    No messages yet. Start the conversation.
                  </div>
                ) : (
                  messages
                    .filter((m) => String(m.content || "").trim())
                    .map((m, idx) => {
                      const isOutgoing =
                        Number(m.sender_id) === Number(currentUserId);
                      return (
                        <div
                          key={`${m.id || idx}-${m.created_at || idx}-${m.content?.slice(0, 8)}`}
                          className={`messages-bubble ${isOutgoing ? "is-outgoing" : "is-incoming"}`}
                          style={{
                            justifySelf: isOutgoing ? "end" : "start",
                            maxWidth: "72%",
                            padding: "0.6rem 0.8rem",
                            borderRadius: 16,
                            background: isOutgoing
                              ? "linear-gradient(135deg, #f97316 0%, #ec4899 100%)"
                              : "linear-gradient(135deg, rgba(30, 41, 59, 0.96), rgba(15, 23, 42, 0.96))",
                            color: isOutgoing ? "#fff" : "#e2e8f0",
                            boxShadow: isOutgoing
                              ? "0 12px 28px rgba(236, 72, 153, 0.18)"
                              : "0 12px 28px rgba(0,0,0,0.24)",
                            border: isOutgoing
                              ? "1px solid rgba(255,255,255,0.08)"
                              : "1px solid rgba(148, 163, 184, 0.12)",
                          }}
                        >
                          <div style={{ fontSize: "0.95rem" }}>{m.content}</div>
                        </div>
                      );
                    })
                )}
              </div>
            ) : (
              <div style={{ color: "#64748b" }}>
                Select a profile on the left to view or start a chat.
              </div>
            )}
          </div>

          <div
            className="messages-composer"
            style={{
              padding: "0.85rem",
              borderTop: "none",
              boxShadow: "inset 0 1px 0 rgba(148, 163, 184, 0.08)",
              display: "flex",
              gap: 10,
              alignItems: "center",
              background: "rgba(8, 15, 31, 0.94)",
            }}
          >
            <input
              className="messages-input"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  // trigger send
                  if (activePeer && draft.trim()) {
                    handleSend();
                  }
                }
              }}
              placeholder={
                activePeer
                  ? "Write a message..."
                  : "Select a conversation first"
              }
              disabled={!activePeer}
              aria-label="message-input"
              style={{
                flex: 1,
                padding: "0.65rem 0.9rem",
                borderRadius: 14,
                border: "1px solid rgba(148, 163, 184, 0.1)",
                background: "rgba(13, 19, 34, 0.96)",
                color: "#f8fafc",
                boxShadow: "inset 0 1px 0 rgba(255,255,255,0.03)",
              }}
            />
            <button
              className="messages-send-button"
              onClick={() => handleSend()}
              disabled={!activePeer || !draft.trim()}
              style={{
                padding: "0.6rem 0.95rem",
                borderRadius: 14,
                border: "none",
                background:
                  activePeer && draft.trim()
                    ? "linear-gradient(135deg, #f97316 0%, #ec4899 100%)"
                    : "rgba(148, 163, 184, 0.14)",
                color: activePeer && draft.trim() ? "#fff" : "#94a3b8",
                fontWeight: 700,
                cursor: activePeer && draft.trim() ? "pointer" : "default",
                boxShadow:
                  activePeer && draft.trim()
                    ? "0 12px 24px rgba(236, 72, 153, 0.24)"
                    : "none",
              }}
            >
              Send
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}
