import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App.jsx";
import "./input.css"; // IMPORTANT - Tailwind goes in here

createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
