import { useEffect, useState, useRef } from "react";
import axios from "axios";
import moment from "moment-timezone"; // Import moment-timezone
import { CSSTransition } from "react-transition-group";
import { Link, useLocation } from "react-router-dom";

const ActiveOrders = () => {
  const [orders, setOrders] = useState([]);
  const location = useLocation();
  const nodeRef = useRef(null); // Tambahkan nodeRef untuk CSSTransition
  const [orderCount, setOrderCount] = useState(0); // Mengganti productList dengan productCount

  const formatPrice = (price) => {
    return price.toLocaleString("id-ID", {
      style: "currency",
      currency: "IDR",
    });
  };

  const fetchOrders = async () => {
    try {
      const response = await axios.get("http://localhost:8080/orders");
      console.log("Fetched orders:", response.data);

      if (Array.isArray(response.data)) {
        const sortedOrders = response.data.sort((a, b) => a.id - b.id);
        console.log("Sorted orders:", sortedOrders);
        setOrders(sortedOrders);
      } else {
        console.error("Expected an array but received:", response.data);
        setOrders([]);
      }
    } catch (error) {
      console.error("Error fetching orders:", error);
      alert("Failed to fetch orders. Please try again later.");
    }
  };
  const fetchOrderCount = async () => {
    try {
      const response = await axios.get(
        "http://localhost:8080/onprogress-count"
      );
      setOrderCount(response.data.order_onprogress_count); // Pastikan key yang benar
    } catch (error) {
      console.error("Error fetching total revenue:", error);
    }
  };

  const completeOrder = async (orderId) => {
    try {
      const response = await axios.post(
        `http://localhost:8080/complete-order?id=${orderId}`
      );
      if (response.status === 200) {
        alert("Order Completed!");
        await fetchOrders();
      }
    } catch (error) {
      console.error("Error completing order:", error);
      alert("Failed to complete the order. Please try again.");
    }
  };
  const cancelOrder = async (orderId) => {
    if (window.confirm("Are you sure you want to cancel this order?")) {
      try {
        const response = await axios.post(
          `http://localhost:8080/cancel-order?id=${orderId}`
        );
        if (response.status === 200) {
          alert("Order Canceled!");
          await fetchOrders();
        }
      } catch (error) {
        console.error("Error completing order:", error);
        alert("Failed to complete the order. Please try again.");
      }
    }
  };

  useEffect(() => {
    fetchOrders();
    fetchOrderCount();
  }, []);

  return (
    <CSSTransition
      in={true}
      appear={true}
      timeout={300}
      classNames="fade"
      nodeRef={nodeRef} // Tambahkan ref di sini
    >
      <div ref={nodeRef} className="page-container">
        {" "}
        {/* Tambahkan ref di elemen ini */}
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
        <h1>Active Orders : {orderCount}</h1>
        <ul className="order-container">
          {Array.isArray(orders) && orders.length > 0 ? (
            orders.map((order) => (
              <li key={order.id} className="card-order">
                <div className="order-details">
                  <strong> ORD - {order.id}</strong> <br />
                  Status: {order.status} <br />
                  Order Placed:{" "}
                  {moment(order.created_at)
                    .utc()
                    .format("MMMM Do YYYY, h:mm A")}{" "}
                  <br />
                  <strong>Total Price:</strong> {formatPrice(order.total_price)}{" "}
                  <br />
                  <div className="order-details">
                    <h4>Ordered Products:</h4>
                    {Array.isArray(order.details) &&
                    order.details.length > 0 ? (
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
                </div>
                <div className="buttonorder">
                  <button
                    onClick={() => completeOrder(order.id)}
                    className="button-complete-order"
                  >
                    Complete Order
                  </button>
                  <button
                    onClick={() => cancelOrder(order.id)}
                    className="button-cancel-order"
                  >
                    Cancel Order
                  </button>
                </div>
              </li>
            ))
          ) : (
            <li className="no-orders">No orders available.</li>
          )}
        </ul>
      </div>
    </CSSTransition>
  );
};

export default ActiveOrders;
