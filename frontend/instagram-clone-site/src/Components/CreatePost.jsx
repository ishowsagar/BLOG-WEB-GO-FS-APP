import { useState } from "react";
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
  const { postData, setPostData } = usePostContext();
  const naviagteBackToFeed = useNavigate();
  //   console.log(postData);

  function handleFormChange(e) {
    const { name, value, type, files } = e.target;    
      setForm((prev) => ({
        ...prev,
        [name]: value,
      }));

  }

  function handleFeedPostSubmit(e) {
    e.preventDefault();
    // * resetting form back to default
    setPostData((prev) => [form, ...prev]);
    setForm(INITIAL_FEED_POST_FORM);
    naviagteBackToFeed("/");
    // console.log("submitting form...");

    // ! usenaviate or navigate user back to home once submitted
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
          maxLength={20}
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
          Blog Body
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

        <button className="createpost_button" type="submit">
          Create Blog
        </button>
      </form>
    </div>
  );
}
