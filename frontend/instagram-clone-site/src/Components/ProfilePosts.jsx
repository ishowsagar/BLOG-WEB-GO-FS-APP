import { use, useEffect, useState } from "react";
import { Link } from "react-router-dom";

const postImg1 =
  "https://images.unsplash.com/photo-1465101046530-73398c7f28ca?auto=format&fit=crop&w=400&q=80";
const postImg2 = "https://picsum.photos/id/1025/400/400";

export default function ProfilePosts() {
  // states
  const [posts, setPosts] = useState([]);
  const [fetchErr, setFetchErr] = useState("");
  const [fetchSuccessfull, setFetchSuccessfull] = useState(false);

  // fetch token of active client
  console.log("/profilePosts");
  const token = localStorage.getItem("token");
  if (!token) {
    return;
  }

  useEffect(() => {
    // fetch posts for active logged-in user
    async function FetchActiveClientPosts() {
      const payload = {
        method: "GET",
        header: {
          Authorization: token,
        },
        url: "http://3.84.111.249:8080/api/feed/client/posts",
      };

      try {
        console.log("fetching profile posts wait...");
        const req = await fetch(payload.url, {
          method: payload.method,
          headers: payload.header,
        });

        const res = await req.json();

        if (!req.ok || !res.Ok) {
          console.log("error :", res);
          throw new Error(res.Status);
        }

        setPosts(res.Data);
        console.log("successfully fetched profile posts✅ :", res.Data);
      } catch (err) {
        setFetchErr(err.message);
        setFetchSuccessfull(false);
      }
    }

    FetchActiveClientPosts();
  }, []);

  // posts el
  const postEls =
    posts.length > 0 &&
    posts.map((post) => {
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
                <span className="feedpost_username">{post.name}</span>
                <span className="feedpost_time">{post.created_at}</span>
              </div>
              <h2 className="feedpost_title">Title - {post.title}</h2>
              <img
                className="feedpost_image"
                src={`https://picsum.photos/seed/${post.id}/500/350`}
                alt="post"
              />
              <div className="feedpost_caption">
                <p className="feedpost_body">{post.content}</p>
              </div>
              <div className="post-interactor-btns">
                <p id="pp-feed-likes">❤️{post.like_count} Likes</p>
                {/* <button>💬</button>
              <button>⛓️‍💥</button> */}
              </div>
            </div>
          </Link>
        </div>
      );
    });
  return <div className="profile_posts_grid">{postEls}</div>;
}
