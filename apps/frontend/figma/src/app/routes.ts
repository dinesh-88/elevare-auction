import { createBrowserRouter } from 'react-router';
import { HomePage } from './pages/HomePage';
import { AuctionListingPage } from './pages/AuctionListingPage';
import { AuctionDetailPage } from './pages/AuctionDetailPage';

export const router = createBrowserRouter([
  {
    path: '/',
    Component: HomePage,
  },
  {
    path: '/auctions',
    Component: AuctionListingPage,
  },
  {
    path: '/auction/:id',
    Component: AuctionDetailPage,
  },
]);
