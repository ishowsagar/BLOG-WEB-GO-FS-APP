import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useContext } from "react";
import { RealtimeContext } from "../Layout/MainLayout";
import { apiUrl } from "../Services/apiConfig";

export default function EachProfile() {
  console.log("/eachProfile");
  // states
  // Free online images for demo
  const avatarImg =
    "https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=facearea&w=256&h=256&facepad=2";
  const highlightImg =
    "https://images.unsplash.com/photo-1506744038136-46273834b3fb?auto=format&fit=crop&w=128&q=80";

  // & fetch profile data
  const [profileData, setProfileData] = useState({}); // sets profile data when mounted of loaded profile
  const [followeeID, setFolloweeID] = useState(null); // storing in state id of loaded profile
  const [alreadyFollowedErr, setAlreadyFollowedErr] = useState("");
  const [hasFollowed, setHasFollowed] = useState(false);
  const [followedLS, setFollowedLS] = useState(null);
  const [error, setError] = useState("");
  const [showDmBox, setShowDmBox] = useState(false);
  const [dmText, setDmText] = useState("");
  const [dmStatus, setDmStatus] = useState("");
  const [dmMessages, setDmMessages] = useState([]);
  const dmListRef = useRef(null);

  //   utils
  const token = localStorage.getItem("token");
  const navigate = useNavigate();
  const realtime = useContext(RealtimeContext);
  const sendDm = realtime?.sendDm;
  const subscribeDm = realtime?.subscribeDm;
  let params = useParams(); // invoking function return object which -> stores params
  const userid = params.userid;
  const followedKey = userid ? `followed_${userid}` : null;
  const followedBoolean = followedKey
    ? localStorage.getItem(followedKey) === "true"
    : false; //* localStorage stores strings values,not boolean, s

  //   on page mount,load profile data, api request on url with userid attached and fetched from current url, id passed from redirect
  useEffect(() => {
    // bug - err when using other hooks inside to fetch param
    // test - testing to load out of useEffect
    // fixed - fixed with just moving hooks out🔥🔥
    async function fetchProfileDataByID() {
      // fetch userid from the url
      console.log("opened profile of user with userID :", userid);
      const url = apiUrl(`/api/user/profile/${userid}`);

      //   fetch

      const reqMethod = "GET";
      try {
        const req = await fetch(url, {
          method: reqMethod,
          headers: {
            Authorization: token, // direct define Auth header in braces
          },
        });

        const response = await req.json();
        console.log("profile data :", response);

        //  if fetch\response struct's bool req was not successfull
        if (req.ok || !response.OK) {
          console.log("error fetching data :", response.Status);
          setError(response.Status);
        }

        setProfileData(response.User); //* if response was a success - storing resp in a state
        console.log("user profile data :", response.User);
      } catch (err) {
        console.log(err);
      }
    }
    fetchProfileDataByID();
  }, [token, hasFollowed]); //mount again if these dependency elements changes

  const followers = `${profileData?.followers_count} followers`;
  const followings = `${profileData?.following_count} following`;

  // resp is undefined -> handler must return these fetched from repo corres method
  //  fixed - now controller method on this route is returning those fields too
  const Username = `${profileData?.username}`;
  const Nickname = `${profileData?.nickname}`;
  const Bio = `${profileData?.bio}`;
  const receiverId = Number(userid);

  function getCurrentUserIdFromToken() {
    if (!token) return null;
    try {
      const payload = token.split(".")[1];
      if (!payload) return null;
      const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
      const decoded = JSON.parse(atob(normalized));
      return decoded.user_id ? Number(decoded.user_id) : null;
    } catch (err) {
      console.log("failed to decode token", err);
      return null;
    }
  }

  // follow need to pass id of person which has to be followed
  async function handleFollow(followeeid) {
    console.log(" attempt to follow a user...");

    //follow req

    // early return
    if (!token) {
      return;
    }

    // logging each search user id when hit follow
    // const userID = searchedUsers.map((user) => console.log(user));
    console.log("followee :", followeeid);

    // request
    // todo - need to fetch user id of fetched user who needed to follow - followeeID
    const payload = {
      url: apiUrl(`/api/users/follow/${followeeid}`),
      header: {
        Authorization: token,
      },
      method: "POST",
    };

    // try-catch block for asynchronous process of fetching data & err handeling very neatly
    try {
      // resetting states to default when func initialized its operation
      setAlreadyFollowedErr("");

      const req = await fetch(payload.url, {
        method: payload.method,
        headers: payload.header,
      });

      const res = await req.json();

      // err check from client and server
      console.log(res);
      if (!req.ok || !res.Ok) {
        throw new Error(res.Status);
      }

      // if it was a successfull follow operation
      setFolloweeID(followeeid); // storing id of recently followed user
      //   setHasFollowed(true); // to true when followed

      // check which div is clicked -> which user is clicked to followe
      // whichUserAlreadyFollowedCheck(followeeid,followeeID)

      // * set to true if already followed err response is returned from server

      console.log(
        "client successfully followed a user, with userID -",
        followeeid,
      );
      // storing following state in ls
      // fixed - now working if once set,get stored in ls and get on comp mount✅✅
      if (followedKey) {
        localStorage.setItem(followedKey, "true");
      }
      setHasFollowed(true);
    } catch (err) {
      // catch recieves err in its err object, err {message : stores err thrown by try block}
      setAlreadyFollowedErr(err.message);
      //   test - testing to flip bool to true if err status matches to conditonally render followed or follow text
      //   fixed - worked, now flipping boolean to true and displaying conditinal texts
      // bug  - * need a way to handle follow, update followers count when followed
      // test - making mount load on err dependency in useEffect if this bool changes -> re fetch data
      if (err.message === "user already followed") {
        // todo - need a way to send server response if user has been already followed or not by server on mount
        // track if client had already followed or not, rn checking on follow
        setHasFollowed(true);

        // window.location.reload(); // make a reload to let useEffect remount on page to update data
      }
      setFolloweeID(null);
      console.log("error :", err.message);
      console.log("stackErr :", err.stack);
    }
  }

  function handleSendDm() {
    const senderId = getCurrentUserIdFromToken();

    if (!senderId || !receiverId || !dmText.trim()) {
      setDmStatus("unable to send message");
      return;
    }

    const outgoingMessage = {
      sender_id: senderId,
      reciever_id: receiverId,
      type: "dm",
      content: dmText.trim(),
    };

    setDmMessages((prev) => [
      ...prev,
      { ...outgoingMessage, direction: "outgoing" },
    ]);
    sendDm(outgoingMessage);

    setDmStatus("message sent");
    setDmText("");
  }

  useEffect(() => {
    const senderId = getCurrentUserIdFromToken();
    if (!token || !senderId || !receiverId || !subscribeDm) return undefined;

    const unsubscribe = subscribeDm((notificationRaw) => {
      // normalize keys that may vary between publishers
      const notification = { ...notificationRaw };
      if (notification.receiver_id && !notification.reciever_id) {
        notification.reciever_id = notification.receiver_id;
      }
      if (notification.senderId && !notification.sender_id) {
        notification.sender_id = notification.senderId;
      }

      if (notification.type !== "dm") return;

      const incomingSenderId = Number(notification.sender_id);
      const incomingReceiverId = Number(notification.reciever_id);
      const incomingReceiverName =
        notification.reciever_name || notification.recieverName;
      const incomingSenderName =
        notification.sender_name || notification.senderName;

      if (incomingSenderId === senderId && incomingReceiverId === receiverId) {
        return;
      }

      if (incomingSenderId === receiverId && incomingReceiverId === senderId) {
        setDmMessages((prev) => [
          ...prev,
          {
            sender_id: incomingSenderId,
            reciever_id: incomingReceiverId,
            type: "dm",
            content: notification.content,
            direction: "incoming",
            senderName: incomingSenderName,
            recieverName: incomingReceiverName,
          },
        ]);
        setDmStatus("new message received");
      }
    });

    return unsubscribe;
  }, [subscribeDm, token, receiverId]);

  useEffect(() => {
    if (!showDmBox || !dmListRef.current) return;

    dmListRef.current.scrollTop = dmListRef.current.scrollHeight;
  }, [dmMessages, showDmBox]);

  return (
    <div className="profile_outer">
      <div className="profile_header_row">
        <img
          src={highlightImg}
          alt="Profile avatar"
          className="profile_avatar"
        />
        <div className="profile_header_info">
          <div className="profile_username">
            {Username} <span>⚙️</span>
          </div>
          <div className="profile_name">
            {Nickname}
            <span>♪</span>
          </div>
          <div className="profile_stats profile_stats_row">
            <span>2 posts</span>
            <span>{followers}</span>
            <span>{followings}</span>
          </div>
        </div>
      </div>
      <div className="profile_bio">
        {Bio}
        <span style={{ color: "#ff69b4" }}></span>
      </div>
      <div className="profile_buttons">
        <button
          onClick={() => handleFollow(profileData.id)}
          className="profile_button"
        >
          {/* conditionally button text if followed from ls==*/}
          {hasFollowed ? `Followed` : `follow`}
        </button>
        <button
          className="profile_button"
          onClick={() => setShowDmBox((prev) => !prev)}
        >
          Message
        </button>
      </div>

      {showDmBox && (
        <div
          style={{
            position: "fixed",
            right: "1rem",
            bottom: "1rem",
            width: "min(360px, calc(100vw - 2rem))",
            maxHeight: "70vh",
            padding: "1rem",
            borderRadius: "18px",
            background: "rgba(15, 15, 24, 0.96)",
            border: "1px solid rgba(255,255,255,0.08)",
            boxShadow: "0 24px 70px rgba(0,0,0,0.38)",
            backdropFilter: "blur(18px)",
            display: "grid",
            gap: "0.85rem",
            zIndex: 9999,
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              gap: "1rem",
            }}
          >
            <div>
              <strong style={{ color: "#fff", display: "block" }}>
                Direct Message
              </strong>
              <span
                style={{ color: "rgba(255,255,255,0.7)", fontSize: "0.88rem" }}
              >
                Chat with {Nickname || Username}
              </span>
            </div>
            <button
              className="profile_button"
              onClick={() => setShowDmBox(false)}
              style={{ padding: "0.45rem 0.85rem" }}
            >
              Close
            </button>
          </div>

          <div
            ref={dmListRef}
            style={{
              maxHeight: "36vh",
              overflowY: "auto",
              display: "grid",
              gap: "0.6rem",
              padding: "0.45rem",
              borderRadius: "14px",
              background: "rgba(0,0,0,0.24)",
            }}
          >
            {dmMessages.length === 0 ? (
              <div style={{ color: "rgba(255,255,255,0.7)" }}>
                Start the conversation.
              </div>
            ) : (
              dmMessages.map((message, index) => (
                <div
                  key={`${message.direction}-${index}-${message.content}`}
                  style={{
                    justifySelf:
                      message.direction === "outgoing" ? "end" : "start",
                    maxWidth: "85%",
                    padding: "0.75rem 0.9rem",
                    borderRadius:
                      message.direction === "outgoing"
                        ? "16px 16px 6px 16px"
                        : "16px 16px 16px 6px",
                    background:
                      message.direction === "outgoing"
                        ? "linear-gradient(135deg, #ff4d9d, #ff8a5b)"
                        : "rgba(255,255,255,0.08)",
                    color: "#fff",
                    lineHeight: 1.4,
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "0.5rem",
                      marginBottom: "0.35rem",
                    }}
                  >
                    <img
                      src={
                        message.direction === "outgoing"
                          ? highlightImg
                          : avatarImg
                      }
                      alt={
                        message.direction === "outgoing" ? "You" : "Receiver"
                      }
                      style={{
                        width: "20px",
                        height: "20px",
                        borderRadius: "50%",
                        objectFit: "cover",
                        border: "1px solid rgba(255,255,255,0.15)",
                      }}
                    />
                    <div
                      style={{
                        fontSize: "0.8rem",
                        opacity: 0.78,
                      }}
                    >
                      {message.direction === "outgoing"
                        ? "You"
                        : `${message.senderName || `User ${message.sender_id}`}`}
                    </div>
                  </div>
                  <div>{message.content}</div>
                </div>
              ))
            )}
          </div>

          <textarea
            placeholder="Write a direct message..."
            value={dmText}
            onChange={(e) => setDmText(e.target.value)}
            rows={4}
            style={{
              width: "100%",
              padding: "0.8rem",
              borderRadius: "12px",
              border: "1px solid rgba(255,255,255,0.12)",
              background: "#111",
              color: "#fff",
              resize: "vertical",
            }}
          />
          <div
            style={{ display: "flex", gap: "0.75rem", alignItems: "center" }}
          >
            <button className="profile_button" onClick={handleSendDm}>
              Send DM
            </button>
            {dmStatus && <span style={{ color: "#9eff9e" }}>{dmStatus}</span>}
          </div>
        </div>
      )}

      {/* <div className="profile_highlights">
        <div className="profile_highlight">
          <img
            src={highlightImg}
            alt="Music heals"
            className="profile_highlight_img"
          />
          <div className="profile_highlight_label">
            Music heals
            <span style={{ color: "#ff69b4" }}>💗</span>
          </div>
        </div>
        <div className="profile_highlight">
          <div
            className="profile_highlight_img"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              background: "#222",
            }}
          >
            <span style={{ fontSize: "2.5rem", color: "#fff" }}>+</span>
          </div>
          <div className="profile_highlight_label">New</div>
        </div>
      </div> */}
      {/* seperately select based on routes */}
      <div className="profile_grid_nav">{/* <ProfileNav /> */}</div>
      {/* these things i want to apprear based on routes */}
      {/* <div className="profile_posts_grid"> */}
      {/* <img src={postImg1} alt="Post 1" className="profile_post_img" />
            <img src={postImg2} alt="Post 2" className="profile_post_img" /> */}
      {/* </div> */}
      {/* <Outlet /> */}
    </div>
  );
}
