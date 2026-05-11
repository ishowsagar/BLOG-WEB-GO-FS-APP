import { Link } from "react-router-dom";
import { usePostContext } from "../Layout/MainLayout";

export default function FeedPost() {
  const { postData } = usePostContext();
  console.log(postData[0]); // data is coming i have confirmed it
  const postElements = postData.map((post) => {
    console.log("post clicked");
    return (
      <div key={`${post.id}`} className="feedpost_container-layout">
        {/* feed/:id -> set by the router as id being param1  */}
        <Link to={`/feed/${post.id}`}>
          <div className="feedpost">
            {/* Post preview */}
            <div className="feedpost_header">
              <img
                className="feedpost_avatar"
                src={`https://i.pravatar.cc/150?img=${post.user_id}`}
                alt="avatar"
              />
              <span className="feedpost_username">user-{post.user_id}</span>
              <span className="feedpost_time">{post.created_at}</span>
            </div>
            <img
              className="feedpost_image"
              src={`https://picsum.photos/seed/${post.id}/500/350`}
              alt="post"
            />
            <div className="feedpost_caption">
              <h2 className="feedpost_title">{post.title}</h2>
              <p className="feedpost_body">{post.content}</p>
            </div>
          </div>
        </Link>
      </div>
    );
  });
  return (
    // map over to render eachPost with passed data,key set on it
    <>{postElements}</>
  );
}
// Posts (root)
// Each post (branch)
// userId
// id
// title
// body
