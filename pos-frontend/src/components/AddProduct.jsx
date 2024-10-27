import { useState } from "react";
import axios from "axios";
import PropTypes from "prop-types";

const AddProduct = ({ onProductAdded }) => {
  const [product, setProduct] = useState({ name: "", price: "" });
  const [selectedFile, setSelectedFile] = useState(null);
  const [error, setError] = useState(""); // State to hold error messages

  const handleChange = (e) => {
    const { name, value } = e.target;
    setProduct({ ...product, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(""); // Reset error state

    // Validate inputs before sending
    if (!product.name || !product.price) {
      setError("Please fill in all fields.");
      return;
    }

    // Additional validation: check if price is a valid positive number
    if (isNaN(product.price) || parseFloat(product.price) <= 0) {
      setError("Please enter a valid positive price.");
      return;
    }

    try {
      const formData = new FormData();
      formData.append("name", product.name);
      formData.append("price", parseFloat(product.price));
      formData.append("image", selectedFile); // Menyertakan file gambar

      await axios.post("http://localhost:8080/create-product", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });
      onProductAdded(); // Call to refresh product list
      alert("Product added successfully");

      // Reset form fields after submission
      setProduct({ name: "", price: "" });
      e.target.reset();
    } catch (error) {
      // More detailed error handling
      if (error.response && error.response.status === 400) {
        setError("Invalid input: " + error.response.data);
      } else if (error.response && error.response.status === 500) {
        setError("Server error. Please try again later.");
      } else {
        setError("Error adding product: " + (error.message || "Unknown error"));
      }
      console.error("Error adding product:", error);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="formproduk">
      <h2>Add Product</h2>
      {error && <p style={{ color: "red" }}>{error}</p>}{" "}
      {/* Display error message */}
      <input
        type="file"
        name="image"
        accept="image/*"
        onChange={(e) => setSelectedFile(e.target.files[0])} // Mengatur file yang dipilih
        required
        className="inputfile"
      />
      <input
        type="text"
        name="name"
        placeholder="Product Name"
        value={product.name}
        onChange={handleChange}
        required
        autoComplete="off"
      />
      <input
        type="number"
        name="price"
        placeholder="Price (IDR)"
        value={product.price}
        onChange={handleChange}
        required
      />
      <button type="submit">Add</button>
    </form>
  );
};

AddProduct.propTypes = {
  onProductAdded: PropTypes.func.isRequired,
};

export default AddProduct;
