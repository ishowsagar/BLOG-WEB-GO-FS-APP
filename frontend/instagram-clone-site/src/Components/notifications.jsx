import DashboardCard from "./notification_dashboard_card";
import { useOutletContext } from "react-router-dom";

// ** PRODUCTION TODOS **//
// 1. replace wsConnector handler url to use production api url
// 2. instead of manual sends,just let notifications come here from pns routed only
// 3. Link notification to posts later to open that notification post
// 4. refine ui later
// ** end ** //

// ** Only for local development and testing
const token = localStorage.getItem("token");

export default function NotificationComponent() {
  // fetching shared state from outlet context
  const { globalNotification } = useOutletContext();

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
      {globalNotification.length > 0 ? (
        <div style={{ color: "white", fontSize: "47px", marginTop: "20px" }}>
          <p>
            Notification Count :
            <span style={{ color: "red", fontWeight: "900" }}>
              {` ${globalNotification.length}`}
            </span>
          </p>
        </div>
      ) : (
        <div style={{ color: "white", fontSize: "47px", marginTop: "20px" }}>
          <p>No new notification...</p>
        </div>
      )}
      <p className="ws_hero_text">
        client online🟢
      </p>
      <div className="dashboard-feed-wrapper">
        {/* Dynamic List Render Loop */}
        {globalNotification.length > 0 ? (
          globalNotification.map((notifyPayload, index) => (
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
    </div>
  );
}
