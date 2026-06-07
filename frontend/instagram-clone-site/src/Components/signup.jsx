import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { apiUrl } from "../Services/apiConfig";

export function Signup() {
  // form initial empty values
  const initialFormData = {
    name: "",
    email: "",
    password: "",
    bio: "",
    nickname: "",
    username: "",
  };

  const [form, setForm] = useState(initialFormData);
  const [signUp, setSignedup] = useState(false);
  const [signupErr, setSignupErr] = useState(null);
  const [step, setStep] = useState(1);

  const redirectToLoginPage = useNavigate();

  // Step-level validation
  const isStep1Valid = form.email.trim() !== "" && form.password.trim() !== "";
  const isStep2Valid = form.name.trim() !== "" && form.username.trim() !== "";
  const isStep3Valid = form.nickname.trim() !== "";

  // The overall form is valid only if all steps are valid
  const isFormValid = isStep1Valid && isStep2Valid && isStep3Valid;

  // Handle input changes dynamically
  function handleFormChange(event) {
    const { name, value } = event.target;
    setForm((prevFormState) => ({
      ...prevFormState,
      [name]: value,
    }));
  }

  // Go to next step
  function handleFormNext() {
    setStep((prevStep) => prevStep + 1);
  }

  // Go to previous step
  function handleFormBack() {
    setStep((prevStep) => prevStep - 1);
  }

  // Handle form submission
  async function handleFormSubmission(event) {
    event.preventDefault();
    setSignedup(true);

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

      // If signup was successful
      setForm(initialFormData);
      setSignedup(false);
      console.log("user signup successfully:", form);
      redirectToLoginPage("/login");
    } catch (err) {
      console.error(err);
      setSignedup(false);
      setSignupErr(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <div className="sign-up-form-div">
      <form onSubmit={handleFormSubmission}>
        {/* Step 1: Email & Password */}
        {step === 1 && (
          <>
            <label htmlFor="email">Email</label>
            <input
              type="email"
              name="email"
              placeholder="some@gmail.com"
              id="email"
              value={form.email || ""}
              onChange={handleFormChange}
            />
            <label htmlFor="password">Password</label>
            <input
              type="password"
              name="password"
              placeholder="Enter your password"
              id="password"
              value={form.password || ""}
              onChange={handleFormChange}
            />
            <button
              type="button"
              onClick={handleFormNext}
              disabled={!isStep1Valid}
              className="disabled-btn"
            >
              Next
            </button>
          </>
        )}

        {/* Step 2: Name & Username */}
        {step === 2 && (
          <>
            <label htmlFor="name">Name</label>
            <input
              type="text"
              name="name"
              placeholder="e.g. john cena"
              id="name"
              value={form.name || ""}
              onChange={handleFormChange}
            />
            <label htmlFor="username">Username</label>
            <input
              type="text"
              name="username"
              placeholder="e.g. biker_rider..."
              id="username"
              value={form.username || ""}
              onChange={handleFormChange}
            />
            <div className="step-navigation-buttons">
              <button
                type="button"
                onClick={handleFormBack}
                className="secondary-btn"
              >
                Back
              </button>
              <button
                type="button"
                onClick={handleFormNext}
                disabled={!isStep2Valid}
                className="disabled-btn"
              >
                Next
              </button>
            </div>
          </>
        )}

        {/* Step 3: Nickname & Bio */}
        {step === 3 && (
          <>
            <label htmlFor="nickname">Nickname</label>
            <input
              type="text"
              name="nickname"
              placeholder="e.g. ghostRider🚲💨..."
              id="nickname"
              value={form.nickname || ""}
              onChange={handleFormChange}
            />
            <label htmlFor="bio">Bio</label>
            <input
              type="text"
              name="bio"
              placeholder="e.g. cool guy on the earth..."
              id="bio"
              value={form.bio || ""}
              onChange={handleFormChange}
            />
            <div className="step-navigation-buttons">
              <button
                type="button"
                onClick={handleFormBack}
                className="secondary-btn"
              >
                Back
              </button>
              <button
                className="disabled-btn"
                type="submit"
                disabled={!isFormValid || signUp}
              >
                {!signUp ? "Sign Up" : "almost there...."}
              </button>
            </div>
          </>
        )}

        {/* Error / Login Link Section */}
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

