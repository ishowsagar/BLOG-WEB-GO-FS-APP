import logo from "../assets/icons/denvergram_logo.png";
export default function Header() {
  return (
    <header>
      <div className="header_container">
        <div className="header_content">
          <div className="header_brand header_brand_centered">
            <img src={logo} alt="Denvergram logo" className="header_logo" width="42" height="42" />
            <h1 className="header_title">Denvergram</h1>
          </div>
        </div>
      </div>
    </header>
  );
}
