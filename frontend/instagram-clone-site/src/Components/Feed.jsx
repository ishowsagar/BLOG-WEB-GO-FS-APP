import FeedPost from "./FeedPost";
import { usePostContext } from "../Layout/MainLayout";
import "./Feed.css";
export default function Feed() {
  // todo - rendering each Post

  const { postBatch, bottomRef } = usePostContext();

  if (!postBatch || !Array.isArray(postBatch)) {
    return (
      <div style={{ color: "#999", textAlign: "center", padding: "2rem" }}>
        Loading feed...
      </div>
    );
  }

  const storiesElements = postBatch.map((post) => {
    // const displayUsername = post.username
    //   ? post.username
    //   : `user${post.userId}`;
    return (
      <div className="story" key={post.id}>
        <img
          src={`https://i.pravatar.cc/60?u=${post.id} `}
          className="story_avatar"
        />
        {/* <span className="story_username">{displayUsername}</span> */}
      </div>
    );
  });
  return (
    <>
      {/* rednering each post */}
      <div className="Site_stories_wrapper">{storiesElements}</div>
      <FeedPost />
      {/* observer observes on ref, and observing where ref is being attached, means this div */}
      <div
        style={{
          marginLeft: "40px",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          padding: "2rem",
        }}
        ref={bottomRef}
      >
        <div className="fidget-spinner"></div>
      </div>
    </>
  );
}
