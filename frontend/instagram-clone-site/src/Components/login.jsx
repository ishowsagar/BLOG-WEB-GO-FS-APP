import {
  useLocation,
  useNavigate,
  Link,
  redirect,
  createCookie,
  NavLink,
} from "react-router-dom";
import { useState } from "react";

export function Login() {
  //  form intial empty values
  const initialFormData = {
    email: "",
    password: "",
  };
  const [form, setForm] = useState(initialFormData);
  const navigateClient = useNavigate();
  const [errorExists, setErrorExists] = useState(false); //! for conditionally rendering something if error exists
  const [loginErr, setLoginErr] = useState(""); // set err when needed
  const [isSubmitting, setIsSubmitting] = useState(false); // for submission
  const location = useLocation(); //* for checking from where client was coming from
  const whereClientComingFrom = location.state?.from?.pathname || "/";
  //  states
  // const {name,email,password} = form

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
  const isFormValid = form.password.trim() !== "" && form.email.trim() !== "";

  //  handle form submission
  async function handleFormSubmission(event) {
    event.preventDefault();
    setIsSubmitting(true); // ciient started submitting the form

    if (!form.email.trim()) {
      console.error("email is required");
      return;
    }

    if (!form.password.trim()) {
      console.error("password cannot be empty");
      return;
    }

    const payload = {
      method: "POST",
      url: "http://localhost:8080/form/login",
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
        throw new Error("invalid credentials");
      }

      console.log("login data :", data);
      if (data?.token) {
        //  todo - store in cookie instead of ls
        localStorage.setItem("token", data.token); //* set token in lc if req was a success, which is checked by auth if it exists or not in req's lc
      }

      // bug - when user logs in, there mismatch happens of token but that token is valid too
      // test - clear token from current window
      //  test - the check for req.ok *must* be done before setting data returned from resp or it will cause some unxpected issues
      //  fixed - bug has been fixed with these changes
      //  - clearing token before login to prevent mismatch
      //  - checking resp before using returned response

      // if client logged in successfully
      setForm(initialFormData);

      // since token would have been issued for the client -

      // * we set the redirect to the location where user would have asked,not just to "/" feed
      // since client is redirected from auth with location set -> we can fetch that object data where client redierected
      //  since loc object have shared state and that store from - the asked location before auth had called
      // & we set that location to navigate back client & setting 'true' for replace so -> this is where route will be replaced finally
      navigateClient(whereClientComingFrom, {
        replace: true, // replacing current with asked,not this login, only if client had already asked url path that has to relaced otherwise ormaly replace with feed as that is default
      });
      console.log("user request sent successfully for login", data);

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
    <div className="login-form-div">
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
          type="text"
          name="password" // name should be name as state for dynamic form setup
          placeholder="Enter password"
          id="password"
          value={form.password || ""}
          onChange={handleFormChange}
        />

        {/* button for submisson click only */}
        {
          <button
            type="submit"
            className="disabled-btn"
            disabled={!isFormValid}
          >
            {!isSubmitting ? "Login" : "Loggin in..."}
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
              <NavLink to="/password/reset">Forgot Password ?</NavLink>
            </span>
            <Link to="/signup">SignUp</Link>
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
