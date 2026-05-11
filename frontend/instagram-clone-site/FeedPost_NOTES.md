# FeedPost Component Update Notes

## What was done:

- Refactored `FeedPost` to render each post as a realistic Instagram-style card.
- Each post card includes:
  - User avatar (using pravatar.cc with userId as seed)
  - Username and timestamp
  - Post image (using picsum.photos with post.id as seed)
  - Post title and body as caption
  - Action icons for like, comment, and share
  - Footer with random likes and comments count
- Added new CSS classes for layout and styling:
  - `.feedpost_container-layout`, `.feedpost`, `.feedpost_header`, `.feedpost_avatar`, `.feedpost_username`, `.feedpost_time`, `.feedpost_image`, `.feedpost_caption`, `.feedpost_title`, `.feedpost_body`, `.feedpost_actions`, `.feedpost_footer`, `.feedpost_likes`, `.feedpost_comments`
- Used `Math.random()` to generate demo likes (50-249) and comments (1-10) for each post.
- Used placeholder image services for avatars and post images to simulate a real feed.

## Why:

- To make the feed visually similar to Instagram posts
- To provide a clean, modern UI for demo purposes
- To help you easily extend or style further as needed

---

This file documents the changes and approach for the FeedPost component update.

# FeedPost Component: Notes & Docs

## Generating Arrays for Image Grids (e.g., Explore Page)

### How to Generate an Array of Random Image URLs

We use the following pattern to create an array of image URLs for grid layouts:

```js
const images = Array.from(
  { length: 100 },
  (_, i) => `https://loremflickr.com/300/300?random=${i}`
);
```

#### Explanation:

- `Array.from({ length: 100 })` creates an array-like object with 100 empty slots.
- The second argument (mapping function) is called for each index (i from 0 to 99).
- For each index, we return a string URL for a random image from LoremFlickr.
- 'Empty' here means the array has no values yet, just slots to fill.
- After mapping, each slot contains a unique image URL string.

#### Example Output:

```
[
  "https://loremflickr.com/300/300?random=0",
  "https://loremflickr.com/300/300?random=1",
  ...
]
```

#### Usage in React:

```jsx
return (
  <div className="explore-grid">
    {images.map((src, idx) => (
      <img key={idx} src={src} alt="explore" className="explore-img" />
    ))}
  </div>
);
```

This approach is useful for generating mock data and displaying grids of images in components like Explore or Feed.

---
