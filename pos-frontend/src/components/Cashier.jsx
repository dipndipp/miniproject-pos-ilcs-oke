import { useEffect, useState } from "react";
import axios from "axios";
import PropTypes from "prop-types";
import QuantityInput from "./QuantityInput"; // Import QuantityInput component

const Cashier = ({ onCheckout }) => {
  const [products, setProducts] = useState([]);
  const [cart, setCart] = useState([]);
  const [total, setTotal] = useState(0);

  const fetchProducts = async () => {
    try {
      const response = await axios.get("http://localhost:8080/products");
      setProducts(response.data);
    } catch (error) {
      console.error("Error fetching products:", error);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, []);

  const addToCart = (product, quantity) => {
    setCart((prevCart) => {
      const existingProduct = prevCart.find(
        (item) => item.product.id === product.id
      );
      let newCart;
      if (existingProduct) {
        newCart = prevCart.map((item) =>
          item.product.id === product.id
            ? { ...item, quantity: item.quantity + quantity }
            : item
        );
      } else {
        newCart = [...prevCart, { product, quantity }];
      }
      calculateTotal(newCart);
      return newCart;
    });
  };

  const removeFromCart = (productId) => {
    setCart((prevCart) => {
      const newCart = prevCart.filter((item) => item.product.id !== productId);
      calculateTotal(newCart); // Hitung total setelah item dihapus
      return newCart;
    });
  };

  const calculateTotal = (cart) => {
    const totalAmount = cart.reduce(
      (acc, item) => acc + item.product.price * item.quantity,
      0
    );
    setTotal(totalAmount);
  };

  const formatPrice = (price) => {
    return price.toLocaleString("id-ID", {
      style: "currency",
      currency: "IDR",
    });
  };

  const handleCheckout = async () => {
    try {
      const orderData = {
        menu: cart
          .map((item) => `${item.product.name} ${item.quantity}x`)
          .join(", "),
        total_price: total, // Pastikan total dikirim
        status: "On Progress",
        items: cart.map((item) => ({
          product_name: item.product.name,
          quantity: item.quantity,
          total_price: item.product.price * item.quantity,
        })), // Tambahkan detail item
      };

      console.log("Order Data:", orderData); // Debugging: Cek data yang dikirim

      const response = await axios.post(
        "http://localhost:8080/create-order",
        orderData
      );
      console.log("Checkout Response:", response.data); // Cek respons dari server
      alert("Checkout successful! Total: " + formatPrice(total));
      setCart([]); // Kosongkan keranjang setelah checkout
      setTotal(0); // Reset total
      onCheckout(); // Panggil callback untuk memberitahu dashboard
    } catch (error) {
      console.error("Error during checkout:", error);
      alert("Checkout failed. Please try again.");
    }
  };

  return (
    <div className="page-container">
      <h1>Cashier Page</h1>
      <div className="cashier-container">
        <div className="cashier-left">
          <h2>Menu</h2>
          <ul className="product-container">
            {products.map((product) => (
              <li key={product.id} className="productlist">
                <img
                  src={product.image_url}
                  alt={product.name}
                  className="cashier-img"
                />
                {product.name} <br /> {formatPrice(product.price)}
                {/* Menggunakan QuantityInput untuk menentukan jumlah produk */}
                <QuantityInput
                  onAddToCart={(quantity) => addToCart(product, quantity)}
                />
              </li>
            ))}
          </ul>
        </div>
        <div className="cashier-right">
          <h2>Cart</h2>
          <ul className="cart-container">
            {cart.map((item, index) => (
              <li key={index} className="cart-item">
                {item.product.name} - {formatPrice(item.product.price)} x
                {item.quantity}
                <br />
                <button
                  onClick={() => removeFromCart(item.product.id)}
                  className="remove-btn"
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
          <h3>Total: {formatPrice(total)}</h3>
          <button
            onClick={handleCheckout}
            disabled={cart.length === 0}
            className="button-place-order"
          >
            Place Order
          </button>
        </div>
      </div>
    </div>
  );
};

Cashier.propTypes = {
  onCheckout: PropTypes.func.isRequired, // Validate onCheckout as a required function
};

export default Cashier;
