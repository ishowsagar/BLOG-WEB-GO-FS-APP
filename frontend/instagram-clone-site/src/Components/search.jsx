import { useState } from "react";
import { Link } from "react-router-dom";
import SearchIcon from "../assets/icons/search.png";
import { apiUrl } from "../Services/apiConfig";

const FILTERS = ["Top", "Accounts", "Reels", "Tags"];

const TRENDING_TOPICS = [
  { title: "#frontend", meta: "1.2M posts" },
  { title: "#uiux", meta: "834K posts" },
  { title: "#coding", meta: "2.4M posts" },
  { title: "#aesthetic", meta: "512K posts" },
];

const SUGGESTIONS = [
  { name: "alex.dev", role: "Frontend creator" },
  { name: "maria.codes", role: "UI engineer" },
  { name: "sam.builds", role: "Product designer" },
];

const RECENT_SEARCHES = [
  "instagram clone",
  "ui ideas",
  "reels trend",
  "dark theme",
];

export default function Search() {
  const [query, setQuery] = useState(""); // stores search query input
  const [activeFilter, setActiveFilter] = useState("Top");
  const [searchedUsers, setSearchedUsers] = useState([]); // storing data in state fetched from fetch 'req'
  const [searchErr, setSearchErr] = useState("");
  const [searchTriggered, setSearchTriggered] = useState(false);
  const [followeeID, setFolloweeID] = useState(null); // starting as null
  const [alreadyFollowedErr, setAlreadyFollowedErr] = useState(""); // for setting err when there is an already flw err
  const [hasAlreadyFollowed, setHasAlreadyFollowed] = useState(false); // for conditonally flipping bool to true if hit any err
  const [hasFollowed, setHasFollowed] = useState(false); // for conditonally flipping bool to true if hit any err
  const [currentFollowee, setCurrentFollowee] = useState(null);

  const showOverlay = searchTriggered && searchedUsers.length > 0;

  // todo - instead of fetching everytime, fetch once and export and use it where'd be needed neccesarily
  const token = localStorage.getItem("token");
  // search users
  async function SearchUsers() {
    setSearchTriggered(true);

    // early return
    if (!token) {
      return;
    }

    // request
    // todo - need validation for empty query too in server
    const payload = {
      //todo - might need a  ternary check for if search query exists in state fetch on this or that
      url: query
        ? apiUrl(`/api/users/search?name=${encodeURIComponent(query)}`)
        : apiUrl("/api/users/search?name=Owner"),
      header: {
        Authorization: token,
      },
      method: "GET",
    };

    // try-catch block for asynchronous process of fetching data & err handeling very neatly
    try {
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

      // if it was successfull
      setSearchedUsers(res.Data); // stores in arr

      // setting to store user which need to follow
      // setFolloweeID(res.Data[0].id)
    } catch (err) {
      // catch recieves err in its err object, err {message : stores err thrown by try block}
      setSearchErr(err.message);
      console.log("error :", err.message);
      console.log("stackErr :", err.stack);
    }
  }

  // search query
  function handleSearchQuery(event) {
    const { value, name } = event.target; // targetted field
    setQuery(value);
    setSearchTriggered(false);
    // console.log(query);
  }

  // follow button - since it is rendering mapped div, has userData, calling this function with passed id
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
      setHasFollowed(true); // set to true when followed

      // check which div is clicked -> which user is clicked to followe
      // whichUserAlreadyFollowedCheck(followeeid,followeeID)

      console.log(
        "client successfully followed a user, with userID -",
        followeeid,
      );
    } catch (err) {
      // catch recieves err in its err object, err {message : stores err thrown by try block}
      setAlreadyFollowedErr(err.message);
      setFolloweeID(null);
      console.log("error :", err.message);
      console.log("stackErr :", err.stack);
    }
  }

  // already follow condition - if both ids match then return trur
  function whichUserAlreadyFollowedCheck(currentFolloweeID, storedFolloweeID) {
    const check = currentFolloweeID === storedFolloweeID;
    if (check === true) {
      console.log(check);
      setCurrentFollowee(true);
    } else {
      console.log(check);
      setCurrentFollowee(false);
    }
  }

  return (
    <section className="search_page">
      <div className="search_page_shell">
        <div className="search_page_hero">
          <div>
            <p className="search_page_kicker">Explore</p>
            <h1 className="search_page_title">
              Search Instagram-style content.
            </h1>
            <p className="search_page_subtitle">
              Find accounts, reels, hashtags, and posts with a clean discover
              page that matches your app vibe.
            </p>
            {alreadyFollowedErr && (
              <div>
                <p
                  style={{
                    color: "#b42318",
                    fontSize: "27px",
                    paddingLeft: "4px",
                    fontWeight: 700,
                    borderLeft: "6px solid #000",
                    marginLeft: "10px",
                    padding: "8px 12px",
                    borderRadius: "6px",
                    background: "rgba(180, 35, 24, 0.04)",
                  }}
                >{`${alreadyFollowedErr}❗`}</p>
              </div>
            )}
          </div>

          <div
            className="search_page_searchbox"
            style={{ position: "relative" }}
          >
            <img
              src={SearchIcon}
              alt="search"
              className="search_page_searchicon"
            />
            <input
              type="text"
              value={query}
              onChange={handleSearchQuery}
              placeholder="Search people, reels, hashtags..."
              className="search_page_input"
            />
            {query && (
              <button
                type="button"
                className="search_page_clear"
                onClick={SearchUsers}
              >
                Search
              </button>
            )}

            {showOverlay && (
              <div
                className="search_users_overlay"
                style={{
                  position: "absolute",
                  top: "calc(100% + 10px)",
                  left: 0,
                  right: 0,
                  background: "#fff",
                  borderRadius: "16px",
                  border: "1px solid #e9e9e9",
                  boxShadow: "0 12px 30px rgba(0, 0, 0, 0.12)",
                  maxHeight: "320px",
                  overflowY: "auto",
                  zIndex: 30,
                }}
              >
                {/* displaying fetched data of searchedUsers, fetched from search fetch request */}
                {searchedUsers.map((user, index) => {
                  const displayName =
                    user?.name || user?.username || "Unknown User";
                  const handle = user?.username || user?.email || "user";
                  const subText =
                    user?.nickname || user?.bio || "Instagram user";
                  const avatarText = displayName.charAt(0).toUpperCase();

                  return (
                    <div
                      key={user?.id || user?.email || `${handle}-${index}`}
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: "12px",
                        padding: "12px 14px",
                        justifyContent: "flex-start",
                        borderBottom:
                          index === searchedUsers.length - 1
                            ? "none"
                            : "1px solid #f0f0f0",
                      }}
                    >
                      <div
                        style={{
                          width: "40px",
                          height: "40px",
                          borderRadius: "50%",
                          display: "grid",
                          placeItems: "center",
                          background:
                            "linear-gradient(135deg, #ffe0b2, #ffcdd2)",
                          fontWeight: 700,
                          color: "#2a2a2a",
                        }}
                      >
                        {avatarText}
                      </div>

                      <div
                        className="box"
                        style={{
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "space-between",
                          width: "100%",
                          gap: "12px",
                          minWidth: 0,
                        }}
                      >
                        {/* meta data of user */}
                        <div style={{ minWidth: 0, flex: 1 }}>
                          <p
                            style={{
                              margin: 0,
                              fontWeight: 700,
                              color: "#111",
                            }}
                          >
                            {displayName}
                          </p>
                          <p
                            style={{
                              margin: "2px 0 0",
                              fontSize: "0.86rem",
                              color: "#555",
                            }}
                          >
                            @{handle}
                          </p>
                          <p
                            style={{
                              margin: "2px 0 0",
                              fontSize: "0.8rem",
                              color: "#7a7a7a",
                            }}
                          >
                            {subText}
                          </p>
                        </div>

                        {/*//* redirects to link which render EachProfile Data, and send id in url --> fetch data of that profile from id gotten from the url poram */}
                        <Link to={`/users/profile/${user.id}`}>
                          <button
                            type="button"
                            // onClick={() => handleFollow(user.id)}
                            style={{
                              border: "none",
                              borderRadius: "999px",
                              padding: "7px 14px",
                              fontSize: "0.8rem",
                              fontWeight: 700,
                              color: "#fff",
                              background:
                                "linear-gradient(135deg, #3ec1fd, #082b95)",
                              cursor: "pointer",
                              boxShadow: "0 4px 12px rgba(255, 95, 109, 0.28)",
                              flexShrink: 0,
                              marginLeft: "auto",
                            }}
                          >
                            Profile
                          </button>
                        </Link>

                        {/* follow button */}
                        <button
                          type="button"
                          onClick={() => handleFollow(user.id)}
                          style={{
                            border: "none",
                            borderRadius: "999px",
                            padding: "7px 14px",
                            fontSize: "0.8rem",
                            fontWeight: 700,
                            color: "#fff",
                            background:
                              "linear-gradient(135deg, #ff5f6d, #ff8c42)",
                            cursor: "pointer",
                            boxShadow: "0 4px 12px rgba(255, 95, 109, 0.28)",
                            flexShrink: 0,
                            marginLeft: "auto",
                          }}
                        >
                          Follow
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <div className="search_page_filters">
            {FILTERS.map((filter) => (
              <button
                key={filter}
                type="button"
                onClick={() => setActiveFilter(filter)}
                className={`search_page_filter ${activeFilter === filter ? "active" : ""}`}
              >
                {filter}
              </button>
            ))}
          </div>
        </div>

        <div className="search_page_grid">
          <article className="search_card search_card_featured">
            <p className="search_card_label">Trending now</p>
            <h2 className="search_card_title">Popular hashtags and topics</h2>
            <div className="search_trending_list">
              {TRENDING_TOPICS.map((topic) => (
                <div className="search_trending_item" key={topic.title}>
                  <div>
                    <h3>{topic.title}</h3>
                    <p>{topic.meta}</p>
                  </div>
                  <span>Explore</span>
                </div>
              ))}
            </div>
          </article>

          <article className="search_card search_card_secondary">
            <p className="search_card_label">Suggested accounts</p>
            <h2 className="search_card_title">People you may want to follow</h2>
            <div className="search_people_list">
              {SUGGESTIONS.map((person) => (
                <div className="search_person_item" key={person.name}>
                  <div className="search_person_avatar">
                    {person.name.slice(0, 1).toUpperCase()}
                  </div>
                  <div>
                    <h3>{person.name}</h3>
                    <p>{person.role}</p>
                  </div>
                </div>
              ))}
            </div>
          </article>

          <article className="search_card search_card_secondary">
            <p className="search_card_label">Recent searches</p>
            <h2 className="search_card_title">Your latest searches</h2>
            <div className="search_recent_list">
              {RECENT_SEARCHES.map((item) => (
                <div className="search_recent_item" key={item}>
                  <span className="search_recent_bullet" />
                  <span>{item}</span>
                </div>
              ))}
            </div>
          </article>

          <article className="search_card search_card_full">
            <p className="search_card_label">Search results</p>
            <h2 className="search_card_title">Powered BY Postgres DB</h2>
            <p className="search_empty_state">
              Handlers call those search endpoint, then render posts, accounts,
              or reels depending on the selected tab:{" "}
              <strong>{activeFilter}</strong>.
            </p>
            <div className="search_empty_placeholder">
              <span className="search_empty_dot" />
              <span className="search_empty_dot" />
              <span className="search_empty_dot" />
            </div>
          </article>
        </div>
      </div>
    </section>
  );
}
