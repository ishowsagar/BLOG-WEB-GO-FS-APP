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
  return (
    <div className="dashboard-card">
      <div className="card-layout">
        {/* LEFT SIDE: Vertical Data Stack */}
        <div className="card-left-stack">
          <span className="card-label-id">{`senderID : ${senderID}`}</span>
          <h3 className="card-title-name">{`SenderName : ${senderName}`}</h3>
          <p className="card-body-content">{`${content}`}</p>
          <span className="card-label-post">{`type : ${type}`}</span>
        </div>

        {/* RIGHT SIDE: Action or position marker */}
        <div className="card-right-element">{`Me : ${recieverName}`}</div>
        <Link to={`/feed/${postID}`}>
          <div className="card-right-element">{`Go to Post_ID : ${postID}`}</div>
        </Link>
      </div>
    </div>
  );
}
