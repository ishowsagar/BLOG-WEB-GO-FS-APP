import icon from "../assets/profile.png";
export default function TagPosts() {
  return (
    <div className="no-posts-container">
      <img className="tagged-post-image" src={icon} alt="" />
      <h2>No Tagged Posts Found! </h2>
    </div>
  );
}
