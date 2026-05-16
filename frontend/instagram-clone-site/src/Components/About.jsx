import React from "react";
import { Link } from "react-router-dom";

export default function About() {
  const features = [
    {
      icon: "📱",
      title: "Feed with Infinite Scroll",
      description:
        "Cursor-based pagination with batch loading for smooth infinite feed",
    },
    {
      icon: "💬",
      title: "Comments & Interactions",
      description:
        "Real-time comments with user avatars, timestamps, and reply functionality",
    },
    {
      icon: "❤️",
      title: "Like System",
      description: "One-click post likes with live like count updates",
    },
    {
      icon: "👥",
      title: "Follow & Profiles",
      description:
        "Follow users, view profiles with post archives and follower counts",
    },
    {
      icon: "🔍",
      title: "User Search",
      description: "Fast user search with filter options and trending results",
    },
    {
      icon: "✉️",
      title: "Direct Messaging",
      description:
        "Real-time WebSocket messaging with active user notifications",
    },
    {
      icon: "📝",
      title: "Create Posts",
      description: "Publish posts with titles and content to your profile",
    },
    {
      icon: "🎥",
      title: "Reels & Explore",
      description: "Discover trending content and explore user-generated media",
    },
    {
      icon: "📖",
      title: "Stories",
      description:
        "View user stories with green Instagram-style outline avatars",
    },
    {
      icon: "🔐",
      title: "Auth & Security",
      description: "Secure login with JWT tokens and protected routes",
    },
  ];

  const techStack = {
    frontend: ["React", "React Router v6", "CSS3 with @layer", "WebSockets"],
    backend: ["Go", "SQL Database", "Redis Cache", "JWT Auth", "RESTful API"],
    features: [
      "Cursor-based Pagination",
      "Real-time WebSockets",
      "Rate Limiting",
      "Request Logging",
    ],
  };

  const appRoutes = [
    { path: "/", name: "Login Page" },
    { path: "/signup", name: "Sign Up" },
    { path: "/feed", name: "Home Feed" },
    { path: "/feed/:id", name: "Single Post View" },
    { path: "/profile", name: "My Profile" },
    { path: "/profile/:userid", name: "User Profile" },
    { path: "/search", name: "Search Users" },
    { path: "/explore", name: "Explore" },
    { path: "/reels", name: "Reels" },
    { path: "/create", name: "Create Post" },
    { path: "/messages", name: "Direct Messages" },
  ];

  return (
    <div className="about_page">
      <div className="about_container">
        {/* Hero Section */}
        <section className="about_hero">
          <div className="about_hero_content">
            <h1 className="about_title">Instagram Clone</h1>
            <p className="about_subtitle">
              A beautiful, modern social media platform built with React and Go
            </p>
            <p className="about_tagline">
              Experience seamless social connectivity with real-time messaging,
              infinite feeds, and stunning design
            </p>
          </div>
        </section>

        {/* Features Section */}
        <section className="about_section">
          <h2 className="about_section_title">✨ Features</h2>
          <div className="about_features_grid">
            {features.map((feature, idx) => (
              <div key={idx} className="about_feature_card">
                <div className="about_feature_icon">{feature.icon}</div>
                <h3 className="about_feature_title">{feature.title}</h3>
                <p className="about_feature_desc">{feature.description}</p>
              </div>
            ))}
          </div>
        </section>

        {/* Tech Stack Section */}
        <section className="about_section">
          <h2 className="about_section_title">🛠️ Tech Stack</h2>
          <div className="about_tech_grid">
            <div className="about_tech_card">
              <h3 className="about_tech_category">Frontend</h3>
              <ul className="about_tech_list">
                {techStack.frontend.map((tech, idx) => (
                  <li key={idx}>
                    <span className="about_tech_dot">•</span> {tech}
                  </li>
                ))}
              </ul>
            </div>
            <div className="about_tech_card">
              <h3 className="about_tech_category">Backend</h3>
              <ul className="about_tech_list">
                {techStack.backend.map((tech, idx) => (
                  <li key={idx}>
                    <span className="about_tech_dot">•</span> {tech}
                  </li>
                ))}
              </ul>
            </div>
            <div className="about_tech_card">
              <h3 className="about_tech_category">Advanced Features</h3>
              <ul className="about_tech_list">
                {techStack.features.map((feature, idx) => (
                  <li key={idx}>
                    <span className="about_tech_dot">•</span> {feature}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </section>

        {/* Routes Section */}
        <section className="about_section">
          <h2 className="about_section_title">🗺️ App Routes</h2>
          <div className="about_routes_container">
            <div className="about_routes_grid">
              {appRoutes.map((route, idx) => (
                <div key={idx} className="about_route_item">
                  <span className="about_route_path">{route.path}</span>
                  <span className="about_route_name">{route.name}</span>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Design Section */}
        <section className="about_section">
          <h2 className="about_section_title">🎨 Design Philosophy</h2>
          <div className="about_design_content">
            <p className="about_design_text">
              Every pixel is crafted with attention to detail. We've built a
              modern, responsive interface with:
            </p>
            <ul className="about_design_list">
              <li>🌊 Cool blue gradients for a fresh, modern aesthetic</li>
              <li>🎯 Dark navy gradient text for superior readability</li>
              <li>✨ Smooth animations and micro-interactions</li>
              <li>📱 Fully responsive design across all devices</li>
              <li>♿ Accessible color contrast and semantic HTML</li>
              <li>🚀 Optimized performance with lazy loading</li>
            </ul>
          </div>
        </section>

        {/* CTA Section */}
        <section className="about_cta_section">
          <h2 className="about_cta_title">Ready to Connect?</h2>
          <div className="about_cta_buttons">
            <Link to="/" className="about_cta_button about_cta_primary">
              Explore Feed
            </Link>
            <Link to="/create" className="about_cta_button about_cta_secondary">
              Create Post
            </Link>
          </div>
        </section>
      </div>
    </div>
  );
}
