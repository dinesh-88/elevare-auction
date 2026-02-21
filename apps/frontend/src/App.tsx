import { Navigate, Route, Routes } from "react-router-dom";
import AdminLayout from "./layouts/AdminLayout";
import PublicLayout from "./layouts/PublicLayout";
import Dashboard from "./pages/admin/Dashboard";
import Auctions from "./pages/admin/Auctions";
import CreateAuction from "./pages/admin/CreateAuction";
import Bidders from "./pages/admin/Bidders";
import Lots from "./pages/admin/Lots";
import FinancialRules from "./pages/admin/FinancialRules";
import CreateLot from "./pages/admin/CreateLot";
import EditLot from "./pages/admin/EditLot";
import AuctionDetail from "./pages/public/AuctionDetail";
import AuctionListing from "./pages/public/AuctionListing";
import Home from "./pages/public/Home";
import Login from "./pages/public/Login";
import LotDetail from "./pages/public/LotDetail";
import Register from "./pages/public/Register";
import TenantSignup from "./pages/public/TenantSignup";
import TenantRequests from "./pages/platform/TenantRequests";
import MyWinnings from "./pages/public/MyWinnings";

export default function App() {
  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route path="/" element={<Home />} />
        <Route path="/auctions" element={<AuctionListing />} />
        <Route path="/auction/:id" element={<AuctionDetail />} />
        <Route path="/lots/:lotId" element={<LotDetail />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/my-winnings" element={<MyWinnings />} />
        <Route path="/signup" element={<TenantSignup />} />
      </Route>

      <Route path="/admin" element={<AdminLayout />}>
        <Route index element={<Dashboard />} />
        <Route path="auctions" element={<Auctions />} />
        <Route path="auctions/:auctionId/lots" element={<Lots />} />
        <Route path="auctions/:auctionId/lots/create" element={<CreateLot />} />
        <Route path="auctions/:auctionId/lots/:lotId/edit" element={<EditLot />} />
        <Route path="auctions/:auctionId/financial-rules" element={<FinancialRules />} />
        <Route path="financial-rules" element={<FinancialRules />} />
        <Route path="create-auction" element={<CreateAuction />} />
        <Route path="bidders" element={<Bidders />} />
      </Route>

      <Route path="/platform/review" element={<TenantRequests />} />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
