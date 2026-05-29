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
  const [threadHasMessages, setThreadHasMessages] = useState({});

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
          const hasMessages = items.length > 0;
          setThreadHasMessages((prev) => ({
            ...prev,
            [activePeer]: hasMessages,
          }));
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
          setThreadHasMessages((prev) => ({
            ...prev,
            [Number(s) === Number(currentUserId) ? Number(r) : Number(s)]: true,
          }));
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
    setThreadHasMessages((prev) => ({ ...prev, [activePeer]: true }));
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

  return (
    <section
      style={{
        padding: "2rem",
        minHeight: "72vh",
        background: "linear-gradient(180deg,#f3f9ff,#eef7ff)",
      }}
    >
      <div
        style={{
          display: "flex",
          gap: 16,
          alignItems: "flex-start",
          maxWidth: 1200,
          margin: "0 auto",
        }}
      >
        <aside
          style={{
            width: 360,
            overflowY: "auto",
            background: "#f8fbff",
            borderRadius: 12,
            padding: "0.5rem",
            boxShadow: "inset 0 1px 0 rgba(255,255,255,0.6)",
            paddingBottom: 20,
          }}
        >
          {profiles.map((p) => (
            <button
              key={p.id}
              onClick={() => {
                setActivePeer(p.id);
                // send a dm-type payload to the DM websocket to start/activate the thread
                try {
                  const payload = {
                    sender_id: currentUserId,
                    reciever_id: p.id,
                    type: "dm",
                    content: "",
                  };
                  if (sendDm) sendDm(payload);
                } catch (e) {
                  console.warn("Failed to send initial DM payload:", e);
                }
              }}
              style={{
                width: "100%",
                padding: "0.85rem 1rem",
                border: "none",
                borderBottom: "1px solid rgba(15,23,42,0.03)",
                textAlign: "left",
                background: activePeer === p.id ? "#ffffff" : "transparent",
                cursor: "pointer",
                display: "flex",
                gap: 12,
                alignItems: "center",
                boxShadow:
                  activePeer === p.id
                    ? "0 6px 18px rgba(16,24,40,0.04)"
                    : "none",
              }}
            >
              <img
                src={resolveAvatar(p)}
                alt={p.name}
                onError={(e) => {
                  const currentSrc = e.currentTarget.dataset.fallbackStage || "0";
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
                  border: "2px solid #eef2ff",
                }}
              />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <div
                    style={{
                      fontWeight: 700,
                      color: "#0f172a",
                      fontSize: "0.98rem",
                    }}
                  >
                    {p.name}
                  </div>
                  <div
                    style={{
                      marginLeft: "auto",
                      fontSize: "0.8rem",
                      color: "#94a3b8",
                    }}
                  >
                    now
                  </div>
                </div>
                <div
                  style={{
                    fontSize: "0.86rem",
                    color: threadHasMessages[p.id] ? "#0f172a" : "#64748b",
                    fontWeight: threadHasMessages[p.id] ? 700 : 400,
                    whiteSpace: "nowrap",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                  }}
                >
                  {threadHasMessages[p.id]
                    ? "New message"
                    : "No messages yet — say hello!"}
                </div>
              </div>
            </button>
          ))}
        </aside>

        <div
          style={{
            flex: 1,
            maxWidth: 820,
            borderRadius: 12,
            overflow: "hidden",
            display: "grid",
            gridTemplateRows: "auto 1fr auto",
            background: "#ffffff",
            boxShadow: "0 12px 36px rgba(16,24,40,0.08)",
            border: "1px solid rgba(15,23,42,0.04)",
            marginLeft: 8,
            alignSelf: "flex-start",
          }}
        >
          <div
            style={{
              padding: "0.9rem 1rem",
              borderBottom: "1px solid #f1f5f9",
              display: "flex",
              alignItems: "center",
              gap: 12,
            }}
          >
            {activePeer ? (
              <>
                <img
                  src={
                    resolveAvatar(
                      profiles.find((x) => x.id === activePeer) || {
                        id: activePeer,
                        pfp: null,
                      },
                    )
                  }
                  onError={(e) => {
                    const activeProfile =
                      profiles.find((x) => x.id === activePeer) || {
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
                    border: "2px solid #eef2ff",
                  }}
                />
                <div>
                  <div style={{ fontWeight: 800, color: "#0f172a" }}>
                    {profiles.find((p) => p.id === activePeer)?.name}
                  </div>
                  <div style={{ fontSize: "0.85rem", color: "#94a3b8" }}>
                    Active now
                  </div>
                </div>
              </>
            ) : (
              <div style={{ fontWeight: 700, color: "#0f172a" }}>
                Select a conversation
              </div>
            )}
          </div>

          <div
            ref={messagesListRef}
            style={{ padding: "1rem", overflowY: "auto", minHeight: 0 }}
          >
            {activePeer ? (
              <div style={{ display: "grid", gap: 12 }}>
                {loadingMessages ? (
                  <div style={{ color: "#94a3b8" }}>Loading messages...</div>
                ) : messagesErr ? (
                  <div style={{ color: "#ef4444" }}>{messagesErr}</div>
                ) : messages.length === 0 ? (
                  <div style={{ color: "#64748b" }}>
                    No messages yet. Start the conversation.
                  </div>
                ) : (
                  messages.map((m, idx) => {
                    const isOutgoing =
                      Number(m.sender_id) === Number(currentUserId);
                    return (
                      <div
                        key={`${m.id || idx}-${m.created_at || idx}-${m.content?.slice(0, 8)}`}
                        style={{
                          justifySelf: isOutgoing ? "end" : "start",
                          maxWidth: "72%",
                          padding: "0.6rem 0.8rem",
                          borderRadius: 12,
                          background: isOutgoing
                            ? "linear-gradient(135deg,#ff8a5b,#ff4d9d)"
                            : "#eef3ff",
                          color: isOutgoing ? "#fff" : "#0f172a",
                          boxShadow: isOutgoing
                            ? "0 6px 18px rgba(255,77,157,0.12)"
                            : "0 2px 6px rgba(15,23,42,0.03)",
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
            style={{
              padding: "0.85rem",
              borderTop: "1px solid #f1f5f9",
              display: "flex",
              gap: 10,
              alignItems: "center",
            }}
          >
            <input
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
                borderRadius: 10,
                border: "1px solid #e6eef6",
                background: "#fff",
              }}
            />
            <button
              onClick={() => handleSend()}
              disabled={!activePeer || !draft.trim()}
              style={{
                padding: "0.6rem 0.95rem",
                borderRadius: 10,
                border: "none",
                background:
                  activePeer && draft.trim()
                    ? "linear-gradient(135deg,#ff4d9d,#ff8a5b)"
                    : "#f1f5f9",
                color: activePeer && draft.trim() ? "#fff" : "#94a3b8",
                fontWeight: 700,
                cursor: activePeer && draft.trim() ? "pointer" : "default",
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
