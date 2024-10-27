import { useState, useEffect } from "react";
import axios from "axios";
import useAuth from "../context/useAuth"; // Asumsikan Anda memiliki UserContext

const AdminControl = () => {
  const { user } = useAuth(); // Mengambil informasi pengguna dari konteks
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [role, setRole] = useState("cashier");
  const [passwordMatch, setPasswordMatch] = useState(true);
  const [message, setMessage] = useState("");
  const [totalAdmin, setTotalAdmin] = useState(0);
  const [totalCashier, setTotalCashier] = useState(0);

  useEffect(() => {
    // Fetch total admin and cashier
    const fetchCounts = async () => {
      try {
        const response = await axios.get("http://localhost:8080/admin-count");
        setTotalAdmin(response.data.admin_count);

        const cashierResponse = await axios.get(
          "http://localhost:8080/cashier-count"
        );
        setTotalCashier(cashierResponse.data.cashier_count);
      } catch (error) {
        console.error("Error fetching counts:", error);
      }
    };

    fetchCounts();
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setPasswordMatch(false);
      return;
    }

    try {
      const response = await axios.post(
        "http://localhost:8080/create-account",
        {
          username,
          password,
          role,
        }
      );
      setMessage(response.data.message);
      setUsername("");
      setPassword("");
      setConfirmPassword("");
      setRole("cashier");
      setPasswordMatch(true);
    } catch (error) {
      setMessage("Failed to create account");
      console.error("Error creating account:", error);
    }
  };

  const handleConfirmPasswordChange = (e) => {
    setConfirmPassword(e.target.value);
    setPasswordMatch(e.target.value === password);
  };

  return (
    <div className="admin-control-page">
      <h1>Admin Control Page</h1>
      <div className="admin-control-container">
        <div className="create-account-form">
          <h2>Create Account</h2>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="username">Username:</label>
              <input
                type="text"
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="password">Password:</label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="confirmPassword">Confirm Password:</label>
              <input
                type="password"
                id="confirmPassword"
                value={confirmPassword}
                onChange={handleConfirmPasswordChange}
                required
                className={!passwordMatch ? "input-error" : ""}
              />
              {!passwordMatch && (
                <p className="error-message">Passwords do not match!</p>
              )}
            </div>
            <div className="form-group">
              <label htmlFor="role">Role:</label>
              <select
                id="role"
                value={role}
                onChange={(e) => setRole(e.target.value)}
                required
              >
                <option value="kasir">Cashier</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <button type="submit">Create Account</button>
          </form>
          {message && <p>{message}</p>}
        </div>
        <div className="dashboard-admin">
          <h3>Your Profile</h3>
          <ul>
            <li>
              <p>Username: {user.username}</p>
            </li>
            <li>
              <p>Role: {user.role}</p>
            </li>
          </ul>

          <h3>User Data</h3>
          <ul>
            <li>
              <p>Total Admin: {totalAdmin}</p>
            </li>
            <li>
              <p>Total Cashier: {totalCashier}</p>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default AdminControl;
