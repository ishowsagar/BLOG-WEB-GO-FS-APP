import React, { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiUrl } from "../Services/apiConfig";

const INITIAL_FEED_POST_FORM = {
  title: "",
  body: "",
};

const MAX_IMAGE_SIZE = 2 * 1024 * 1024;
const ALLOWED_IMAGE_TYPES = new Set(["image/jpeg", "image/png", "image/webp"]);

export default function CreatePost() {
  const [form, setForm] = useState(INITIAL_FEED_POST_FORM);
  const [postCreationErr, setPostCreationErr] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isFormValid, setIsFormValid] = useState(false);
  const [selectedImage, setSelectedImage] = useState(null);
  const [selectedImagePreview, setSelectedImagePreview] = useState("");
  const navigateBackToFeed = useNavigate();
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!selectedImage) {
      setSelectedImagePreview("");
      return undefined;
    }

    const previewUrl = URL.createObjectURL(selectedImage);
    setSelectedImagePreview(previewUrl);

    return () => URL.revokeObjectURL(previewUrl);
  }, [selectedImage]);

  const selectedImageSummary = useMemo(() => {
    if (!selectedImage) return "No image selected yet";

    const sizeInMb = (selectedImage.size / (1024 * 1024)).toFixed(2);
    return `${selectedImage.name} - ${sizeInMb} MB`;
  }, [selectedImage]);

  function validateForm(formData) {
    const isTitleValid = formData.title.trim().length > 0;
    const isBodyValid = formData.body.trim().length >= 10;
    setIsFormValid(isTitleValid && isBodyValid);
  }

  function handleFormChange(e) {
    const { name, value } = e.target;

    setForm((prev) => {
      const nextForm = {
        ...prev,
        [name]: value,
      };

      validateForm(nextForm);
      return nextForm;
    });
  }

  function handleImageChange(e) {
    const file = e.target.files?.[0] || null;

    if (!file) {
      setSelectedImage(null);
      return;
    }

    if (!ALLOWED_IMAGE_TYPES.has(file.type)) {
      setPostCreationErr("Please choose a JPEG, PNG, or WEBP image.");
      e.target.value = "";
      return;
    }

    if (file.size > MAX_IMAGE_SIZE) {
      setPostCreationErr("Image must be 2 MB or smaller.");
      e.target.value = "";
      return;
    }

    setPostCreationErr("");
    setSelectedImage(file);
  }

  function clearSelectedImage() {
    setSelectedImage(null);
  }

  async function addDelay(ms) {
    return new Promise((res) => setTimeout(res, ms));
  }

  async function createPost() {
    const postReq = await fetch(apiUrl("/api/post/create"), {
      method: "POST",
      headers: {
        Authorization: token,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        title: form.title,
        content: form.body,
      }),
    });

    const postCreationResponse = await postReq.json();

    if (!postReq.ok || !postCreationResponse.Ok) {
      throw new Error(postCreationResponse.Status || "failed to create post");
    }

    return postCreationResponse.Post || postCreationResponse.post;
  }

  async function uploadPostImage(postId) {
    if (!selectedImage) return;

    const uploadRes = await fetch(
      apiUrl(`/api/user/upload/post/image?postid=${postId}`),
      {
        method: "POST",
        headers: {
          Authorization: token,
          "Content-Type": selectedImage.type || "application/octet-stream",
        },
        body: selectedImage,
      },
    );

    const uploadResponse = await uploadRes.json();

    if (!uploadRes.ok || !uploadResponse.Ok) {
      throw new Error(
        uploadResponse.Error ||
          uploadResponse.Status ||
          "failed to upload image",
      );
    }
  }

  async function handleFeedPostSubmit(e) {
    e.preventDefault();

    if (!isFormValid) {
      setPostCreationErr(
        "Please fill in all fields (Title and at least 10 characters for body)",
      );
      return;
    }

    if (!token) {
      setPostCreationErr("You must be signed in to create a post.");
      return;
    }

    setIsSubmitting(true);
    setPostCreationErr("");

    try {
      const createdPost = await createPost();
      const createdPostId = createdPost?.id ?? createdPost?.ID;

      if (!createdPostId) {
        throw new Error("post was created, but no post id was returned");
      }

      await uploadPostImage(createdPostId);
      await addDelay(900);
      navigateBackToFeed("/");
    } catch (err) {
      setPostCreationErr(err.message || "failed to create post");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="createpost_wrapper">
      <form className="createpost_form" onSubmit={handleFeedPostSubmit}>
        <div className="createpost_hero">
          <p className="createpost_kicker">New post</p>
          <h1>Share a moment with an image-first post</h1>
          <p>
            Write the caption, pick your image, and the app will create the post
            first, then upload the file using the returned post id.
          </p>
        </div>

        <label className="createpost_label" htmlFor="title">
          Title
        </label>
        <input
          className="createpost_input"
          type="text"
          name="title"
          id="title"
          placeholder="i.e. felt cute to post"
          maxLength={200}
          value={form.title}
          onChange={handleFormChange}
          min={1}
          aria-label="enter title"
        />

        <label className="createpost_label" htmlFor="body">
          Post Body
        </label>
        <textarea
          className="createpost_input createpost_textarea"
          name="body"
          id="body"
          maxLength={200}
          value={form.body}
          onChange={handleFormChange}
          minLength={10}
          aria-label="enter post body"
          placeholder="What's on your mind?"
          required
        />

        <div className="createpost_image-card">
          <div className="createpost_image-copy">
            <label
              className="createpost_label createpost_label-inline"
              htmlFor="file"
            >
              Cover image
            </label>
            <p className="createpost_helper-text">
              Add a JPEG, PNG, or WEBP image up to 2 MB. The image uploads after
              the post is created.
            </p>
          </div>

          <label className="createpost_upload-zone" htmlFor="file">
            <input
              type="file"
              id="file"
              accept="image/jpeg,image/png,image/webp"
              onChange={handleImageChange}
              aria-label="upload post image"
              className="createpost_file-input"
            />
            <span className="createpost_upload-title">
              {selectedImage ? "Change image" : "Drop or choose an image"}
            </span>
            <span className="createpost_upload-subtitle">
              {selectedImage
                ? selectedImageSummary
                : "Tap to browse or drag a file here"}
            </span>
          </label>

          {selectedImagePreview && (
            <div className="createpost_image-preview">
              <img src={selectedImagePreview} alt="Selected post preview" />
              <button
                type="button"
                className="createpost_clear-image"
                onClick={clearSelectedImage}
              >
                Remove image
              </button>
            </div>
          )}
        </div>

        <button
          className="createpost_button"
          type="submit"
          disabled={!isFormValid || isSubmitting}
        >
          {isSubmitting ? "Publishing..." : "Create Post"}
        </button>

        {postCreationErr && (
          <div className="createpost_error">❌ {postCreationErr}</div>
        )}
      </form>
    </div>
  );
}
