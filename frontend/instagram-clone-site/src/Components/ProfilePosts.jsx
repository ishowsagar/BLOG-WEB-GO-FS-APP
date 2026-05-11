const postImg1 =
  "https://images.unsplash.com/photo-1465101046530-73398c7f28ca?auto=format&fit=crop&w=400&q=80";
const postImg2 = "https://picsum.photos/id/1025/400/400";
export default function ProfilePosts() {
  return (
    <div className="profile_posts_grid">
      <img src={postImg1} alt="Post 1" className="profile_post_img" />
      <img src={postImg2} alt="Post 2" className="profile_post_img" />
    </div>
  );
}
