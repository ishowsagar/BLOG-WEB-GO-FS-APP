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
      <div className="notifications-unauthenticated">
        <h1>Login expired or invalid token. Please login again.</h1>
      </div>
    );
  }

  return (
    <div className="notifications-viewport">
      <div className="notifications-container">
        <div className="notifications-header">
          <h1 className="notifications-title">Notifications</h1>
          {globalNotification.length > 0 && (
            <span className="notifications-badge-count">
              {globalNotification.length} new
            </span>
          )}
        </div>

        <div className="notifications-list">
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
            <div className="notifications-empty-state">
              <div className="notifications-empty-icon">🔔</div>
              <p className="notifications-empty-text">No new notifications</p>
              <p className="notifications-empty-subtext">
                When you get likes, comments, or followers, they will show up here.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
