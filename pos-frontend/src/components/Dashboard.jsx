import React, { useState, useEffect } from "react";
import axios from "axios";

const Dashboard = () => {
  const [topSellers, setTopSellers] = useState([]); // Ganti null dengan array untuk menyimpan semua top seller
  const [totalRevenue, setTotalRevenue] = useState(0);
  const [productCount, setProductCount] = useState(0); // Mengganti productList dengan productCount

  useEffect(() => {
    // Fetch Top Seller
    const fetchTopSellers = async () => {
      try {
        const response = await axios.get(
          "http://localhost:8080/top-selling-menu"
        );
        console.log("Top Seller Data:", response.data); // Tambahkan log untuk memeriksa data
        setTopSellers(response.data); // Menyimpan semua top seller
      } catch (error) {
        console.error("Error fetching top seller:", error);
      }
    };

    // Fetch Total Revenue
    const fetchTotalRevenue = async () => {
      try {
        const response = await axios.get("http://localhost:8080/total-revenue");
        setTotalRevenue(response.data.total_revenue); // Pastikan key yang benar
      } catch (error) {
        console.error("Error fetching total revenue:", error);
      }
    };

    // Fetch Product Count
    const fetchProductCount = async () => {
      try {
        const response = await axios.get("http://localhost:8080/product-count");
        setProductCount(response.data.product_count); // Mengatur state untuk product count
      } catch (error) {
        console.error("Error fetching product count:", error);
      }
    };

    fetchTopSellers();
    fetchTotalRevenue();
    fetchProductCount();
  }, []);

  return (
    <div className="dashboard">
      <h1>Dashboard</h1>
      <div className="dashboard-grid">
        <div className="dashboard-section">
          <h2>Top Seller</h2>
          {topSellers.length > 0 ? ( // Memeriksa apakah ada top sellers
            <ul>
              {topSellers.map(
                (
                  item,
                  index // Iterasi untuk menampilkan setiap item
                ) => (
                  <li key={index}>
                    <p>
                      <strong>Product Name:</strong> {item.product_name}
                    </p>
                    <p>
                      <strong>Quantity Sold:</strong> {item.total_sold}
                    </p>
                  </li>
                )
              )}
            </ul>
          ) : (
            <p>Loading...</p>
          )}
        </div>

        <div className="dashboard-section">
          <h2>Total Revenue</h2>
          <h1>{`Rp ${totalRevenue.toLocaleString()}`}</h1>
        </div>

        <div className="dashboard-section">
          <h2>Total Products</h2>
          <h1>{productCount}</h1>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
