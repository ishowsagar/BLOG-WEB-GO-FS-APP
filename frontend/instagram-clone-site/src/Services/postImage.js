const FALLBACK_POST_IMAGE = (postId) =>
  `https://picsum.photos/seed/${postId || "post"}/500/350`;

export function resolvePostImage(post) {
  const imageSource =
    post?.image_source || post?.imageSource || post?.ImageSource || "";

  if (typeof imageSource === "string" && imageSource.trim()) {
    if (imageSource === "default.png") {
      return FALLBACK_POST_IMAGE(post?.id || post?.ID);
    }

    return imageSource;
  }

  return FALLBACK_POST_IMAGE(post?.id || post?.ID);
}
