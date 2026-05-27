import Header from "../Components/Header";
import Footer from "../Components/Footer";
import Sidebar from "../Components/Sidebar";
import { WebSocketDebug } from "../Components/WebSocketDebug";
import { Outlet, useLocation } from "react-router-dom";
import { useEffect, useRef, useState, createContext, useContext } from "react";
import { useWebSocket } from "../hooks/useWebSocket";
import sendIcon from "../assets/icons/send.png";

const postDataContext = createContext();
export const RealtimeContext = createContext(null);

export default function MainLayout() {
  const location = useLocation();
  const isAboutRoute = location.pathname === "/about";

  const [postData, setPostData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [notification, setNotification] = useState(null);
  const [showDmInbox, setShowDmInbox] = useState(false);
  const [dmMessages, setDmMessages] = useState([]);
  const [activeDmPeerId, setActiveDmPeerId] = useState(null);
  const [dmDraft, setDmDraft] = useState("");
  const [roomDraft, setRoomDraft] = useState("");
  const [dmInboxStatus, setDmInboxStatus] = useState("");
  const [showRoomInbox, setShowRoomInbox] = useState(false);
  const [roomMessages, setRoomMessages] = useState([]);
  const [activeRoomId, setActiveRoomId] = useState(null);
  const [postBatch, SetPostBatch] = useState([]);
  const [cursor, setCursor] = useState("");
  const [nextBatchLoading, setNextBatchLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [feedLoadFailed, setFeedLoadFailed] = useState(false);
  const bottomRef = useRef(null);
  const dmThreadRef = useRef(null);

  const token = localStorage.getItem("token");
  const { subscribe: subscribeNotifications, send: sendNotifications } =
    useWebSocket(token);
  const { subscribe: subscribeDm, send: sendDm } = useWebSocket(
    token,
    "/api/ws/dm",
  );

  function joinRoom(roomId) {
    if (!currentUserId || !sendNotifications) return;
    sendNotifications({
      sender_id: currentUserId,
      reciever_id: 0,
      room_id: roomId,
      room_status: true,
      type: "room_msg",
      content: "",
      post_id: 0,
    });
  }

  function getCurrentUserIdFromToken() {
    if (!token) return null;
    try {
      const payload = token.split(".")[1];
      if (!payload) return null;
      const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
      const decoded = JSON.parse(atob(normalized));
      return decoded.user_id ? Number(decoded.user_id) : null;
    } catch (err) {
      console.log("failed to decode token", err);
      return null;
    }
  }

  const currentUserId = getCurrentUserIdFromToken();

  useEffect(() => {
    let hideTimer = null;

    const unsubscribe = subscribeNotifications((incomingNotification) => {
      // normalize common key variants from different publishers
      if (incomingNotification) {
        if (
          incomingNotification.receiver_id &&
          !incomingNotification.reciever_id
        ) {
          incomingNotification.reciever_id = incomingNotification.receiver_id;
        }
        if (incomingNotification.senderId && !incomingNotification.sender_id) {
          incomingNotification.sender_id = incomingNotification.senderId;
        }
        // normalize room id variants
        if (incomingNotification.roomId && !incomingNotification.room_id) {
          incomingNotification.room_id = incomingNotification.roomId;
        }
        if (incomingNotification.room && !incomingNotification.room_id) {
          incomingNotification.room_id = incomingNotification.room;
        }
      }

      setNotification(incomingNotification);

      const notificationType = String(
        incomingNotification?.type || "",
      ).toLowerCase();
      const hasDmShape =
        incomingNotification?.sender_id != null &&
        incomingNotification?.reciever_id != null &&
        (notificationType === "dm" || notificationType === "dm_msg");

      if (hasDmShape) {
        const { peerId } = appendDmMessage(
          setDmMessages,
          incomingNotification,
          currentUserId,
        );

        setDmInboxStatus(
          Number(incomingNotification.sender_id) === currentUserId
            ? "message sent"
            : "new message received",
        );

        if (activeDmPeerId === null) {
          setActiveDmPeerId(peerId);
        }
      }

      if (incomingNotification?.room_id) {
        // log the incoming room notification for debugging
        console.log("📩 Room notification received:", incomingNotification);

        const incomingRoomId = Number(incomingNotification.room_id);

        // If this message originated from the current user, skip appending it
        // because the UI already uses optimistic updates when sending.
        const isFromSelf =
          incomingNotification?.sender_id != null &&
          Number(incomingNotification.sender_id) === currentUserId;

        // ensure the room panel is visible and the active room is set
        setShowRoomInbox(true);
        setActiveRoomId(incomingRoomId);

        if (isFromSelf) {
          console.log(
            "Skipping append for own message (server-broadcast), sender:",
            incomingNotification.sender_id,
          );
        } else {
          const roomMessage = {
            ...incomingNotification,
            room_id: incomingRoomId,
            direction:
              Number(incomingNotification.sender_id) === currentUserId
                ? "outgoing"
                : "incoming",
          };

          setRoomMessages((prev) => {
            const next = [...prev, roomMessage];
            console.log("roomMessages updated", next.length, next);
            return next;
          });
        }
      }

      if (hideTimer) {
        clearTimeout(hideTimer);
      }

      hideTimer = setTimeout(() => setNotification(null), 5000);
    });

    return () => {
      if (hideTimer) clearTimeout(hideTimer);
      unsubscribe();
    };
  }, [subscribeNotifications, currentUserId, activeRoomId]);

  useEffect(() => {
    if (!token || !currentUserId) return undefined;

    const unsubscribe = subscribeDm((incomingMessage) => {
      if (incomingMessage.type !== "dm") return;

      const { peerId } = appendDmMessage(
        setDmMessages,
        incomingMessage,
        currentUserId,
      );
      setDmInboxStatus(
        Number(incomingMessage.sender_id) === currentUserId
          ? "message sent"
          : "new message received",
      );

      if (activeDmPeerId === null) {
        setActiveDmPeerId(peerId);
      }
    });

    return unsubscribe;
  }, [subscribeDm, token, currentUserId, activeDmPeerId]);

  useEffect(() => {
    if (!showDmInbox || !dmThreadRef.current) return;
    dmThreadRef.current.scrollTop = dmThreadRef.current.scrollHeight;
  }, [dmMessages, showDmInbox]);

  const dmThreads = dmMessages.reduce((accumulator, message) => {
    const peerId = message.peer_id;
    if (!accumulator[peerId]) accumulator[peerId] = [];
    accumulator[peerId].push(message);
    return accumulator;
  }, {});

  const dmThreadEntries = Object.entries(dmThreads)
    .map(([peerId, messages]) => ({
      peerId: Number(peerId),
      messages,
      lastMessage: messages[messages.length - 1],
      peerName:
        messages[messages.length - 1]?.sender_name ||
        messages[messages.length - 1]?.senderName ||
        `User ${peerId}`,
    }))
    .sort(
      (left, right) =>
        right.lastMessage?.created_at?.localeCompare?.(
          left.lastMessage?.created_at ?? "",
        ) ?? 0,
    );

  const activeThreadMessages = activeDmPeerId
    ? dmThreads[activeDmPeerId] || []
    : [];
  const activeThreadTitle =
    dmThreadEntries.find((thread) => thread.peerId === activeDmPeerId)
      ?.peerName || `User ${activeDmPeerId}`;

  const roomThreads = roomMessages.reduce((accumulator, message) => {
    const roomId = Number(message.room_id);
    if (!accumulator[roomId]) accumulator[roomId] = [];
    accumulator[roomId].push(message);
    return accumulator;
  }, {});

  const roomThreadEntries = Object.entries(roomThreads)
    .map(([roomId, messages]) => ({
      roomId: Number(roomId),
      messages,
      lastMessage: messages[messages.length - 1],
    }))
    .sort((left, right) => right.roomId - left.roomId);

  const activeRoomMessages = activeRoomId
    ? roomThreads[activeRoomId] || []
    : [];

  function appendDmMessage(setter, incomingMessage, currentUserIdValue) {
    const senderId = Number(incomingMessage.sender_id);
    const receiverId = Number(incomingMessage.reciever_id);
    const peerId = senderId === currentUserIdValue ? receiverId : senderId;
    const direction = senderId === currentUserIdValue ? "outgoing" : "incoming";

    const normalizedMessage = {
      ...incomingMessage,
      peer_id: peerId,
      direction,
    };

    setter((prev) => {
      const lastMessage = prev[prev.length - 1];
      const isDuplicate =
        lastMessage?.sender_id === normalizedMessage.sender_id &&
        lastMessage?.reciever_id === normalizedMessage.reciever_id &&
        lastMessage?.content === normalizedMessage.content &&
        lastMessage?.direction === normalizedMessage.direction;

      if (isDuplicate) {
        return prev;
      }

      return [...prev, normalizedMessage];
    });

    return { peerId, direction };
  }

  function handleSendInboxDm() {
    if (!currentUserId || !activeDmPeerId || !dmDraft.trim()) return;

    const payload = {
      sender_id: currentUserId,
      reciever_id: activeDmPeerId,
      type: "dm",
      content: dmDraft.trim(),
    };

    setDmMessages((prev) => [
      ...prev,
      { ...payload, peer_id: activeDmPeerId, direction: "outgoing" },
    ]);
    sendDm(payload);
    setDmDraft("");
    setDmInboxStatus("message sent");
  }

  const batchReq = {
    url: cursor
      ? `http://3.84.111.249:8080/api/feed/batch?limit=4&nextCursor=${cursor}`
      : `http://3.84.111.249:8080/api/feed/batch?limit=4`,
    header: { Authorization: token },
    method: "GET",
  };

  async function LoadBatchesFeed() {
    if (feedLoadFailed || nextBatchLoading || !hasMore) return;
    setNextBatchLoading(true);

    try {
      if (!token) return;
      const request = await fetch(batchReq.url, {
        method: batchReq.method,
        headers: batchReq.header,
      });

      const responseText = await request.text();
      let response = {};

      try {
        response = JSON.parse(responseText);
      } catch {
        response = { Status: responseText };
      }

      if (!request.ok) {
        if (request.status === 404) {
          setFeedLoadFailed(true);
          setHasMore(false);
          throw new Error("feed batch route not found");
        }
        throw new Error(response.Status || "failed to load batch feed");
      }

      if (!response.Ok) {
        console.error("batch response not ok", response);
        throw new Error(response.Status || "failed to load batch feed");
      }

      const batchArray = Array.isArray(response.Batch) ? response.Batch : [];
      if (!Array.isArray(response.Batch)) {
        console.warn(
          "response.Batch is not an array, using empty array",
          response.Batch,
        );
      }

      SetPostBatch((prevBatch) => [...prevBatch, ...batchArray]);
      setCursor(response.NextCursor || "");
      setHasMore(Boolean(response.HasMore));
    } catch (err) {
      console.log(err);
    } finally {
      setLoading(false);
      setNextBatchLoading(false);
    }
  }

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }

    if (isAboutRoute) {
      setLoading(false);
      return;
    }

    LoadBatchesFeed();
  }, [token, isAboutRoute]);

  useEffect(() => {
    if (!bottomRef.current || isAboutRoute) return;

    const observer = new IntersectionObserver((entries) => {
      if (
        entries[0].isIntersecting &&
        hasMore &&
        !nextBatchLoading &&
        !feedLoadFailed
      ) {
        LoadBatchesFeed();
      }
    });

    observer.observe(bottomRef.current);
    return () => observer.disconnect();
  }, [hasMore, cursor, nextBatchLoading, feedLoadFailed]);

  if (loading && !hasMore) {
    return (
      <div
        style={{ textAlign: "center", marginTop: "4rem", fontSize: "1.5rem" }}
      >
        Loading...
      </div>
    );
  }

  return (
    <>
      <RealtimeContext.Provider value={{ subscribeDm, sendDm, currentUserId }}>
        <postDataContext.Provider
          value={{ postBatch, SetPostBatch, bottomRef }}
        >
          <section className="Site_wrapper">
            <button
              onClick={() => {
                const willShow = !showRoomInbox;
                setShowRoomInbox(willShow);
                if (willShow) joinRoom(1); // auto-join room 1 when opening
              }}
              style={{
                position: "fixed",
                bottom: "4.2rem",
                right: "1rem",
                padding: "0.85rem 1rem",
                borderRadius: "999px",
                border: "none",
                background: "linear-gradient(135deg, #5b8cff, #6ee7ff)",
                color: "#fff",
                fontWeight: 700,
                boxShadow: "0 16px 40px rgba(0,0,0,0.35)",
                display: "flex",
                alignItems: "center",
                gap: "0.5rem",
                zIndex: 9998,
              }}
              className="floating-msg-btn"
            >
              Group Chat{" "}
              {roomThreadEntries.length > 0
                ? `(${roomThreadEntries.length})`
                : ""}
            </button>

            <button
              onClick={() => setShowDmInbox((prev) => !prev)}
              style={{
                position: "fixed",
                bottom: "1rem",
                right: "1rem",
                padding: "0.85rem 1rem",
                borderRadius: "999px",
                border: "none",
                background: "linear-gradient(135deg, #ff4d9d, #ff8a5b)",
                color: "#fff",
                fontWeight: 700,
                boxShadow: "0 16px 40px rgba(0,0,0,0.35)",
                display: "flex",
                alignItems: "center",
                gap: "0.5rem",
                zIndex: 9998,
              }}
              className="floating-msg-btn"
            >
              <img
                src={sendIcon}
                alt="messages"
                style={{ width: "18px", height: "18px", objectFit: "contain" }}
              />
              Messages{" "}
              {dmThreadEntries.length > 0 ? `(${dmThreadEntries.length})` : ""}
            </button>

            {showDmInbox && (
              <div
                style={{
                  position: "fixed",
                  bottom: "5rem",
                  right: "1rem",
                  width: "min(420px, calc(100vw - 2rem))",
                  height: "min(70vh, 620px)",
                  zIndex: 9999,
                  display: "grid",
                  gridTemplateRows: "auto 1fr auto",
                  borderRadius: "20px",
                  overflow: "hidden",
                  background: "rgba(16, 16, 20, 0.98)",
                  border: "1px solid rgba(255,255,255,0.08)",
                  boxShadow: "0 24px 70px rgba(0,0,0,0.45)",
                  color: "#fff",
                }}
              >
                <div
                  style={{
                    padding: "1rem",
                    borderBottom: "1px solid rgba(255,255,255,0.08)",
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  <div>
                    <div style={{ fontWeight: 800 }}>Inbox</div>
                    <div style={{ fontSize: "0.84rem", opacity: 0.72 }}>
                      Live receiver-side messages
                    </div>
                  </div>
                  <button
                    className="profile_button"
                    onClick={() => setShowDmInbox(false)}
                    style={{ padding: "0.45rem 0.8rem" }}
                  >
                    Close
                  </button>
                </div>

                <div
                  style={{
                    display: "grid",
                    gridTemplateColumns: "140px 1fr",
                    minHeight: 0,
                  }}
                >
                  <div
                    style={{
                      borderRight: "1px solid rgba(255,255,255,0.08)",
                      overflowY: "auto",
                    }}
                  >
                    {dmThreadEntries.length === 0 ? (
                      <div
                        style={{
                          padding: "1rem",
                          color: "rgba(255,255,255,0.65)",
                          fontSize: "0.92rem",
                        }}
                      >
                        No messages yet.
                      </div>
                    ) : (
                      dmThreadEntries.map((thread) => (
                        <button
                          key={thread.peerId}
                          onClick={() => setActiveDmPeerId(thread.peerId)}
                          style={{
                            width: "100%",
                            textAlign: "left",
                            padding: "0.9rem 1rem",
                            border: "none",
                            borderBottom: "1px solid rgba(255,255,255,0.05)",
                            background:
                              activeDmPeerId === thread.peerId
                                ? "rgba(255,255,255,0.08)"
                                : "transparent",
                            color: "#fff",
                          }}
                        >
                          <div style={{ fontWeight: 700 }}>
                            {thread.peerName}
                          </div>
                          <div
                            style={{
                              fontSize: "0.82rem",
                              opacity: 0.72,
                              whiteSpace: "nowrap",
                              overflow: "hidden",
                              textOverflow: "ellipsis",
                            }}
                          >
                            {thread.lastMessage?.content}
                          </div>
                        </button>
                      ))
                    )}
                  </div>

                  <div
                    style={{
                      display: "grid",
                      gridTemplateRows: "1fr auto",
                      minHeight: 0,
                    }}
                  >
                    <div
                      style={{
                        padding: "1rem",
                        borderBottom: "1px solid rgba(255,255,255,0.08)",
                        fontWeight: 700,
                      }}
                    >
                      {activeDmPeerId
                        ? `Chat with ${activeThreadTitle}`
                        : "Select a conversation"}
                    </div>

                    <div
                      ref={dmThreadRef}
                      style={{
                        minHeight: 0,
                        overflowY: "auto",
                        padding: "1rem",
                        display: "grid",
                        gap: "0.65rem",
                        alignContent: "start",
                        background:
                          "linear-gradient(180deg, rgba(255,255,255,0.02), rgba(255,255,255,0.01))",
                      }}
                    >
                      {activeThreadMessages.length === 0 ? (
                        <div style={{ color: "rgba(255,255,255,0.65)" }}>
                          Open a conversation to start chatting.
                        </div>
                      ) : (
                        activeThreadMessages.map((message, index) => (
                          <div
                            key={`${message.direction}-${index}-${message.content}`}
                            style={{
                              justifySelf:
                                message.direction === "outgoing"
                                  ? "end"
                                  : "start",
                              maxWidth: "82%",
                              padding: "0.8rem 0.9rem",
                              borderRadius:
                                message.direction === "outgoing"
                                  ? "16px 16px 6px 16px"
                                  : "16px 16px 16px 6px",
                              background:
                                message.direction === "outgoing"
                                  ? "linear-gradient(135deg, #ff4d9d, #ff8a5b)"
                                  : "rgba(255,255,255,0.08)",
                              color: "#fff",
                              lineHeight: 1.38,
                            }}
                          >
                            <div
                              style={{
                                fontSize: "0.76rem",
                                opacity: 0.8,
                                marginBottom: "0.3rem",
                              }}
                            >
                              {message.direction === "outgoing"
                                ? "You"
                                : message.sender_name ||
                                  message.senderName ||
                                  `User ${message.sender_id}`}
                            </div>
                            <div>{message.content}</div>
                          </div>
                        ))
                      )}
                    </div>

                    <div
                      style={{
                        padding: "1rem",
                        borderTop: "1px solid rgba(255,255,255,0.08)",
                        display: "grid",
                        gap: "0.75rem",
                      }}
                    >
                      <textarea
                        placeholder={
                          activeDmPeerId
                            ? "Write a reply..."
                            : "Select a conversation first..."
                        }
                        value={dmDraft}
                        onChange={(e) => setDmDraft(e.target.value)}
                        rows={3}
                        disabled={!activeDmPeerId}
                        style={{
                          width: "100%",
                          padding: "0.85rem",
                          borderRadius: "12px",
                          border: "1px solid rgba(255,255,255,0.12)",
                          background: "#111",
                          color: "#fff",
                          resize: "vertical",
                          opacity: activeDmPeerId ? 1 : 0.6,
                        }}
                      />
                      <div
                        style={{
                          display: "flex",
                          gap: "0.75rem",
                          alignItems: "center",
                        }}
                      >
                        <button
                          className="profile_button"
                          onClick={handleSendInboxDm}
                          disabled={!activeDmPeerId}
                        >
                          Send
                        </button>
                        {dmInboxStatus && (
                          <span
                            style={{ color: "#9eff9e", fontSize: "0.9rem" }}
                          >
                            {dmInboxStatus}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {showRoomInbox && (
              <div
                style={{
                  position: "fixed",
                  bottom: "5rem",
                  right: "1rem",
                  width: "min(420px, calc(100vw - 2rem))",
                  height: "min(70vh, 620px)",
                  zIndex: 9999,
                  display: "grid",
                  gridTemplateRows: "auto 1fr auto",
                  borderRadius: "20px",
                  overflow: "hidden",
                  background: "rgba(16, 16, 20, 0.98)",
                  border: "1px solid rgba(255,255,255,0.08)",
                  boxShadow: "0 24px 70px rgba(0,0,0,0.45)",
                  color: "#fff",
                }}
              >
                <div
                  style={{
                    padding: "1rem",
                    borderBottom: "1px solid rgba(255,255,255,0.08)",
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  <div>
                    <div style={{ fontWeight: 800 }}>Group Chat</div>
                    <div style={{ fontSize: "0.84rem", opacity: 0.72 }}>
                      Live room messages
                    </div>
                  </div>
                  <button
                    className="profile_button"
                    onClick={() => setShowRoomInbox(false)}
                    style={{ padding: "0.45rem 0.8rem" }}
                  >
                    Close
                  </button>
                </div>

                <div
                  style={{
                    display: "grid",
                    gridTemplateColumns: "140px 1fr",
                    minHeight: 0,
                  }}
                >
                  <div
                    style={{
                      borderRight: "1px solid rgba(255,255,255,0.08)",
                      overflowY: "auto",
                    }}
                  >
                    {roomThreadEntries.length === 0 ? (
                      <div
                        style={{
                          padding: "1rem",
                          color: "rgba(255,255,255,0.65)",
                          fontSize: "0.92rem",
                        }}
                      >
                        No room messages yet.
                      </div>
                    ) : (
                      roomThreadEntries.map((thread) => (
                        <button
                          key={thread.roomId}
                          onClick={() => setActiveRoomId(thread.roomId)}
                          style={{
                            width: "100%",
                            textAlign: "left",
                            padding: "0.9rem 1rem",
                            border: "none",
                            borderBottom: "1px solid rgba(255,255,255,0.05)",
                            background:
                              activeRoomId === thread.roomId
                                ? "rgba(255,255,255,0.08)"
                                : "transparent",
                            color: "#fff",
                          }}
                        >
                          <div style={{ fontWeight: 700 }}>
                            Room {thread.roomId}
                          </div>
                          <div
                            style={{
                              fontSize: "0.82rem",
                              opacity: 0.72,
                              whiteSpace: "nowrap",
                              overflow: "hidden",
                              textOverflow: "ellipsis",
                            }}
                          >
                            {thread.lastMessage?.content ||
                              thread.lastMessage?.type ||
                              "Room event"}
                          </div>
                        </button>
                      ))
                    )}
                  </div>

                  <div
                    style={{
                      minHeight: 0,
                      overflowY: "auto",
                      padding: "1rem",
                      display: "grid",
                      gap: "0.65rem",
                      alignContent: "start",
                      background:
                        "linear-gradient(180deg, rgba(255,255,255,0.02), rgba(255,255,255,0.01))",
                    }}
                  >
                    {activeRoomMessages.length === 0 ? (
                      <div style={{ color: "rgba(255,255,255,0.65)" }}>
                        Open a room to view messages.
                      </div>
                    ) : (
                      activeRoomMessages.map((message, index) => (
                        <div
                          key={`${message.room_id}-${message.direction}-${index}-${message.content}`}
                          style={{
                            justifySelf:
                              message.direction === "outgoing"
                                ? "end"
                                : "start",
                            maxWidth: "82%",
                            padding: "0.8rem 0.9rem",
                            borderRadius:
                              message.direction === "outgoing"
                                ? "16px 16px 6px 16px"
                                : "16px 16px 16px 6px",
                            background:
                              message.direction === "outgoing"
                                ? "linear-gradient(135deg, #5b8cff, #6ee7ff)"
                                : "rgba(255,255,255,0.08)",
                            color: "#fff",
                            lineHeight: 1.38,
                          }}
                        >
                          <div
                            style={{
                              fontSize: "0.76rem",
                              opacity: 0.8,
                              marginBottom: "0.3rem",
                            }}
                          >
                            {message.direction === "outgoing"
                              ? "You"
                              : message.sender_name ||
                                message.senderName ||
                                `User ${message.sender_id}`}
                          </div>
                          <div>
                            {message.content ||
                              (message.room_status
                                ? "joined the room"
                                : "left the room")}
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                  <div
                    style={{
                      padding: "1rem",
                      borderTop: "1px solid rgba(255,255,255,0.08)",
                      display: "grid",
                      gap: "0.75rem",
                    }}
                  >
                    <textarea
                      placeholder={
                        activeRoomId
                          ? "Write to room..."
                          : "Open a room first..."
                      }
                      value={roomDraft}
                      onChange={(e) => setRoomDraft(e.target.value)}
                      rows={3}
                      disabled={!activeRoomId}
                      style={{
                        width: "100%",
                        padding: "0.85rem",
                        borderRadius: "12px",
                        border: "1px solid rgba(255,255,255,0.12)",
                        background: "#111",
                        color: "#fff",
                        resize: "vertical",
                        opacity: activeRoomId ? 1 : 0.6,
                      }}
                    />
                    <div
                      style={{
                        display: "flex",
                        gap: "0.75rem",
                        alignItems: "center",
                      }}
                    >
                      <button
                        className="profile_button"
                        onClick={() => {
                          if (
                            !currentUserId ||
                            !activeRoomId ||
                            !roomDraft.trim()
                          )
                            return;
                          const payload = {
                            sender_id: currentUserId,
                            reciever_id: 0,
                            room_id: activeRoomId,
                            room_status: false,
                            type: "room_msg",
                            content: roomDraft.trim(),
                            post_id: 0,
                          };
                          // optimistic UI
                          setRoomMessages((prev) => [
                            ...prev,
                            {
                              ...payload,
                              direction: "outgoing",
                              room_id: activeRoomId,
                            },
                          ]);
                          if (sendNotifications) sendNotifications(payload);
                          setRoomDraft("");
                        }}
                        disabled={!activeRoomId}
                      >
                        Send
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {notification && (
              <div
                style={{
                  position: "fixed",
                  top: "1rem",
                  right: "1rem",
                  zIndex: 9999,
                  minWidth: "280px",
                  maxWidth: "360px",
                  padding: "0.9rem 1rem",
                  borderRadius: "14px",
                  background: "rgba(18, 18, 18, 0.96)",
                  color: "#fff",
                  boxShadow: "0 18px 50px rgba(0,0,0,0.35)",
                  border: "1px solid rgba(255,255,255,0.08)",
                }}
              >
                <div
                  style={{
                    fontSize: "0.85rem",
                    opacity: 0.8,
                    marginBottom: "0.35rem",
                  }}
                >
                  New notification recieved
                </div>
                <div style={{ fontWeight: 200, marginBottom: "0.25rem" }}>
                  {notification.type}
                </div>
              </div>
            )}

            <Header />
            <div className="Site_content">
              <div className="Site_sidebar">
                <Sidebar />
              </div>
              <main>
                <Outlet />
              </main>
            </div>

            <Footer />
          </section>
        </postDataContext.Provider>
      </RealtimeContext.Provider>
      <WebSocketDebug token={token} />
    </>
  );
}

export const usePostContext = () => useContext(postDataContext);
