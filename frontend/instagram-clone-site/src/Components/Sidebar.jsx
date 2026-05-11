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

export default function Sidebar() {
  return (
    <aside>
      <nav className="sidebar_nav">
        <ul className="sidebar_list_items">
          <li>
            <img className="sidebar_icon" src={HomeIcon} />
            <Link to="/">Home</Link>
          </li>
          <li>
            <img className="sidebar_icon" src={SearchIcon} />
            <a href="#">Search</a>
          </li>
          <li>
            <img className="sidebar_icon" src={ExploreIcon} />
            <Link to="/explore">Explore</Link>
          </li>
          <li>
            <img className="sidebar_icon" src={CreateIcon} />
            <Link to="/create">Create</Link>
          </li>
          <li>
            <img className="sidebar_icon" src={ReelsIcon} />
            <Link to="/reels">Reels</Link>
          </li>
          <li>
            <img className="sidebar_icon" src={MessagesIcon} />
            <a href="#">Messages</a>
          </li>
          <li>
            <img className="sidebar_icon" src={NotificationsIcon} />
            <a href="#">Notifications</a>
          </li>
          <li>
            <img className="sidebar_icon" src={UserIcon} />
            <Link to="/profile">Profile</Link>
          </li>
          <li>
            <img className="sidebar_icon" src={AiIcon} />
            <a href="#">Denver AI</a>
          </li>
        </ul>
      </nav>
    </aside>
  );
}
