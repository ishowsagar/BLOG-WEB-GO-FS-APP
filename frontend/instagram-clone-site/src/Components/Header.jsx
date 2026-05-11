import logo from "../assets/icon.png"
export default function Header() {
  return (
    <header>
      <div className="header_container">
        <div className="header_content">
          <img src={logo} className="header_logo"/>
          <h1 className="header_title" >Instagram</h1>
        </div>
      </div>
    </header>
  );
}
