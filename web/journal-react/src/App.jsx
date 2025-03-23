import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import AddTrade from "./components/AddTrades";

function App() {
  return (
    <Router>
      <Routes>
        <Route
          path="/"
          element={
            <Layout>
              <div>Home Page</div>
            </Layout>
          }
        />
        <Route
          path="/trades"
          element={
            <Layout>
              <div>Trades Page</div>
            </Layout>
          }
        />
        <Route path="/add_trade" element={<Layout><AddTrade /></Layout>} />
        <Route
          path="/analytics"
          element={
            <Layout>
              <div>Analytics Page</div>
            </Layout>
          }
        />
      </Routes>
    </Router>
  );
}

export default App;
