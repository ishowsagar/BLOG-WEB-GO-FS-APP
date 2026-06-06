import React from "react";
import { Link } from "react-router-dom";

export default function DashboardCard({
  senderID,
  senderName,
  content,
  recieverID,
  recieverName,
  type,
  postID,
}) {
  const displaySenderName = senderName || `User ${senderID || "Unknown"}`;
  
  // Get first letter of sender name for avatar placeholder
  const avatarLetter = displaySenderName.charAt(0).toUpperCase();

  // Generate human-readable message depending on notification type
  const getNotificationMessage = () => {
    switch (type) {
      case "like_posted":
        return "liked your post.";
      case "comment_posted":
        return `commented: "${content || "Nice post!"}"`;
      case "follow_posted":
        return "started following you.";
      case "dm":
        return `sent you a message: "${content || ""}"`;
      default:
        return content || "sent you a notification.";
    }
  };

  // Render action button on the right side
  const renderAction = () => {
    if (postID && postID !== 0) {
      return (
        <Link to={`/feed/${postID}`} className="notification-row-action-btn">
          View Post
        </Link>
      );
    }
    if (type === "follow_posted") {
      return (
        <Link to="/profile" className="notification-row-action-btn follow-btn">
          Profile
        </Link>
      );
    }
    if (type === "dm") {
      return (
        <Link to="/messages" className="notification-row-action-btn message-btn">
          Chat
        </Link>
      );
    }
    return null;
  };

  return (
    <div className="notification-row-item">
      <div className="notification-row-left">
        {/* Instagram-style circular avatar placeholder */}
        <div className="notification-avatar">
          {avatarLetter}
        </div>
        
        {/* Notification Text */}
        <div className="notification-row-text">
          <span className="notification-sender-name">{displaySenderName}</span>
          {" "}
          <span className="notification-action-text">{getNotificationMessage()}</span>
        </div>
      </div>

      {/* Action target link on the right */}
      <div className="notification-row-right">
        {renderAction()}
      </div>
    </div>
  );
}
