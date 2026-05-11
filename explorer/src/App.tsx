import { Routes, Route } from 'react-router-dom';
import Layout from './components/layout/Layout';
import HomePage from './pages/HomePage';
import BlocksPage from './pages/BlocksPage';
import BlockDetailPage from './pages/BlockDetailPage';
import TxDetailPage from './pages/TxDetailPage';
import AddressPage from './pages/AddressPage';
import BlueCoinListPage from './pages/BlueCoinListPage';
import BlueCoinDetailPage from './pages/BlueCoinDetailPage';
import ValidatorsPage from './pages/ValidatorsPage';
import WalletPage from './pages/WalletPage';
import NotFoundPage from './pages/NotFoundPage';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="blocks" element={<BlocksPage />} />
        <Route path="block/:height" element={<BlockDetailPage />} />
        <Route path="block/hash/:hash" element={<BlockDetailPage />} />
        <Route path="tx/:hash" element={<TxDetailPage />} />
        <Route path="address/:address" element={<AddressPage />} />
        <Route path="bluecoins" element={<BlueCoinListPage />} />
        <Route path="bluecoin/:tokenId" element={<BlueCoinDetailPage />} />
        <Route path="validators" element={<ValidatorsPage />} />
        <Route path="wallet" element={<WalletPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
