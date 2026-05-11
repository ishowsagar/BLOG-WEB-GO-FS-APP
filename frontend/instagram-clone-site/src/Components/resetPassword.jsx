import {
  useLocation,
  useNavigate,
  Link,
  redirect,
  createCookie,
  NavLink,
} from "react-router-dom";
import { useEffect, useState } from "react";

export function ResetPassword() {
  //  form intial empty values
  const initialFormData = {
    email: "",
    password: "",
    captchaInput: "",
  };
  const [form, setForm] = useState(initialFormData);
  const navigateClient = useNavigate();
  const [errorExists, setErrorExists] = useState(false); //! for conditionally rendering something if error exists
  const [loginErr, setLoginErr] = useState(""); // set err when needed
  const [isSubmitting, setIsSubmitting] = useState(false); // for submission
  const [captchaValue, setCaptchaValue] = useState("");

  //  states
  const navigate = useNavigate();
  // const {name,email,password} = form

  function generateCaptcha() {
    const characters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
    const captchaLength = 6;
    let nextCaptcha = "";

    for (let index = 0; index < captchaLength; index += 1) {
      const randomIndex = Math.floor(Math.random() * characters.length);
      nextCaptcha += characters[randomIndex];
    }

    return nextCaptcha;
  }

  useEffect(() => {
    setCaptchaValue(generateCaptcha());
  }, []);

  // handle form
  function handleFormChange(event) {
    const { name, value } = event.target;
    setForm((prevFormState) => ({
      // keep old state but update the current field
      ...prevFormState,
      [name]: value,
    }));
  }

  // form validation
  const isFormValid =
    form.password.trim() !== "" &&
    form.email.trim() !== "" &&
    form.captchaInput.trim() !== "";

  const isCaptchaValid =
    form.captchaInput.trim().toUpperCase() === captchaValue;

  //  handle form submission
  async function handleFormSubmission(event) {
    event.preventDefault();
    setIsSubmitting(true); // ciient started submitting the form

    if (!isCaptchaValid) {
      setErrorExists(true);
      setLoginErr("Captcha does not match");
      setIsSubmitting(false);
      setCaptchaValue(generateCaptcha());
      setForm((prevFormState) => ({
        ...prevFormState,
        captchaInput: "",
      }));
      return;
    }

    if (!form.email.trim()) {
      console.error("email is required");
      setIsSubmitting(false);
      return;
    }

    if (!form.password.trim()) {
      console.error("password cannot be empty");
      setIsSubmitting(false);
      return;
    }

    const payload = {
      method: "POST",
      url: "http://localhost:8080/form/password/reset",
    };

    localStorage.removeItem("token");

    try {
      const loginReq = await fetch(payload.url, {
        method: payload.method,
        headers: {
          "Content-Type": "application/json",
        },
        // must send stringified data
        body: JSON.stringify(form), // json formatted struct type data ingres for server
      });

      const data = await loginReq.json();
      if (!loginReq.ok) {
        throw new Error(data.Status);
      }

      console.log("login data :", data);

      // if client logged in successfully
      setForm(initialFormData);
      setCaptchaValue(generateCaptcha());

      navigate("/login");
      console.log("password resetted successfully", data);

      // navigate client to the feed page
    } catch (err) {
      // catch's err is an object, it stores caught err in its message
      console.log(err);
      setErrorExists(true);
      setIsSubmitting(false);
      setLoginErr("Invalid credentials"); //& setting err from backend res
    }
  }

  return (
    <div className="reset-password-form-div">
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

        {/* <label htmlFor="name">Name</label>
        <input
          type="text"
          name="name" // name should be name as state for dynamic form setup
          placeholder="e.g john cena"
          id="name"
          value={form.name || ""}
          onChange={(event) => handleFormChange(event)}
        ></input> */}

        <label htmlFor="password">Password</label>
        <input
          type="password"
          name="password" // name should be name as state for dynamic form setup
          placeholder="Enter password"
          id="password"
          value={form.password || ""}
          onChange={handleFormChange}
        />

        <div className="captcha-box" aria-label="captcha challenge">
          <div className="captcha-box__label">Captcha</div>
          <div className="captcha-box__value">{captchaValue}</div>
          <button
            type="button"
            className="captcha-box__refresh"
            onClick={() => {
              setCaptchaValue(generateCaptcha());
              setForm((prevFormState) => ({
                ...prevFormState,
                captchaInput: "",
              }));
              setErrorExists(false);
              setLoginErr("");
            }}
          >
            Refresh captcha
          </button>
        </div>

        <label htmlFor="captchaInput">Enter Captcha</label>
        <input
          type="text"
          name="captchaInput"
          placeholder="Type the captcha above"
          id="captchaInput"
          value={form.captchaInput || ""}
          onChange={handleFormChange}
          autoComplete="off"
        />

        {/* button for submisson click only */}
        {
          <button
            type="submit"
            className="disabled-btn"
            disabled={!isFormValid || !isCaptchaValid}
          >
            {!isSubmitting ? "Reset Password" : "almost there..."}
          </button>
        }
        {/* ~ conditionally render the error to let client know somethinig was real bad */}
        {errorExists ? (
          <>
            <p
              style={{
                color: "red",
                fontWeight: "bold",
                fontFamily: "sans-serif",
              }}
            >
              {loginErr}
            </p>
            <span
              style={{
                color: "blue",
                fontWeight: "400",
                fontFamily: "sans-serif",
                fontSize: "27px",
              }}
            >
              <NavLink to="/reset/password">Forgot Password ?</NavLink>
            </span>
            <Link to="/signup">Signup</Link>
          </>
        ) : (
          <div>
            <p>
              Wait a minute,don't have an account ?...
              <Link
                to="/signup"
                style={{
                  color: "darkblue",
                  fontWeight: "bolder",
                  fontSize: "20px",
                }}
              >
                Signup
              </Link>
            </p>
          </div>
        )}
      </form>
    </div>
  );
}
