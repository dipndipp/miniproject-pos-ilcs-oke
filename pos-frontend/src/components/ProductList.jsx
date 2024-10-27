import { useEffect, useState } from "react";
import axios from "axios";
import AddProduct from "./AddProduct";
import { useNavigate } from "react-router-dom";
import useAuth from "../context/useAuth";

const ProductList = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const { user } = useAuth();

  // Pindahkan fetchProducts keluar dari useEffect agar bisa diakses di tempat lain
  const fetchProducts = async () => {
    setLoading(true);
    try {
      const response = await axios.get("http://localhost:8080/products");
      setProducts(response.data);
    } catch (error) {
      setError("Error fetching products: " + error.message);
      console.error("Error fetching products:", error);
    } finally {
      setLoading(false);
    }
  };

  const formatPrice = (price) => {
    return price.toLocaleString("id-ID", {
      style: "currency",
      currency: "IDR",
    });
  };
  useEffect(() => {
    // Cek apakah user masih belum diinisialisasi
    if (user === null) {
      console.log("User belum terinisialisasi dari localStorage.");
      return; // Tunggu sampai user terinisialisasi
    }

    // Log user yang ada setelah inisialisasi
    console.log("Current user in ProductList:", user);

    // Jika user masih null setelah inisialisasi, redirect ke halaman login
    if (!user) {
      alert("Please login first");
      navigate("/"); // Redirect ke halaman login jika user tidak terautentikasi
      return;
    }

    // Cek role user
    if (user.role !== "admin") {
      navigate("/"); // Redirect jika bukan admin
      return;
    }

    // Panggil fetchProducts
    fetchProducts();
  }, [navigate, user]);

  const handleDelete = async (id) => {
    if (window.confirm("Are you sure you want to delete this product?")) {
      try {
        await axios.delete(`http://localhost:8080/delete-product/${id}`);
        setProducts((prevProducts) =>
          prevProducts.filter((product) => product.id !== id)
        ); // Update state untuk menghapus produk
        alert("Product deleted successfully");
      } catch (error) {
        setError("Error deleting product: " + error.message);
        console.error("Error deleting product:", error);
      }
    }
  };

  const handleEdit = (id) => {
    navigate(`/edit-product/${id}`);
  };

  return (
    <div>
      <div className="list-container">
        {/* Pastikan fetchProducts diteruskan ke komponen AddProduct */}
        <AddProduct onProductAdded={fetchProducts} />

        <div className="listproduct-wrapper">
          <h1>Product List</h1>
          {error && <p style={{ color: "red" }}>{error}</p>}
          {loading ? (
            <p>Loading products...</p>
          ) : (
            <ul className="product-container">
              {products.map((product) => (
                <li key={product.id} className="productlist">
                  <img
                    src={product.image_url}
                    alt={product.name}
                    className="product-image"
                  />
                  <br />
                  {product.name} <br /> {formatPrice(product.price)}
                  <button
                    className="edit-button"
                    onClick={() => handleEdit(product.id)}
                  >
                    Edit
                  </button>
                  <button
                    className="delete-button"
                    onClick={() => handleDelete(product.id)}
                  >
                    Delete
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

export default ProductList;
