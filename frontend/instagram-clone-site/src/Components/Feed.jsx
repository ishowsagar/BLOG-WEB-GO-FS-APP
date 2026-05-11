import FeedPost from "./FeedPost";
import { usePostContext } from "../Layout/MainLayout";
export default function Feed() {
  // todo - rendering each Post

  const { postData } = usePostContext();

  const storiesElements = postData.map((post) => {
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
    </>
  );
}
