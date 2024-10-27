import { useState } from "react";
import {
  BrowserRouter as Router,
  Route,
  Routes,
  Navigate,
} from "react-router-dom";
import PropTypes from "prop-types";
import Navbar from "./Navbar";
import ProductList from "./components/ProductList";
import ActiveOrders from "./components/ActiveOrders";
import Cashier from "./components/Cashier";
import HistoryOrder from "./components/HistoryOrder";
import EditProduct from "./components/EditProduct";
import Login from "./components/Login";
import { AuthProvider } from "./context/AuthContext";
import useAuth from "./context/useAuth"; // Gunakan file baru
import Dashboard from "./components/Dashboard";
import AdminControl from "./admincomponents/AdminControl";

const PrivateRoute = ({ children }) => {
  const { user } = useAuth();
  return user ? children : <Navigate to="/login" />;
};

const App = () => {
  const [refreshOrders, setRefreshOrders] = useState(false);

  const handleCheckout = () => {
    setRefreshOrders((prev) => !prev);
  };

  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="*"
            element={
              <>
                <Navbar />
                <Routes>
                  <Route
                    path="/"
                    element={
                      <PrivateRoute>
                        <Cashier onCheckout={handleCheckout} />
                      </PrivateRoute>
                    }
                  />
                  <Route
                    path="/products"
                    element={
                      <PrivateRoute>
                        <ProductList />
                      </PrivateRoute>
                    }
                  />
                  <Route
                    path="/dashboard"
                    element={
                      <PrivateRoute>
                        <Dashboard />
                      </PrivateRoute>
                    }
                  />
                  <Route
                    path="/active-orders"
                    element={
                      <PrivateRoute>
                        <ActiveOrders refresh={refreshOrders} />
                      </PrivateRoute>
                    }
                  />

                  <Route
                    path="/history-orders"
                    element={
                      <PrivateRoute>
                        <HistoryOrder />
                      </PrivateRoute>
                    }
                  />
                  <Route
                    path="/edit-product/:id"
                    element={
                      <PrivateRoute>
                        <EditProduct />
                      </PrivateRoute>
                    }
                  />
                  <Route path="*" element={<Navigate to="/" />} />
                  <Route
                    path="/admincontrol"
                    element={
                      <PrivateRoute>
                        <AdminControl />
                      </PrivateRoute>
                    }
                  />
                </Routes>
              </>
            }
          />
        </Routes>
      </Router>
    </AuthProvider>
  );
};
PrivateRoute.propTypes = {
  children: PropTypes.node.isRequired, // Menandakan bahwa children harus ada
};

export default App;
