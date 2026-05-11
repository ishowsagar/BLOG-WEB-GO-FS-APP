import { Link, useLocation } from "react-router-dom";
export default function ProfileNav() {
  const { pathname } = useLocation();
  return (
    <>
      <Link to="/profile">
        <span
          className={`profile_grid_nav_icon${
            pathname === "/profile" ? " active" : ""
          }`}
        >
          ▦
        </span>
      </Link>
      <Link to="/profile/archieves">
        <span
          className={`profile_grid_nav_icon${
            pathname === "/profile/archieves" ? " active" : ""
          }`}
        >
          🔖
        </span>
      </Link>
      <Link to="/profile/adminOnly">
        <span
          className={`profile_grid_nav_icon${
            pathname === "/profile/adminOnly" ? " active" : ""
          }`}
        >
          👤
        </span>
      </Link>
    </>
  );
}
