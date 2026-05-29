import React, { useEffect, useMemo, useRef, useState } from "react";
import { useLocation, Outlet, useNavigate, Link } from "react-router-dom";
import ProfileNav from "./ProfileNav";
import WebSocketDebug from "./WebSocketDebug";
import { apiUrl } from "../Services/apiConfig";
export default function Profile() {
  console.log("/profile");
  // states
  const [logout, setLogout] = useState(false); // boolean for conditional rendering n store log out states
  const [showPfpOverlay, setShowPfpOverlay] = useState(false);
  const [selectedPfpFile, setSelectedPfpFile] = useState(null);
  const [selectedPfpPreview, setSelectedPfpPreview] = useState("");
  const [isUploadingPfp, setIsUploadingPfp] = useState(false);
  const [pfpMessage, setPfpMessage] = useState("");
  const [profileAvatarSrc, setProfileAvatarSrc] = useState("");
  const fileInputRef = useRef(null);
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
  const profileAvatar = useMemo(
    () =>
      profileData?.avatar ||
      profileData?.avatar_url ||
      profileData?.profile_picture ||
      profileData?.pfp ||
      "",
    [profileData],
  );

  useEffect(() => {
    async function loadProfilePicture() {
      if (!token) return;

      try {
        const response = await fetch(apiUrl("/api/s3/pfp"), {
          method: "GET",
          headers: {
            Authorization: token,
            "Content-Type": "application/json",
          },
        });

        const data = await response.json();

        if (response.ok && data.Ok && data.ImageURL) {
          setProfileAvatarSrc(data.ImageURL);
          return;
        }
      } catch {
        // keep the profile data avatar or empty state below
      }

      setProfileAvatarSrc(profileAvatar);
    }

    loadProfilePicture();
  }, [profileAvatar]);

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

  useEffect(() => {
    if (!selectedPfpFile) {
      setSelectedPfpPreview("");
      return undefined;
    }

    const previewUrl = URL.createObjectURL(selectedPfpFile);
    setSelectedPfpPreview(previewUrl);

    return () => URL.revokeObjectURL(previewUrl);
  }, [selectedPfpFile]);

  async function uploadProfilePhoto() {
    if (!selectedPfpFile || !token || isUploadingPfp) return;

    setIsUploadingPfp(true);
    setPfpMessage("");

    const uploadUrl = apiUrl("/api/user/pfp/upload");

    try {
      const response = await fetch(uploadUrl, {
        method: "POST",
        headers: {
          Authorization: token,
          "Content-Type": selectedPfpFile.type || "application/octet-stream",
        },
        body: selectedPfpFile,
      });

      const data = await response.json();

      if (!response.ok || !data.Ok) {
        throw new Error(data.Error || data.Status || "failed to upload photo");
      }

      const pfpResponse = await fetch(apiUrl("/api/s3/pfp"), {
        method: "GET",
        headers: {
          Authorization: token,
          "Content-Type": "application/json",
        },
      });

      const pfpData = await pfpResponse.json();

      if (!pfpResponse.ok || !pfpData.Ok) {
        throw new Error(
          pfpData.Error || pfpData.Status || "failed to fetch uploaded photo",
        );
      }

      const uploadedUrl =
        pfpData.ImageURL ||
        pfpData.ImageUrl ||
        pfpData.imageUrl ||
        pfpData.URL ||
        "";

      if (uploadedUrl) {
        setProfileAvatarSrc(uploadedUrl);

        setProfileData((prev) => ({
          ...prev,
          avatar: uploadedUrl,
          avatar_url: uploadedUrl,
          profile_picture: uploadedUrl,
          pfp: uploadedUrl,
        }));
      }

      setPfpMessage("Profile photo updated");
      setShowPfpOverlay(false);
      setSelectedPfpFile(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
    } catch (err) {
      setPfpMessage(err.message || "failed to upload photo");
    } finally {
      setIsUploadingPfp(false);
    }
  }

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
        <div className="profile_avatar_shell">
          <img
            src={profileAvatarSrc || profileAvatar}
            alt="Profile avatar"
            className="profile_avatar"
          />
          <button
            type="button"
            className="profile_avatar_edit_button"
            onClick={() => setShowPfpOverlay(true)}
            aria-label="upload profile photo"
          >
            <span aria-hidden="true">✎</span>
          </button>
        </div>
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
      {pfpMessage && (
        <div className="profile_pfp_message" role="status">
          {pfpMessage}
        </div>
      )}
      <div className="profile_buttons">
        <button className="profile_button">
          <Link to="/profile/edit">Edit Profile</Link>
        </button>
        <button className="profile_button">View archive</button>
        <button onClick={handleClientLogout} className="profile_button">
          Log out
        </button>
      </div>
      {showPfpOverlay && (
        <div
          className="profile_pfp_overlay"
          onClick={() => setShowPfpOverlay(false)}
        >
          <div
            className="profile_pfp_modal"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="profile_pfp_modal_header">
              <div>
                <div className="profile_pfp_modal_title">Update photo</div>
                <div className="profile_pfp_modal_subtitle">
                  Choose an image and upload it to your profile
                </div>
              </div>
              <button
                type="button"
                className="profile_pfp_modal_close"
                onClick={() => setShowPfpOverlay(false)}
              >
                Close
              </button>
            </div>

            <div className="profile_pfp_dropzone">
              <input
                ref={fileInputRef}
                className="profile_pfp_input"
                type="file"
                accept="image/*"
                onChange={(event) => {
                  const nextFile = event.target.files?.[0] || null;
                  setSelectedPfpFile(nextFile);
                  setPfpMessage("");
                }}
              />

              <div className="profile_pfp_preview_wrap">
                <img
                  className="profile_pfp_preview"
                  src={selectedPfpPreview || profileAvatarSrc || profileAvatar}
                  alt="Selected preview"
                />
                <div className="profile_pfp_preview_hint">
                  Click the file picker to choose a new profile photo.
                </div>
              </div>

              <div className="profile_pfp_actions">
                <button
                  type="button"
                  className="profile_button"
                  onClick={uploadProfilePhoto}
                  disabled={!selectedPfpFile || isUploadingPfp}
                >
                  {isUploadingPfp ? "Uploading..." : "Upload photo"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

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
