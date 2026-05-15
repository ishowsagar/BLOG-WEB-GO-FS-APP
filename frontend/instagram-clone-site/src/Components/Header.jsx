import logo from "../assets/icon.png";
export default function Header() {
  return (
    <header>
      <div className="header_container">
        <div className="header_content">
          <div className="header_brand header_brand_centered">
            <img src={logo} alt="Instagram logo" className="header_logo" />
            <h1 className="header_title">Instagram</h1>
          </div>
        </div>
      </div>
    </header>
  );
}
