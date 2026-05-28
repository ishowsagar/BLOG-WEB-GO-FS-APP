import { useState } from "react";
import { Link, redirect, useNavigate } from "react-router-dom";
import { apiUrl } from "../Services/apiConfig";

export function Signup() {
  //  form intial empty values
  const initialFormData = {
    name: "",
    email: "",
    password: "",
    // adding newly added fields so it includes in user struct
    bio: "",
    nickname: "",
    username: "",
  };
  const [form, setForm] = useState(initialFormData);
  const [signUp, setSignedup] = useState(false);
  const [errorExists, setErrorExists] = useState(false); // for c.r (conditionally rendering during err,if app hit any)
  const [signupErr, setSignupErr] = useState(null);

  // form validation from client side that form has been filled before signup button to be clicked
  //  * since everthing is stored in form state, we can check on that directly - checking if these are non-empty
  const isFormValid =
    form.name.trim() !== "" &&
    form.password.trim() !== "" &&
    form.email.trim() !== "" &&
    form.nickname.trim() !== "" &&
    form.username.trim() !== "";

  //  states
  // const {name,email,password} = form
  const redirectToLoginPage = useNavigate();

  // handle form
  function handleFormChange(event) {
    event.preventDefault();
    const { name, value } = event.target;

    //todo -  client side validation needed

    setForm((prevFormState) => ({
      //  keep old states as it is but change current target name's field state with value
      ...prevFormState,
      [name]: value,
    }));

    // console.log(`form filled successfully - ${form}`);
  }

  //  handle form submission
  async function handleFormSubmission(event) {
    //  uncommented so it refreshes site to get token
    event.preventDefault();
    setSignedup(true); // set to true when client hit sign up button to contionally render text

    try {
      const res = await fetch(apiUrl("/form/register"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "failed to signup");
      }

      // if sign up was a successfull operation
      setForm(initialFormData);
      setSignedup(false);
      console.log("user signup successfully :", form);
      redirectToLoginPage("/login");
    } catch (err) {
      console.error(err);
      setErrorExists(true);
      setSignedup(false);
      setSignupErr(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <div className="sign-up-form-div">
      <form onSubmit={handleFormSubmission}>
        <label htmlFor="email">Email</label>
        <input
          type="email"
          name="email" // name should be name as state for dynamic form setup
          placeholder="some@gmail.com"
          id="email"
          value={form.email || ""}
          onChange={(event) => handleFormChange(event)}
        ></input>
        <label htmlFor="name">Name</label>
        <input
          type="text"
          name="name" // name should be name as state for dynamic form setup
          placeholder="e.g john cena"
          id="name"
          value={form.name || ""}
          onChange={(event) => handleFormChange(event)}
        ></input>
        <label htmlFor="password">Password</label>
        <input
          type="password"
          name="password" // name should be name as state for dynamic form setup
          placeholder=""
          id="password"
          value={form.password || ""}
          onChange={(event) => handleFormChange(event)}
        ></input>

        {/* * adding new fields for newly added fields in user struct */}
        <label htmlFor="username">Username</label>
        <input
          type="username"
          name="username" // name should be name as state for dynamic form setup
          placeholder="e.g biker_rider..."
          id="username"
          value={form.username || ""}
          onChange={(event) => handleFormChange(event)}
        ></input>
        <label htmlFor="nickname">Nickname</label>
        <input
          type="nickname"
          name="nickname" // name should be name as state for dynamic form setup
          placeholder="e.g ghostRider🚲💨..."
          id="nickname"
          value={form.nickname || ""}
          onChange={(event) => handleFormChange(event)}
        ></input>
        <label htmlFor="bio">Bio</label>
        <input
          type="bio"
          name="bio" // name should be name as state for dynamic form setup
          placeholder="e.g cool guy on the earth..."
          id="bio"
          value={form.bio || ""}
          onChange={(event) => handleFormChange(event)}
        ></input>

        {/* button for submisson click only, render only if form is fully filled */}
        {
          <button
            className="disabled-btn"
            type="submit"
            disabled={!isFormValid || signUp}
          >
            {!signUp ? "Sign Up" : "almost there...."}
          </button>
        }
        {/* && for if this true && do this, no need for ternary operation */}

        {/* ! error hit section only */}
        {signupErr ? (
          <p>{signupErr}</p>
        ) : (
          <p>
            Already a user?...
            <Link
              to="/login"
              style={{
                color: "darkblue",
                fontWeight: "bold",
                fontSize: "20px",
                cursor: "pointer",
              }}
            >
              Login
            </Link>
          </p>
        )}
      </form>
    </div>
  );
}
