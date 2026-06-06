import { useContext, useEffect, useRef, useState } from "react";
import { jwtDecode } from "jwt-decode";
import DashboardCard from "./notification_dashboard_card";
import { wsUrl } from "../Services/apiConfig";
import { RealtimeContext } from "../Layout/MainLayout";

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

  const { subscribeNotifications } = useContext(RealtimeContext)

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



  useEffect(() => {
    const unsubscribe = subscribeNotifications((parsedPayload) => {

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
      return () => unsubscribe()
    })
  }, [subscribeNotifications])


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
        {hasWsConnEstablished
          ? `client online🟢`
          : "connecting to the notification service..."}
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
