import React from "react";
import "./table.css";

export default function StyledTable({ children }) {
  return <div className="table-container"><table>{children}</table></div>;
}
