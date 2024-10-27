import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import axios from "axios";

const EditProduct = () => {
  const { id } = useParams(); // Dapatkan ID produk dari URL
  const navigate = useNavigate(); // Memanggil useNavigate untuk navigasi
  const [product, setProduct] = useState(null);
  const [name, setName] = useState("");
  const [price, setPrice] = useState("");
  const [imageFile, setImageFile] = useState(null); // Untuk menyimpan file gambar

  useEffect(() => {
    const fetchProduct = async () => {
      try {
        const response = await axios.get(`http://localhost:8080/product/${id}`);
        const productData = response.data;
        setProduct(productData);
        setName(productData.name);
        setPrice(productData.price);
      } catch (error) {
        console.error("Error fetching product:", error);
      }
    };

    fetchProduct();
  }, [id]);

  const handleSave = async () => {
    try {
      const formData = new FormData();
      formData.append("name", name);
      formData.append("price", price);
      if (imageFile) {
        formData.append("image", imageFile); // Menambahkan file image jika ada
      }

      // Request PUT
      await axios.put(`http://localhost:8080/update-product/${id}`, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      alert("Product updated successfully");
      navigate("/products");
    } catch (error) {
      console.error("Error updating product:", error);
      alert("Make sure all fields are filled in correctly");
    }
  };

  return product ? (
    <div className="editproduk container">
      <h1>Edit Product</h1>
      {/* Menampilkan gambar yang sudah ada */}
      {product.image_url && (
        <img
          src={product.image_url}
          alt={name}
          style={{ maxWidth: "200px", marginBottom: "10px" }}
        />
      )}
      <input
        type="file"
        name="image"
        accept="image/*"
        onChange={(e) => setImageFile(e.target.files[0])} // Menggunakan setImageFile
      />
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Product Name"
      />
      <input
        type="number"
        value={price}
        onChange={(e) => setPrice(e.target.value)}
        placeholder="Product Price"
      />
      <button onClick={handleSave}>Save</button>
    </div>
  ) : (
    <p>Loading product data...</p>
  );
};

export default EditProduct;
