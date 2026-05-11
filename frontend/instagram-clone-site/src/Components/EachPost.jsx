import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

export default function EachPost() {
  const [eachPost, setEachPost] = useState([]);
  const [postLikesCount, setPostLikesCount] = useState(0);
  const [showCommentBox, setShowCommentBox] = useState(false);
  const [showPostComments, setShowPostComments] = useState(false); // for all comments
  const [commentText, setCommentText] = useState("");
  const [postComments, setPostComments] = useState([]); // for storing all comments in [:]
  const [postCommentsErr, setPostCommentsErr] = useState("");
  const [postedComment, setPostedComment] = useState({});
  //   fetch id from url
  const { id } = useParams(); //* for fetching urlParams from /feed/:id
  const token = localStorage.getItem("token");
  //   console.log(id);
  //   console.log(token);

  //  make api call to that url to fetch that post exactly
  //   ! never call async on wrapper func but func inside it
  useEffect(() => {
    //* server controller method will send response{struct has ok bool to check if it was a sucess} from this url having id in its url param
    if (!id || !token) return; // early return
    const url = `http://localhost:8080/api/feed/post/${id}`; // if id and token exists✅
    async function fetchEachPostData() {
      try {
        const req = await fetch(url, {
          method: "GET",
          //need to send token header on every req - all handlers validate it + auth middleware
          headers: { Authorization: token },
        });
        const response = await req.json();
        // console.log("feed status:", response.status, response);
        // console.log(token);
        // server sends response with data struct, check Ok boolean on it
        if (!response.Ok) {
          throw new Error("failed to load post");
        }
        //  if response was Ok
        console.log("each post response loaded :", response.Post);
        setEachPost(response.Post);
        setPostLikesCount(response.Post?.like_count ?? 0);
      } catch (err) {
        console.log(err);
      }
    }
    fetchEachPostData();
  }, [token, id]);

  // & toggle comment box -> conditionally render comment box
  function handleCommentBox() {
    setShowCommentBox((prev) => !prev);
  }

  // & toggle all comments feed on the post -> conditionally render commnts feed of that post
  function handleShowAllComments() {
    // todo - add element which shows up when toggled true
    // fixed - add to show all comments
    setShowPostComments((prev) => !prev);
  }

  // * imp things need before comments thing
  // #1 Current post data - avaliable in "eachPost" state
  // #2 User who is commenting - active client "token"

  // add comment
  async function postComment(eachPost) {
    const commentsUrl = `http://localhost:8080/api/post/comment/${eachPost.id}`;
    try {
      const commentContent = {
        content: commentText,
      };
      const req = await fetch(commentsUrl, {
        method: "POST",
        //need to send token header on every req - all handlers validate it + auth middleware
        headers: { Authorization: token },
        body: JSON.stringify(commentContent),
      });
      const response = await req.json();
      // server sends response with data struct, check Ok boolean on it
      if (!response.Ok) {
        throw new Error("failed to post comments");
      }
      //  if response was Ok
      console.log(
        `comment successfully posted on this ${eachPost.id}`,
        response.Data,
      );
      //  since we are adding comments into db,just fetching once in loadComments so no need to store
      // * but we can store it in state for sake of dependency array reload, if posted it reloads
      setPostedComment(response.Data);
    } catch (err) {
      console.log(err);
    }
  }

  // invoke add comment on submission
  function handlePostComment() {
    // bug - it gives unknown err when posted comment, we making sure when cmt is posted
    // add dependency [:] being on this posted comment
    postComment(eachPost);
    setCommentText("");
  }

  // load all comments from - "/api/feed/comments/:postid"
  useEffect(() => {
    async function loadAllComments(eachPost) {
      console.log("comments loaded");
      // providing id from fetched post from current url in the eachPost
      const commentsUrl = `http://localhost:8080/api/feed/comments/${eachPost.id}`;
      try {
        const req = await fetch(commentsUrl, {
          method: "GET",
          //need to send token header on every req - all handlers validate it + auth middleware
          headers: { Authorization: token },
        });
        const response = await req.json();
        // console.log("feed status:", response.status, response);
        // console.log(token);
        // server sends response with data struct, check Ok boolean on it
        if (!response.Ok) {
          throw new Error("failed to load post comments");
        }
        //  if response was Ok
        console.log(
          `all comments are loaded of this post ${eachPost.id}`,
          response,
        );
        setPostComments(response.Data);
      } catch (err) {
        setPostCommentsErr(err.message);
        console.log(err);
      }
    }
    // invoke it
    // bug - it was returning undefined when fetching data from eachPost.id cause -> can't access directly its inside useEffect block - can't access variables inside useEffect inside func
    // fixed - passed generic parameter and while invoking with agruemented 'eachPost' state data - at this level var is accessible and passed down with ease
    loadAllComments(eachPost);
  }, [postComment]);

  if (!eachPost || eachPost.length === 0) {
    return <div style={{ fontSize: "20px", padding: "2rem" }}>loading...</div>;
  }

  const post = eachPost; // Use fetched post data
  console.log("post render on id:", post.id);

  // post  like
  async function handleLike() {
    // * get token from localstorage
    const token = localStorage.getItem("token");
    if (!token) {
      throw new Error("token expired or not found");
    }

    // * like post req
    const payload = {
      method: "POST",
      body: {
        post_id: eachPost.id,
      },
    };
    try {
      const likeReq = await fetch("http://localhost:8080/api/like", {
        method: payload.method,
        headers: {
          Authorization: token,
          "Content-type": "application/json",
        },
        body: JSON.stringify(payload.body),
      });

      const response = await likeReq.json();
      // if hit errors
      if (!likeReq.ok) {
        throw new Error("failed to like post");
      }
      console.log("liked post");
      console.log("post likes handler response :", response);
      setPostLikesCount(response.Like.like_count);
    } catch (err) {
      console.log(err);
    }
  }

  const displayUsername = eachPost.user_id && `insta-user-${eachPost.user_id}`;
  const displayImage = `https://picsum.photos/seed/${post.id}/500/350`;
  const avatarImg = eachPost.user_id
    ? `https://i.pravatar.cc/150?img=${eachPost.user_id}`
    : `https://i.pravatar.cc/150?u=${eachPost.id}`;

  return (
    <>
      <div className="feedpost_header">
        <img className="feedpost_avatar" src={avatarImg} alt="avatar" />
        <span className="feedpost_username">{displayUsername}</span>
        <span className="feedpost_time">{eachPost.created_at}</span>
      </div>
      <img className="feedpost_image" src={displayImage} alt="post" />
      <div className="feedpost_caption">
        <h2 className="feedpost_title">{eachPost.title}</h2>
        <p className="feedpost_body">{eachPost.content}</p>
      </div>
      <div className="feedpost_actions">
        <span className="like-btn" onClick={handleLike} aria-label="like">
          ❤️
        </span>
        <span
          className="like-btn"
          role="img"
          aria-label="comment"
          onClick={handleCommentBox}
        >
          💬
        </span>
        <span className="like-btn" role="img" aria-label="share">
          ➡️
        </span>
      </div>
      <div className="feedpost_footer">
        <span className="feedpost_likes">{postLikesCount} likes</span>
        <span onClick={handleShowAllComments} className="feedpost_comments">
          View all comments
        </span>
      </div>
      {/* this is how we toggle comment box on off using state flipping */}
      {showCommentBox && (
        <form
          className="comment-form"
          onSubmit={(e) => {
            e.preventDefault();
            console.log("new comment:", commentText);
            // placeholder: wire API here later
            setCommentText("");
            setShowCommentBox(false);
          }}
        >
          <textarea
            placeholder="Write a comment..."
            value={commentText}
            onChange={(e) => setCommentText(e.target.value)}
            rows={3}
          />
          <div className="comment-form__actions">
            <button
              onClick={handlePostComment}
              type="submit"
              disabled={!commentText.trim()}
            >
              Post
            </button>
          </div>
        </form>
      )}

      {/* show all comments -> if this bool is true <-=> then render feed*/}
      {showPostComments && (
        <div className="post-comments-feed">
          <p className="post-comments-feed__title">Comments</p>
          <div className="post-comments-feed__list">
            {postComments.map((comment) => (
              <div className="post-comments-feed__item" key={comment.id}>
                <span className="post-comments-feed__user">
                  UserID : {comment.user_id}
                </span>
                <p>{comment.content}</p>
              </div>
            ))}
          </div>
        </div>
      )}
    </>
  );
}
