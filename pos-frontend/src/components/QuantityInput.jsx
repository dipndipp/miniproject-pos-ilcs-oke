import { useState } from "react";
import PropTypes from "prop-types";

const QuantityInput = ({ onAddToCart }) => {
  const [quantity, setQuantity] = useState(1);

  const handleIncrease = () => {
    setQuantity((prevQuantity) => prevQuantity + 1);
  };

  const handleDecrease = () => {
    if (quantity > 1) {
      setQuantity((prevQuantity) => prevQuantity - 1);
    }
  };

  const handleChange = (e) => {
    const value = e.target.value;

    // Validasi hanya angka
    if (/^\d*$/.test(value)) {
      const parsedValue = parseInt(value, 10);

      // Pastikan value adalah angka valid atau minimal 1
      if (!isNaN(parsedValue) && parsedValue > 0) {
        setQuantity(parsedValue);
      } else {
        setQuantity(1); // Default ke 1 jika input tidak valid
      }
    }
  };

  return (
    <div className="quantity-input-container">
      <div className="quantity-input">
        <button onClick={handleDecrease} className="quantity-btn">
          -
        </button>
        <input
          type="text"
          value={quantity}
          onChange={handleChange}
          inputMode="numeric" // Membatasi input keyboard pada angka
          className="quantity-value"
        />
        <button onClick={handleIncrease} className="quantity-btn">
          +
        </button>
      </div>
      <button onClick={() => onAddToCart(quantity)} className="add-to-cart-btn">
        Add to Cart
      </button>
    </div>
  );
};

QuantityInput.propTypes = {
  onAddToCart: PropTypes.func.isRequired,
};

export default QuantityInput;
