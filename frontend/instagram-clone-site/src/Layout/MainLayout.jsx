import Header from "../Components/Header";
import Footer from "../Components/Footer";
import Sidebar from "../Components/Sidebar";
import { Outlet, useNavigate } from "react-router-dom";
import { useEffect, useState, createContext, useContext, useRef } from "react";
import { useWebSocket } from "../hooks/useWebSocket";
const postDataContext = createContext();

export default function MainLayout() {
  // window.location.reload();
  const [postData, setPostData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [notification, setNotification] = useState(null);
  const [showDmInbox, setShowDmInbox] = useState(false);
  const [dmMessages, setDmMessages] = useState([]);
  const [activeDmPeerId, setActiveDmPeerId] = useState(null);
  const [dmDraft, setDmDraft] = useState("");
  const [dmInboxStatus, setDmInboxStatus] = useState("");
  const [postBatch, SetPostBatch] = useState([]); // for storing batches
  const [cursor, setCursor] = useState(""); // for cursor detection -> what was last post ID - triggering next batch fetch
  const [nextBatchLoading, setNextBatchLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true); // telling it still has posts,don't trigger next batch load
  const [feedLoadFailed, setFeedLoadFailed] = useState(false);
  const bottomRef = useRef(null); // for bottom scroll detection
  const dmThreadRef = useRef(null);
  // ** PAGINATION ** //
  // batch req url - api/feed/batch?
  // queryParams := limit=?&nextCursor=?
  // idea - when cursor fliped true-> conditionally trigger render of next batch, append batch to array which holds sequential post data

  if (loading && !hasMore) {
    console.error("can't load more feed");
    return;
  }

  // token from client header
  const token = localStorage.getItem("token");
  console.log(token);

  // WebSocket for real-time notifications
  const { subscribe: subscribeNotifications } = useWebSocket(token);
  const { subscribe: subscribeDm, send: sendDm } = useWebSocket(
    token,
    "/api/ws/dm",
  );

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

  // Subscribe to WebSocket notifications
  useEffect(() => {
    let hideTimer = null;

    const unsubscribe = subscribeNotifications((notification) => {
      console.log("🔔 New notification:", notification);
      setNotification(notification);

      if (hideTimer) {
        clearTimeout(hideTimer);
      }

      hideTimer = setTimeout(() => {
        setNotification(null);
      }, 5000);

      // Handle different notification types
      switch (notification.type) {
        case "post_created":
          console.log(
            `📝 New post from user ${notification.sender_id}:`,
            notification.content,
          );
          break;
        case "like_posted":
          console.log(
            `❤️ User ${notification.sender_id} liked post ${notification.post_id}`,
          );
          break;
        case "comment_posted":
          console.log(
            `💬 User ${notification.sender_id} commented on post ${notification.post_id}:`,
            notification.content,
          );
          break;
        default:
          console.log("Unknown notification type:", notification.type);
      }
    });

    return () => {
      if (hideTimer) {
        clearTimeout(hideTimer);
      }
      unsubscribe();
    };
  }, [subscribeNotifications]);

  useEffect(() => {
    if (!token || !currentUserId) return undefined;

    const unsubscribe = subscribeDm((incomingMessage) => {
      if (incomingMessage.type !== "dm") return;

      const senderId = Number(incomingMessage.sender_id);
      const receiverId = Number(incomingMessage.reciever_id);
      const peerId = senderId === currentUserId ? receiverId : senderId;

      const normalizedMessage = {
        ...incomingMessage,
        peer_id: peerId,
        direction: senderId === currentUserId ? "outgoing" : "incoming",
      };

      setDmMessages((prev) => [...prev, normalizedMessage]);
      setDmInboxStatus(
        senderId === currentUserId ? "message sent" : "new message received",
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
    if (!accumulator[peerId]) {
      accumulator[peerId] = [];
    }
    accumulator[peerId].push(message);
    return accumulator;
  }, {});

  const dmThreadEntries = Object.entries(dmThreads)
    .map(([peerId, messages]) => ({
      peerId: Number(peerId),
      messages,
      lastMessage: messages[messages.length - 1],
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

  // batch request meta
  const batchReq = {
    // url depends on cursor -> if it has last batch post ID -> if yes with cursor&limit otherwise just with limit
    // "GET" - /api/feed?limit=X{limitOffset}&cursor=Y{lastCursor} - return feed[],cursor,hasMore
    url: cursor
      ? `http://localhost:8080/api/feed/batch?limit=4&nextCursor=${cursor}`
      : `http://localhost:8080/api/feed/batch?limit=4`,
    header: {
      Authorization: token,
    },
    method: "GET",
  };

  // load batch according to the cursor and limit offset
  async function LoadBatchesFeed() {
    console.log("batch feed is loading...");
    if (feedLoadFailed || nextBatchLoading || !hasMore) {
      return;
    }
    setNextBatchLoading(true);
    try {
      if (!token) {
        return;
      }
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
        throw new Error(response.Status || "failed to load batch feed");
      }

      console.log(response);

      // batch fetched
      SetPostBatch((prevBatch) => [...prevBatch, ...response.Batch]); //& stores data in postBatch state

      // backend is fetching response as desc order so lastPost id means first, also its fetching from downwards, so limit kept decreasing it by 4, so id<40,batch -4 => id < 4-4
      // next batch req -> server sends last postID -> sets in url -> ask for next batch => which means ask for results entries where id < 40-lastBatchEntryCount => batch entries id < 36 -> kept going
      setCursor(response.NextCursor); // cursor's -> last post id is fetched from go server resonse struct -> which is used to set cursor in url to fetch the next batch round
      setHasMore(response.HasMore);
      console.log("batch loaded");
    } catch (err) {
      console.log(err);
    } finally {
      // setting loading to false after everything is loaded in batch and all set
      setLoading(false);
      setNextBatchLoading(false);
    }
  }

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }

    LoadBatchesFeed();
  }, [token]);

  // observer
  // observer for client's scroll - useEffect mounts on site loads
  useEffect(() => {
    if (!bottomRef.current) return;
    // Browser API which tells -> if client has left the described viewport
    const observer = new IntersectionObserver((entries) => {
      // if entries arr's first el is in interaction
      //  res comes in desc order -> that's why first == actually last
      const firstEntryPost = entries[0]; // array's first el at 0th index is 1st el
      if (
        entries[0].isIntersecting &&
        hasMore &&
        !nextBatchLoading &&
        !feedLoadFailed
      ) {
        // *so if last el is in interation of the client
        //$ load more data
        LoadBatchesFeed(); // new batch will be fetched
      }
    });

    // observer observes there where this ref is attached
    observer.observe(bottomRef.current); //* this will be observed where this ref is attached and referenced to

    // return as cleanup function
    return () => observer.disconnect();
  }, [hasMore, cursor, nextBatchLoading, feedLoadFailed]);

  // todo - API service repo method which -> fetch data in batches using QParams
  // fixed - added method to fetch batch requested data

  // end //

  // useEffect(() => {
  //   console.log("MainLayout token:", token);
  //   if (!token) {
  //     // token missing: stop loading and allow UI to render (or redirect to login)
  //     console.warn("No token found in localStorage (key: 'token')");
  //     setLoading(false);
  //     return;
  //   }
  //   const payloadBody = {
  //     url: "http://localhost:8080/api/feed",
  //     Method: "GET",
  //     //* this is working now, it is fetching token from local storage
  //     AuthorizationHeaderToken: token,
  //   };
  //   async function fetchData() {
  //     try {
  //       const response = await fetch(payloadBody.url, {
  //         method: payloadBody.Method,
  //         headers: { Authorization: payloadBody.AuthorizationHeaderToken },
  //       });
  //       const data = await response.json();
  //       console.log("feed status:", response.status, data);
  //       console.log(payloadBody.AuthorizationHeaderToken);
  //       // const returnPostIfExists = Array.isArray(postData) ? postData : []
  //       if (!response.ok) {
  //         throw new Error(data?.Status || data?.error || "error fetching data");
  //       }
  //       setPostData(data.Post);
  //       setLoading(false);

  //       //  reload page, fetch current url from window{browser}.Location
  //     } catch (err) {
  //       console.log(err);
  //     } finally {
  //       setLoading(false);
  //     }
  //   }
  //   fetchData();
  // }, [token]);

  if (loading) {
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
      <postDataContext.Provider value={{ postBatch, SetPostBatch, bottomRef }}>
        <section className="Site_wrapper">
          <button
            onClick={() => setShowDmInbox((prev) => !prev)}
            style={{
              position: "fixed",
              bottom: "1rem",
              right: "1rem",
              zIndex: 10000,
              padding: "0.85rem 1rem",
              borderRadius: "999px",
              border: "none",
              background: "linear-gradient(135deg, #ff4d9d, #ff8a5b)",
              color: "#fff",
              fontWeight: 700,
              boxShadow: "0 16px 40px rgba(0,0,0,0.35)",
            }}
          >
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
                          User {thread.peerId}
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
                      ? `Chat with User ${activeDmPeerId}`
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
                            lineHeight: 1.45,
                          }}
                        >
                          <div
                            style={{
                              fontSize: "0.76rem",
                              opacity: 0.78,
                              marginBottom: "0.25rem",
                            }}
                          >
                            {message.direction === "outgoing"
                              ? "You"
                              : `User ${message.sender_id}`}
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
                      gap: "0.65rem",
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
                        <span style={{ color: "#9eff9e", fontSize: "0.9rem" }}>
                          {dmInboxStatus}
                        </span>
                      )}
                    </div>
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
                New notification recieved📩
              </div>
              <div style={{ fontWeight: 200, marginBottom: "0.25rem" }}>
                {/* //* might need to fetch user by that userID to display its name and all */}
                <span
                  style={{
                    color: "red",
                    fontFamily: "fantasy",
                    fontSize: "17px",
                  }}
                >
                  {" "}
                  {notification.type === "like_posted" &&
                    `user${notification.sender_id} liked your post`}
                </span>
                <span
                  style={{
                    color: "blueviolet",
                    fontFamily: "fantasy",
                    fontSize: "17px",
                  }}
                >
                  {notification.type === "comment_posted" &&
                    `user${notification.sender_id} commented on your post`}
                </span>
                {notification.type === "post_created" && "New post created"}
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
    </>
  );
}

// export default function MainLayout() {
//   // window.location.reload();
//   const [postData, setPostData] = useState([]);
//   const [loading, setLoading] = useState(true);

//   // const redirectToLoginPageIfTokenNotFound = useNavigate()

//   // bug - it was a bug before when we user logged in and feed was giving token mismatch erros
//   // fixed by normally fetching token here from ls(local storage) and only setting postData once req was ok not before checking this and reload if token chances to reload flow for rendering feed
//   // * get token from localstorage
//   const token = localStorage.getItem("token");

//   // * remember - on fresh login,token gets updated too

//   useEffect(() => {
//     console.log("MainLayout token:", token);
//     if (!token) {
//       // token missing: stop loading and allow UI to render (or redirect to login)
//       console.warn("No token found in localStorage (key: 'token')");
//       setLoading(false);
//       return;
//     }
//     const payloadBody = {
//       url: "http://localhost:8080/api/feed",
//       Method: "GET",
//       //* this is working now, it is fetching token from local storage
//       AuthorizationHeaderToken: token,
//     };
//     async function fetchData() {
//       try {
//         const response = await fetch(payloadBody.url, {
//           method: payloadBody.Method,
//           // ! this is how you send token in header from client frontend
//           headers: { Authorization: payloadBody.AuthorizationHeaderToken },
//         });
//         const data = await response.json();
//         console.log("feed status:", response.status, data);
//         console.log(payloadBody.AuthorizationHeaderToken);
//         // const returnPostIfExists = Array.isArray(postData) ? postData : []
//         if (!response.ok) {
//           throw new Error(data?.Status || data?.error || "error fetching data");
//         }
//         setPostData(data.Post);
//         setLoading(false);

//         //  reload page, fetch current url from window{browser}.Location
//       } catch (err) {
//         console.log(err);
//       } finally {
//         setLoading(false);
//       }
//     }
//     fetchData();
//   }, [token]);

//   if (loading) {
//     return (
//       <div
//         style={{ textAlign: "center", marginTop: "4rem", fontSize: "1.5rem" }}
//       >
//         Loading...
//       </div>
//     );
//   }
//   return (
//     <>
//       <postDataContext.Provider value={{ postData, setPostData }}>
//         <section className="Site_wrapper">
//           <Header />
//           <div className="Site_content">
//             <div className="Site_sidebar">
//               <Sidebar />
//             </div>
//             <main>
//               <Outlet />
//             </main>
//           </div>

//           <Footer />
//         </section>
//       </postDataContext.Provider>
//     </>
//   );
// }
export const usePostContext = () => useContext(postDataContext);
