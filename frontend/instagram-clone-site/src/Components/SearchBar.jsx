import React, { useState } from "react";
import SearchIcon from "../assets/icons/search.png";
import "./SearchBar.css";

export default function SearchBar() {
  const [searchQuery, setSearchQuery] = useState("");
  const [isSearchFocused, setIsSearchFocused] = useState(false);

  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
    // TODO: Implement search logic here
  };

  const handleSearchFocus = () => {
    setIsSearchFocused(true);
    // TODO: Show search results/suggestions on focus
  };

  const handleSearchBlur = () => {
    setIsSearchFocused(false);
    // TODO: Hide search results on blur
  };

  const handleClearSearch = () => {
    setSearchQuery("");
    // TODO: Clear search results
  };

  return (
    <div className="searchbar_wrapper">
      <div className={`searchbar_container ${isSearchFocused ? "focused" : ""}`}>
        <img src={SearchIcon} alt="search" className="searchbar_icon" />
        <input
          type="text"
          placeholder="Search posts, users..."
          className="searchbar_input"
          value={searchQuery}
          onChange={handleSearchChange}
          onFocus={handleSearchFocus}
          onBlur={handleSearchBlur}
        />
        {searchQuery && (
          <button className="searchbar_clear" onClick={handleClearSearch}>
            ✕
          </button>
        )}
      </div>
      {/* TODO: Add search results dropdown here */}
    </div>
  );
}
