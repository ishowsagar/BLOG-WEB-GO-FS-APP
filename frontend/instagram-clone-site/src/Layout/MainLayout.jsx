import Header from "../Components/Header";
import Footer from "../Components/Footer";
import Sidebar from "../Components/Sidebar";
import { Outlet, useNavigate } from "react-router-dom";
import { useEffect, useState, createContext, useContext } from "react";
const postDataContext = createContext();

export default function MainLayout() {
  // window.location.reload();
  const [postData, setPostData] = useState([]);
  const [loading, setLoading] = useState(true);

  // const redirectToLoginPageIfTokenNotFound = useNavigate()

  // bug - it was a bug before when we user logged in and feed was giving token mismatch erros
  // fixed by normally fetching token here from ls(local storage) and only setting postData once req was ok not before checking this and reload if token chances to reload flow for rendering feed
  // * get token from localstorage
  const token = localStorage.getItem("token");
  
  // * remember - on fresh login,token gets updated too

  useEffect(() => {
    console.log("MainLayout token:", token);
    if (!token) {
      // token missing: stop loading and allow UI to render (or redirect to login)
      console.warn("No token found in localStorage (key: 'token')");
      setLoading(false);
      return;
    }
    const payloadBody = {
      url: "http://localhost:8080/api/feed",
      Method: "GET",
      //* this is working now, it is fetching token from local storage
      AuthorizationHeaderToken: token,
    };
    async function fetchData() {
      try {
        const response = await fetch(payloadBody.url, {
          method: payloadBody.Method,
          // ! this is how you send token in header from client frontend
          headers: { Authorization: payloadBody.AuthorizationHeaderToken },
        });
        const data = await response.json();
        console.log("feed status:", response.status, data);
        console.log(payloadBody.AuthorizationHeaderToken);
        // const returnPostIfExists = Array.isArray(postData) ? postData : []
        if (!response.ok) {
          throw new Error(data?.Status || data?.error || "error fetching data");
        }
        setPostData(data.Post);
        setLoading(false);
        
        //  reload page, fetch current url from window{browser}.Location
      } catch (err) {
        console.log(err);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [token]);

  if (loading) {
    return (
      <div
        style={{ textAlign: "center", marginTop: "4rem", fontSize: "1.5rem" }}
      >
        Loading...
      </div>
    );
  }
  return (
    <>
      <postDataContext.Provider value={{ postData, setPostData }}>
        <section className="Site_wrapper">
          <Header />
          <div className="Site_content">
            <div className="Site_sidebar">
              <Sidebar />
            </div>
            <main>
              <Outlet />
            </main>
          </div>

          <Footer />
        </section>
      </postDataContext.Provider>
    </>
  );
}
export const usePostContext = () => useContext(postDataContext);
