import { useEffect, useState } from "react";
import { apiUrl } from "../Services/apiConfig";

export default function TestFeed() {
  const token =
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnkiOiIyMDI2LTA1LTAyVDE3OjQ1OjE4LjkyOTcyODYrMDU6MzAiLCJ1c2VyX2lkIjoxfQ.iqfGc5InWC3uIbGoffDHM9Sq5QokCQ69xYuuwhKvzHs";

  const [posts, setPosts] = useState([]);
  useEffect(() => {
    async function load() {
      if (!token) return; // no token -> won't call protected API
      const res = await fetch(apiUrl("/api/feed"), {
        method: "GET",
        headers: { Authorization: token },
      });
      if (!res.ok) return;
      const data = await res.json();
      setPosts(data.Post || []);
    }
    load();
  }, [token]);

  return (
    <>
      <h1>Blog titles</h1>
      {posts.map((blog) => (
        <div key={blog.id} style={{ backgroundColor: "skyblue" }}>
          <h3>{blog.title}</h3>
          <p>{blog.content}</p>
        </div>
      ))}
    </>
  );
}
