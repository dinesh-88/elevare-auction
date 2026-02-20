import { Navigate, Route, Routes } from "react-router-dom";
import AdminLayout from "./layouts/AdminLayout";
import PublicLayout from "./layouts/PublicLayout";
import Dashboard from "./pages/admin/Dashboard";
import Auctions from "./pages/admin/Auctions";
import CreateAuction from "./pages/admin/CreateAuction";
import Bidders from "./pages/admin/Bidders";
import AuctionDetail from "./pages/public/AuctionDetail";
import Home from "./pages/public/Home";
import Login from "./pages/public/Login";
import Register from "./pages/public/Register";
import TenantSignup from "./pages/public/TenantSignup";
import TenantRequests from "./pages/platform/TenantRequests";

export default function App() {
  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route path="/" element={<Home />} />
        <Route path="/auction/:id" element={<AuctionDetail />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/signup" element={<TenantSignup />} />
      </Route>

      <Route path="/admin" element={<AdminLayout />}>
        <Route index element={<Dashboard />} />
        <Route path="auctions" element={<Auctions />} />
        <Route path="create-auction" element={<CreateAuction />} />
        <Route path="bidders" element={<Bidders />} />
      </Route>

      <Route path="/platform/review" element={<TenantRequests />} />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
