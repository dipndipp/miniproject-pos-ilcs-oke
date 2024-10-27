import { useEffect, useState, useRef } from "react";
import axios from "axios";
import moment from "moment-timezone";
import { CSSTransition } from "react-transition-group";
import { Link, useLocation } from "react-router-dom";

const HistoryOrders = () => {
  const [orders, setOrders] = useState([]);
  const location = useLocation();
  const nodeRef = useRef(null); // Tambahkan ref di sini

  const formatPrice = (price) => {
    return price.toLocaleString("id-ID", {
      style: "currency",
      currency: "IDR",
    });
  };

  const fetchOrders = async () => {
    try {
      const response = await axios.get(
        "http://localhost:8080/completed-orders"
      );
      if (response.data && Array.isArray(response.data)) {
        const sortedOrders = response.data.sort((a, b) => a.id - b.id);
        setOrders(sortedOrders);
      } else {
        setOrders([]);
      }
    } catch (error) {
      alert("Failed to fetch completed orders.");
    }
  };

  const deleteOrder = async (orderId) => {
    if (window.confirm("Are you sure you want to delete this order?")) {
      try {
        const response = await axios.delete(
          `http://localhost:8080/delete-order?id=${orderId}`
        );
        if (response.status === 200) {
          alert("Riwayat Pesanan Berhasil Dihapus!");
          fetchOrders();
        }
      } catch (error) {
        alert("Failed to delete the order.");
      }
    }
  };

  useEffect(() => {
    fetchOrders();
  }, []);

  return (
    <CSSTransition
      in={true}
      appear={true}
      timeout={300}
      classNames="fade"
      nodeRef={nodeRef} // Ref yang ditambahkan
    >
      <div className="page-container" ref={nodeRef}>
        {" "}
        {/* Ref pada elemen ini */}
        <ul className="sub-navbar">
          <li
            className={location.pathname === "/active-orders" ? "active" : ""}
          >
            <Link to="/active-orders">Active Orders</Link>
          </li>
          <li
            className={location.pathname === "/history-orders" ? "active" : ""}
          >
            <Link to="/history-orders">History Orders</Link>
          </li>
        </ul>
        <h1>Completed Orders</h1>
        <ul className="order-container">
          {orders.length > 0 ? (
            orders.map((order) => (
              <li key={order.id} className="card-order">
                <div className="order-details">
                  <strong>ORD - {order.id}</strong> <br />
                  <span
                    className={`${
                      order.status === "Order Canceled" ? "order-canceled" : ""
                    }`}
                  >
                    {" "}
                    Status: {order.status}
                  </span>{" "}
                  <br />
                  Order Completed:{" "}
                  {moment(order.created_at)
                    .utc()
                    .format("MMMM Do YYYY, h:mm A")}{" "}
                  <br />
                  <strong>Total Price:</strong> {formatPrice(order.total_price)}{" "}
                  <br />
                </div>

                <div className="order-details">
                  <h4>Ordered Products:</h4>
                  {order.details.length > 0 ? (
                    <ul>
                      {order.details.map((detail, index) => (
                        <li key={index}>
                          {detail.product_name} - x{detail.quantity}
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p>No products found for this order.</p>
                  )}
                </div>

                <button
                  onClick={() => deleteOrder(order.id)}
                  className="button-cancel-order"
                >
                  Delete
                </button>
              </li>
            ))
          ) : (
            <li>No completed orders available.</li>
          )}
        </ul>
      </div>
    </CSSTransition>
  );
};

export default HistoryOrders;
