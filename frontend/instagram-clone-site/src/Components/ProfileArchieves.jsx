import React from "react";
import demo1 from "../assets/demo1.jpg";
import demo2 from "../assets/demo2.jpg";
export default function ProfileArchieves() {
  // Free online image for demo
const postImg2 = "https://picsum.photos/id/1025/400/400";

  return (
    <div className="profile_posts_grid">
      <img src={demo1} alt="Post 1" className="profile_post_img" />
      <img src={demo2} alt="Post 2" className="profile_post_img" />
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "300px",
        }}
      >
        <img
          src={postImg2}
          alt="Archive"
          style={{
            width: "220px",
            height: "220px",
            borderRadius: "16px",
            objectFit: "cover",
            boxShadow: "0 2px 12px rgba(0,0,0,0.18)",
          }}
        />
      </div>
    </div>
  );
}
