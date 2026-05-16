import React, { useState } from "react";
import { nanoid } from "nanoid";
import { useNavigate } from "react-router-dom";
import { usePostContext } from "../Layout/MainLayout";
// const INITIAL_FEED_POST_FORM = {
//   name: "",
//   title: "",
//   body: "",
//   image: "",
//   id: null,
// };

const INITIAL_FEED_POST_FORM = {
  title: "",
  body: "",
};

//  setting intial form to be default empty form
export default function CreatePost() {
  const [form, setForm] = useState(INITIAL_FEED_POST_FORM);
  //  ! fetching data from all postData containing Arr
  // const { postBatch, setPostBatch } = usePostContext();
  const [postCreationErr, setPostCreationErr] = useState(""); // state that holds err seamlessly
  const [hasSubmitted, setHasSubmitted] = useState(false); // state that holds if post created successfully
  const [isFormValid, setIsFormValid] = useState(false);
  const naviagteBackToFeed = useNavigate();
  //   console.log(postData);
  const token = localStorage.getItem("token");

  // create post
  async function CreatePost() {
    // false state on startup
    const payload = {
      // url depends on cursor -> if it has last batch post ID -> if yes with cursor&limit otherwise just with limit
      // "GET" - /api/feed?limit=X{limitOffset}&cursor=Y{lastCursor} - return feed[],cursor,hasMore
      url: `http://localhost:8080/api/post/create`,
      header: {
        Authorization: token,
      },
      method: "POST",
      body: {
        title: form.title,
        content: form.body,
      },
    };
    try {
      const postReq = await fetch(payload.url, {
        method: payload.method,
        headers: payload.header,
        body: JSON.stringify(payload.body),
      });

      console.log("uploading post...");
      const postCreationResponse = await postReq.json();
      console.log("response :", postCreationResponse);
      //err check
      if (!postCreationResponse.Ok || !postReq.ok) {
        throw new Error(res.Status || "failed to upload post...");
      }

      //if it was a success call ✅✅
      setHasSubmitted(true);
      setPostCreationErr("");
    } catch (err) {
      // all errors throws are caught here and set from here
      setPostCreationErr(err.message | "failed to create post");
      console.log(err);
    }
  }

  function handleFormChange(e) {
    const { name, value, type, files } = e.target;
    setForm((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Real-time validation
    validateForm({ ...form, [name]: value });
  }

  function validateForm(formData) {
    const isTitleValid = formData.title.trim().length > 0;
    const isBodyValid = formData.body.trim().length >= 10;
    setIsFormValid(isTitleValid && isBodyValid);
  }

  // add some delay in between - keep other things in await too for full effect nurishment
  async function addDelay(ms) {
    return new Promise((res) => setTimeout(res, ms));
  }

  async function handleFeedPostSubmit(e) {
    e.preventDefault();

    // Validate before submitting
    if (!isFormValid) {
      setPostCreationErr(
        "Please fill in all fields (Title and at least 10 characters for body)",
      );
      return;
    }

    console.log("submitting post");
    setHasSubmitted(true);

    await CreatePost();
    // add some delay before navigating in seconds to the home feed
    await addDelay(1000);
    naviagteBackToFeed("/"); // send client to home feed
  }

  return (
    <div className="createpost_wrapper">
      <form className="createpost_form" onSubmit={handleFeedPostSubmit}>
        {/* <label className="createpost_label" htmlFor="username">
          Username
        </label>
        <input
          className="createpost_input"
          type="text"
          name="username"
          id="username"
          placeholder="i.e @ishowdenver"
          maxLength={15}
          value={form.username}
          onChange={handleFormChange}
          min={6}
          required
          aria-label="enter your username" */}
        {/* /> */}

        <label className="createpost_label" htmlFor="title">
          Title
        </label>
        <input
          className="createpost_input"
          type="text"
          name="title"
          id="title"
          placeholder="i.e felt cute to post!"
          maxLength={200}
          value={form.title || ""}
          onChange={handleFormChange}
          min={1}
          aria-label="enter title"
        />

        {/* <label className="createpost_label" htmlFor="image">
          Image Upload
        </label>
        <input
          className="createpost_input"
          type="file"
          name="image"
          id="image"
          accept="image/*"
          onChange={handleFormChange}
          aria-label="upload post image"
        /> */}

        {/* {form.image && (
          <div className="createpost_image-preview">
            <img
              src={form.image}
              alt="Preview"
              style={{ maxWidth: "100%", maxHeight: "200px" }}
            />
          </div>
        )} */}

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

        <button
          className="createpost_button"
          type="submit"
          disabled={!isFormValid || hasSubmitted}
        >
          {hasSubmitted ? "uploading...." : `Create Post`}
        </button>
        {/* conditionally render this div if hit any err during post creation */}
        {postCreationErr && (
          <div
            style={{
              color: "red",
              padding: "0.5rem",
              textAlign: "center",
              fontSize: "0.9rem",
            }}
          >
            ❌ {postCreationErr}
          </div>
        )}
      </form>
    </div>
  );
}
