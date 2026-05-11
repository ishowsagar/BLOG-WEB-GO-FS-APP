import { Navigate, Outlet, redirect, useLocation } from "react-router-dom";

// RequireAuth: wrap protected routes with this component.
// Usage (react-router v6):
// <Route path="/profile" element={<RequireAuth><Profile/></RequireAuth>} />

export default function RequireAuth() {
  const token = localStorage.getItem("token");
  const location = useLocation(); //* from where client is coming from

  // If no token, redirect to login and preserve attempted location
  if (!token || token === "") {
    // & sending location state object data -> which holds original asked client location in state.form
    // * setting it to replace with whatever -> would be asked to replace route
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Optionally: validate token shape/expiry here or via an API call
  // For now, treat presence as authenticated and render children
  return <Outlet />;
}
