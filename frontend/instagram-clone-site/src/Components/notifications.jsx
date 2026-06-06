import { useEffect, useRef, useState } from "react";
import { jwtDecode } from "jwt-decode";
import DashboardCard from "./notification_dashboard_card";
import { wsUrl } from "../Services/apiConfig";

// ** PRODUCTION TODOS **//
// 1. replace wsConnector handler url to use production api url
// 2. instead of manual sends,just let notifications come here from pns routed only
// 3. Link notification to posts later to open that notification post
// 4. refine ui later
// ** end ** //

// ** Only for local development and testing
// uncomment prod apiUrl to use that, and uncomment this to remove this
const token = localStorage.getItem("token");

export default function NotificationComponent() {
  const [hasWsConnEstablished, setHasWsConnEstablished] = useState(false); // conditional for tracking wsConn when opened/closed
  const [hasWsConnHitErr, setHasWsConnHitErr] = useState(false); // conditional for tracking wsConn when closed abruptly or way it closed the conn
  const [writerResponse, SetWriterResponse] = useState([]); // for setting data in the arr when recieved from the client's writer's response
  const [notificationOfTypeDM, setNotificationOfTypeDM] = useState(false); // for conditionally rendering div if this becomes true

  const wsConnHolderRef = useRef(null); //* would be an holder for "wsConn"
  // & States
  const DEVELOPMENT_API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  // ** end //

  if (token === "") {
    console.log("login expired or token not found ");
    return;
  }

  // ! since we are no longer sending id from here manually, our handler extracts token from query and
  //  extract token str and by parsing, retrieves the userID and attches it to the client
  //   testing - explicityly sending dynamic userID for reader's payload check
  const decodedToken = jwtDecode(token);
  //   gives us - decodedToken {expiry: '2026-06-04T07:37:50.960431745Z', user_id: 16}
  // console.log("decodedToken", decodedToken);

  const senderID = decodedToken.user_id;
  const recieverID = senderID === 41 ? 16 : 41;
  const sender_name = senderID === 16 ? "brave" : "ronaldo";
  const reciever_name = recieverID === 16 ? "brave" : "ronaldo";

  //** this fixed one thing -> now dynamically setting senderID,not like for every client it do trigger send */
  // bug - must include all fields or else it will fail on consumer routing,
  // fixed - now carries full payload
  //   {
  // "sender_id": 16,
  // "reciever_id": 41,
  // "sender_name": "brave",
  // "reciever_name": "ronaldo",
  // "type": "dm",
  // "room_id" : 0,
  // "content": "hey ronaldo wassup man!",
  // "post_id": 0
  // }
  const likeNotifyPayload = {
    // payload of type NotifyPayload that handler expects
    room_id: 0,
    sender_id: senderID, //orton
    reciever_id: recieverID, //instaUser
    sender_name: "ronaldo",
    reciever_name: "brave",
    type: "like_posted", // most imp for diffrentiation of payloads and routing purpose
    content: `${sender_name} has liked your post`,
    post_id: 1,
  };

  const commentNotifPayload = {
    // payload of type NotifyPayload that handler expects
    room_id: 0,
    sender_id: senderID, //orton
    reciever_id: recieverID, //instaUser
    sender_name: senderID === 16 ? "brave" : "ronaldo",
    reciever_name: recieverID === 16 ? "brave" : "ronaldo",
    type: "like_posted", // most imp for diffrentiation of payloads and routing purpose
    content: `${sender_name} commented 'nice post bro'`,
    post_id: 11,
  };
  const jsonNotificationPayload = JSON.stringify(commentNotifPayload);
  // it would be where handler is intercepting requests - ws://localhost:8080/api/ws

  //  bug - have to strip out http:// <- this would crash app
  //   fixed - it mounts again only if connStr changes + clean url without 'http' marka
  // const cleanBaseURL = DEVELOPMENT_API_BASE_URL.replace(/^https?:\/\//, ""); // \ \ for espacing and using them nested inside
  const wsConnURlString = token
    ? `wss://${window.location.host}/api/ws?token=${encodeURIComponent(token)}`
    : null;
  console.log(wsConnURlString);
  //* 1- mouting ws connection instance - handler expects conn on route path -"/api/ws/" , ~/dm for dms which gives pesrsistent messages
  useEffect(() => {
    // since its not a http request, i can't send token in header for token verification, would have to send via encoded url utility in the queryParam
    if (!wsConnURlString) return;
    const webSocketConn = new WebSocket(wsConnURlString);
    wsConnHolderRef.current = webSocketConn; //! assigning and stored connection in 'ref's current' state

    //* 2- mouting all 4 handlers~interceptors

    // & Transmitter
    webSocketConn.onopen = () => {
      // Once conn is on, we could send payload to the client's ws conn handled by the handler
      console.log(
        // console.log("ws connection established on", wsConnURlString);
        "ws connection is alive and ported for bidirectional information exchange is in action⚡.",
      );
      //** ONCE BIDIRECTIONAL WS CONNECTION PORTAL IS OPENED FOR ACTIVE CLIENT, IT CAN TALK TO SERVER **/
      //@ wsconnection instance has send for sending payloads to the server
      //   webSocketConn.send(jsonNotificationPayload);
      setHasWsConnEstablished(true);
    };

    // & Reciever (eventObj)
    webSocketConn.onmessage = (recievedPayload) => {
      // note - if you try to log recievedRes/generalRes directly accessing its field -> undefined -> unparsed stringified data could not be retrieved without parsing
      console.log("payload intercepted");
      try {
        if (recievedPayload.data) {
          const parsedPayload = JSON.parse(recievedPayload.data);

          // since its parsed,we can put field checks on it, so only stores payload of type "dm"
          if (
            // todo - intercept p2p actual 'notification' , of these types
            // todo - 1. must send exclusive payloads and render correctly in the card
            parsedPayload.type === "like_posted" ||
            parsedPayload.type === "follow_posted" ||
            parsedPayload.type === "comment_posted" ||
            parsedPayload.type === "dm"
          ) {
            SetWriterResponse((prevArrData) => [...prevArrData, parsedPayload]); // saving 'data' in state that writer would have responded with
            console.log(
              "intercepted incoming payload of type -",
              parsedPayload.type,
              "payload -",
              parsedPayload,
            );
            setNotificationOfTypeDM(true); //* setting it true that we recieved tyoe payloa
          }
        } else {
          throw new Error("only supported payload of type 'dm' ");
        }
      } catch (err) {
        // ! throws err to the err interceptor handler -> works in sync
        console.error(err);
        setHasWsConnHitErr(true);
      }
    };

    // & Closer -> safely closes ws connection from this end
    webSocketConn.onclose = () => {
      console.log("client disconnected;closing webSocket connection portal");
      setHasWsConnEstablished(false); // now conn is closed, so setting it false altering that conn is closed
    };

    // & Err interceptor
    webSocketConn.onerror = (err) => {
      console.log("unexpected error occured", err);
      setHasWsConnHitErr(true);
    };

    // useEffect cleaner fires up when needed to close ws conn for clean up purpose ; cleans up unhandled ws conn closed abnormally
    return () => {
      if (
        webSocketConn.readyState === WebSocket.OPEN ||
        webSocketConn.readyState === WebSocket.CONNECTING
      ) {
        // % if wsConn current state matches these conditions, gracefully close the connection like we do in client's reader
        webSocketConn.close();
      }
    };
  }, [wsConnURlString]); // remount only if that changes, that would depend on token

  // todo - Now its time to implement webSocket connection here
  // todo - render incoming notification data inside these fillers
  // const [posts, setPosts] = useState([
  //   {
  //     id: "usr_node_01X",
  //     name: "Administrator_System",
  //     body: "Database migration pipeline completed successfully for production nodes.",
  //     reference: "pkt_99210_db",
  //     badge: "system normal",
  //   },
  //   {
  //     id: "usr_node_88B",
  //     name: "Gateway_Service",
  //     body: "WebSocket connection established handshakes with 14 active browser sessions.",
  //     reference: "pkt_99211_ws",
  //     badge: "live status",
  //   },
  // ]);

  // const testData = [
  //   {
  //     senderId: "usr_99_alpha",
  //     senderName: "helloworld",
  //     content:
  //       "Title - Hello, world! Testing out the new real-time WebSocket architecture on denvergram.me.",
  //     postId: "pkt_101_init",
  //     rightElement: "Group Chat",
  //   },
  // ];

  // handles on-click sending payload via ref which -> stores wsConn Instance
  function handleSendNotificationPayload(notificationPayload) {
    wsConnHolderRef.current.send(notificationPayload); // calling sender via ref.current holder
    console.log("sent payload to the handler", notificationPayload);
  }

  // fallback ui
  if (!token || token === "") {
    return (
      <div style={{ color: "white", padding: "20px" }}>
        <h1>login expired or invalid token. Please login again.</h1>
      </div>
    );
  }

  return (
    <div className="dashboard-feed-viewport">
      {writerResponse.length > 0 ? (
        <div style={{ color: "white", fontSize: "47px", marginTop: "20px" }}>
          <p>
            Notification Count :
            <span style={{ color: "red", fontWeight: "900" }}>
              {` ${writerResponse.length}`}
            </span>
          </p>
        </div>
      ) : (
        <div style={{ color: "white", fontSize: "47px", marginTop: "20px" }}>
          <p>No new notification...</p>
        </div>
      )}
      <p className="ws_hero_text">
        {hasWsConnEstablished ? `client online🟢` : "client offline🔴"}
      </p>
      {/* <button
        disabled={!hasWsConnEstablished}
        onClick={() => handleSendNotificationPayload(jsonNotificationPayload)}
        style={{
          padding: "4px 6px",
          fontWeight: "bolder",
          fontSize: "17px",
          cursor: "pointer",
          borderRadius: "7px",
          boxShadow: "skyblue -3px 3.8px",
        }}
      >
        Send Notification
      </button> */}
      {/* only conditionally render notification if they are of type 'dm' */}
      {notificationOfTypeDM && (
        <div className="dashboard-feed-wrapper">
          {/* Dynamic List Render Loop */}
          {writerResponse.length > 0 ? (
            writerResponse.map((notifyPayload, index) => (
              <DashboardCard
                key={index}
                senderID={notifyPayload.sender_id}
                senderName={notifyPayload.sender_name}
                content={notifyPayload.content}
                recieverName={notifyPayload.reciever_name}
                recieverID={notifyPayload.reciever_id}
                type={notifyPayload.type}
                postID={notifyPayload.post_id}
              />
            ))
          ) : (
            <p className="feed-empty-state">
              No live logs or active posts available.
            </p>
          )}
        </div>
      )}
    </div>
  );
}
