import HomeIcon from "../assets/icons/homes.png";
import SearchIcon from "../assets/icons/search.png";
import ExploreIcon from "../assets/icons/explore.png";
import AiIcon from "../assets/icons/Ai.png";
import ReelsIcon from "../assets/icons/reels.png";
import CreateIcon from "../assets/icons/Create.png";
import MessagesIcon from "../assets/icons/msg.png";
import NotificationsIcon from "../assets/icons/noti.png";
import UserIcon from "../assets/icons/user.png";

import { Link } from "react-router-dom";

export default function Sidebar({ globalNotification }) {
  return (
    <aside>
      <nav className="sidebar_nav">
        <ul className="sidebar_list_items">
          <li>
            <Link className="sidebar_link" to="/">
              <img className="sidebar_icon" src={HomeIcon} alt="Home" />
              Home
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/search">
              <img className="sidebar_icon" src={SearchIcon} alt="Search" />
              Search
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/explore">
              <img className="sidebar_icon" src={ExploreIcon} alt="Explore" />
              Explore
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/create">
              <img className="sidebar_icon" src={CreateIcon} alt="Create" />
              Create
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/reels">
              <img className="sidebar_icon" src={ReelsIcon} alt="Reels" />
              Reels
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/messages">
              <img className="sidebar_icon" src={MessagesIcon} alt="Messages" />
              Messages
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/notify">
              <img className="sidebar_icon" src={NotificationsIcon} alt="Notifications" />
              Notification
              {/* conditionally rendering the badge number */}
              {globalNotification?.length > 0 && (
                <span className="sidebar-notification-badge">
                  {globalNotification.length}
                </span>
              )}
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/profile">
              <img className="sidebar_icon" src={UserIcon} alt="Profile" />
              Profile
            </Link>
          </li>
          <li>
            <Link className="sidebar_link" to="/denai">
              <img className="sidebar_icon" src={AiIcon} alt="Denver AI" />
              Denver AI
            </Link>
          </li>
        </ul>
      </nav>
    </aside>
  );
}
