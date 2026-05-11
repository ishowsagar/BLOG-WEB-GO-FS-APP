import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Feed from "./Components/Feed";
import MainLayout from "./Layout/MainLayout";
import CreatePost from "./Components/CreatePost";
import Explore from "./Components/Explore";
import Reels from "./Components/Reels";
import Profile from "./Components/Profile";
import ProfilePosts from "./Components/ProfilePosts";
import ProfileArchieves from "./Components/ProfileArchieves";
import TagPosts from "./Components/TagPosts";
import TestFeed from "./Components/testfeed";
import { Signup } from "./Components/signup";
import { Login } from "./Components/login";
import RequireAuth from "./middleware/middleware";
import EachPost from "./Components/EachPost";
import { ResetPassword } from "./Components/resetPassword";

createRoot(document.getElementById("root")).render(
  // ! Main layout - shared header and footer, everything else will be rendered as,
  //!  whereever user navigates to that route
  <BrowserRouter>
    <Routes>
      {/*todo - set up Route for each component to be rendered   */}
      {/* ! parent layout, everything else -> loaded by outlet <- whatever path it goes into */}
      <Route path="/signup" element={<Signup />} />
      <Route path="/login" element={<Login />} />
      <Route path="/password/reset" element={<ResetPassword />} />
      {/* //! no route, cause it will check for token, outlet - render everything else on route  */}
      <Route element={<RequireAuth />}>
        {/*! auth  - will only render children if auth check passed*/}
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Feed />} />
          {/* render each post when client is navigated to this url */}
          <Route path="feed/:id" element={<EachPost />} />
          <Route path="test" element={<TestFeed />} />
          <Route path="create" element={<CreatePost />} />
          <Route path="explore" element={<Explore />} />
          <Route path="reels" element={<Reels />} />
          <Route path="profile" element={<Profile />}>
            {/* render everything of parent Component --> but render child routes where u placed outlet based on route */}
            <Route index element={<ProfilePosts />} />
            <Route path="archieves" element={<ProfileArchieves />} />
            <Route path="adminOnly" element={<TagPosts />} />
          </Route>
        </Route>
      </Route>
    </Routes>
  </BrowserRouter>,
);
