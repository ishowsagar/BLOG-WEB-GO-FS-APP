import { useEffect, useState } from "react";
import DashboardCard from "./notification_dashboard_card";
import { wsUrl } from "../Services/apiConfig";
import { useOutletContext } from "react-router-dom"
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

  // fetching shared state from outlet context
  const { globalNotification, setGlobalNotifications } = useOutletContext();

  // const wsConnHolderRef = useRef(null); //* would be an holder for "wsConn"
  // & States
  // const apiUrl = import.meta.env.VITE_API_BASE_URL;


  if (token === "") {
    console.log("login expired or token not found ");
    return;
  }

  useEffect(() => {
    console.log("notification component is mounted.")

    const wsConnUrl = `${wsUrl("/api/ws")}?token=${encodeURIComponent(token)}`
    const ws = new WebSocket(wsConnUrl)

    //& loading all four interceptors
    ws.onopen = () => {
      console.log("connection is successfully established on the notifications page")
      console.log("wsConnUrl - ", wsConnUrl)
      setHasWsConnEstablished(true)
    }

    // notification reciever interceptor
    ws.onmessage = (event) => {
      console.log("notification event data recieved")

      try {

        const notification_event_payload = JSON.parse(event.data)
        console.log("parsed notification event payload : ", notification_event_payload)

        const allowedEventDataTypes = {
          "like_posted": true,
          "comment_posted": true,
          "follow_posted": true,
          "dm": true
        }

        // pass only allowed types
        if (!allowedEventDataTypes[notification_event_payload.type]) {
          console.log("wrong notification data type")
          throw new Error("invalid notification data type")
        }


        // if succesfully parsed data into js object type of format and data type is matched
        setHasWsConnEstablished(true)
        SetWriterResponse((prevDataArr) => [...prevDataArr, notification_event_payload])
        setGlobalNotifications((prev) => [...prev, notification_event_payload])
      } catch (err) {
        setHasWsConnHitErr(true)

        console.error(err)
      }
    }

    ws.onclose = () => {
      console.log("ws connection has been closed")
    }

    // any error - ws closed with an error
    ws.onerror = () => {
      setHasWsConnHitErr(true)
      setHasWsConnEstablished(false)
      console.log("ws connection error")
    }

    return () => {
      // close websocket connection when component unmounts
      ws.close()
    }
  }, [token])

  // ! since we are no longer sending id from here manually, our handler extracts token from query and
  //  extract token str and by parsing, retrieves the userID and attches it to the client
  //   testing - explicityly sending dynamic userID for reader's payload check
  // const decodedToken = jwtDecode(token);
  //   gives us - decodedToken {expiry: '2026-06-04T07:37:50.960431745Z', user_id: 16}
  // console.log("decodedToken", decodedToken);

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
      {/* only conditionally render notification if they are of type 'dm' */}
      {hasWsConnEstablished && (
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
