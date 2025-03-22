import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import NavBar from './components/NavBar';
import Header from './components/Header';

function App() {
  return (
  <Router>
    <NavBar />
    <Routes>
      <Route path="/" element={<Header />}/>
      <Route path="/trades" element={<div>Trades Page</div>} />
      <Route path="/add_trade" element={<div>Add Trade Page</div>} />
      <Route path="/analytics" element={<div>Analytics Page</div>} />
    </Routes>
  </Router>
  )
}

export default App;
