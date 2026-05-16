export default function Footer() {
  return (
    <footer>
      <div className="footer_container">
        <div className="footer_content">
          <nav className="footer_nav">
            <ul className="footer_nav_items">
              <li>
                <a href="about">About</a>
              </li>
              <li>
                <a href="#">Help</a>
              </li>
              <li>
                <a href="#">Privacy</a>
              </li>
              <li>
                <a href="#">Terms</a>
              </li>
            </ul>
          </nav>
          <div className="footer_text_content">
            <a className="footer_text" href="mailto:ishowdenver@gmail.com">ishowdenver@gmail.com</a>
            <span className="footer_text">• Built by Denver &amp; Co</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
