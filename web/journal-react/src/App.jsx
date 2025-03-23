import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import AddTrade from "./components/AddTrades";

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={
          <Layout>
            <div className="p-4">Home Page</div>
          </Layout>
        } />
        <Route path="/trades" element={
          <Layout>
            <div className="p-4">Trades Page</div>
          </Layout>
        } />
        <Route path="/add_trade" element={
          <Layout>
            <AddTrade />
          </Layout>
        } />
        <Route path="/analytics" element={
          <Layout>
            <div className="p-4">Analytics Page</div>
          </Layout>
        } />
        <Route path="*" element={
          <Layout>
            <div className="p-4 text-red-500">404 - Page not found</div>
          </Layout>
        } />
      </Routes>
    </Router>
  );
}

export default App;