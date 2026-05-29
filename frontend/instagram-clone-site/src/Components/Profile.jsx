import React, { useEffect, useState } from "react";
import { useLocation, Outlet, useNavigate, Link } from "react-router-dom";
import pfp from "../assets/pfp.jpg";
import ProfileNav from "./ProfileNav";
import WebSocketDebug from "./WebSocketDebug";
import { apiUrl } from "../Services/apiConfig";
export default function Profile() {
  console.log("/profile");
  // states
  const [logout, setLogout] = useState(false); // boolean for conditional rendering n store log out states
  const navigate = useNavigate();
  // Free online images for demo
  const avatarImg =
    "https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=facearea&w=256&h=256&facepad=2";
  const highlightImg =
    "https://images.unsplash.com/photo-1506744038136-46273834b3fb?auto=format&fit=crop&w=128&q=80";

  const location = useLocation();
  const { pathname } = location;
  console.log(pathname);

  // & fetch profile data
  const [profileData, setProfileData] = useState({});
  const [error, setError] = useState("");

  const token = localStorage.getItem("token");
  useEffect(() => {
    async function fetchProfile() {
      //  todo - add method to fetch data
      // fixed - added controller and repo's corresponding methods to fetch data sequentially
      //  need to pass userID to fetch data for that profile - user struct data type
      // frotend does not need to pass id, it will be fetched by backend to req data
      const url = apiUrl("/api/profile");
      const reqMethod = "GET";
      try {
        const req = await fetch(url, {
          method: reqMethod,
          headers: {
            Authorization: token, // direct define Auth header in braces
          },
        });

        const response = await req.json();

        //  if fetch\response struct's bool req was not successfull
        if (req.ok || !response.OK) {
          console.log("error fetching data :", response.Status);
          setError(response.Status);
        }

        setProfileData(response.Data); //* if response was a success - storing resp in a state
        console.log("user profile data :", response.User);
      } catch (err) {
        console.log(err);
      }
    }
    fetchProfile();
  }, [token]);

  // logs out user from the current login session
  function handleClientLogout(e) {
    // log out logic - clear token, send user to login Page

    localStorage.removeItem("token");
    navigate("/login");
    setLogout(true);
    console.log("user logged out :", logout);
  }

  const followers = `${profileData?.followers_count} followers`;
  const followings = `${profileData?.following_count} following`;

  // resp is undefined -> handler must return these fetched from repo corres method
  //  fixed - now controller method on this route is returning those fields too
  const Username = profileData?.username || profileData?.user_name || "user";
  const DisplayName =
    profileData?.name ||
    profileData?.nickname ||
    profileData?.display_name ||
    Username;
  const postsCount = profileData?.post_count || profileData?.posts_count || 0;
  const Bio = profileData?.bio || "";

  return (
    <div className="profile_outer">
      <div className="profile_header_row">
        <img src={pfp} alt="Profile avatar" className="profile_avatar" />
        <div className="profile_header_info">
          <div className="profile_username">
            {Username} <span>⚙️</span>
          </div>
          <div className="profile_name">
            {DisplayName}
            <span>♪</span>
          </div>
          <div className="profile_stats profile_stats_row">
            <span>{postsCount} posts</span>
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
        <button className="profile_button">
          <Link to="/profile/edit">Edit Profile</Link>
        </button>
        <button className="profile_button">View archive</button>
        <button onClick={handleClientLogout} className="profile_button">
          Log out
        </button>
      </div>

      <div className="profile_highlights">
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
      </div>
      {/* seperately select based on routes */}
      <div className="profile_grid_nav">
        <ProfileNav />
        <WebSocketDebug token={token} />
      </div>
      {/* these things i want to apprear based on routes */}
      {/* <div className="profile_posts_grid"> */}
      {/* <img src={postImg1} alt="Post 1" className="profile_post_img" />
            <img src={postImg2} alt="Post 2" className="profile_post_img" /> */}
      {/* </div> */}
      <Outlet />
    </div>
  );
}
