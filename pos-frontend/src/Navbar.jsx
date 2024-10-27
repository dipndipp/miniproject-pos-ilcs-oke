import { Link, useNavigate, useLocation } from "react-router-dom";
import useAuth from "./context/UseAuth"; // Gunakan file baru

const Navbar = () => {
  const { logout, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  if (location.pathname === "/login") {
    return null;
  }

  return (
    <nav className="navbar">
      <div className="left-section">
        <div className="user-mark">
          <Link to={user && user.role === "admin" ? "/admincontrol" : "#"}>
            <span>{user ? user.username : "User"}</span>
          </Link>
        </div>
        <ul className="nav-links">
          {user && user.role === "admin" && (
            <li>
              <Link to="/products">Edit Products</Link>
            </li>
          )}
          <li>
            <Link to="/">Cashier</Link>
          </li>
          <li>
            <Link to="/active-orders">Orders</Link>
          </li>
          <li>
            <Link to="/dashboard">Dashboard</Link>
          </li>
        </ul>
      </div>

      <div className="right-section">
        <div className="logout-button">
          <button onClick={handleLogout} className="btn btn-danger">
            Logout
          </button>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
