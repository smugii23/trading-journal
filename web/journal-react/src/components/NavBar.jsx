import { Link } from 'react-router-dom';
import { useState } from 'react';

function NavBar() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      {/* Import Manrope font */}
      <style jsx global>{`
        @import url('https://fonts.googleapis.com/css2?family=Manrope:wght@400;500;600;700&display=swap');
        body {
          font-family: 'Manrope', sans-serif;
        }
      `}</style>

      <nav className="bg-white p-4 shadow-md" style={{ fontFamily: 'Manrope, sans-serif' }}>
        <div className="max-w-4xl mx-auto flex justify-between items-center">
          {/* Logo */}
          <Link to="/" className="text-gray-800 text-lg font-semibold hover:text-blue-600">
            Trading Journal
          </Link>

          {/* Desktop Links */}
          <div className="hidden md:flex space-x-6">
            <Link to="/trades" className="text-gray-700 hover:text-blue-600 transition duration-300">
              Trades
            </Link>
            <Link to="/add_trade" className="text-gray-700 hover:text-blue-600 transition duration-300">
              Add Trade
            </Link>
            <Link to="/analytics" className="text-gray-700 hover:text-blue-600 transition duration-300">
              Analytics
            </Link>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="md:hidden text-gray-800 focus:outline-none"
            onClick={() => setIsOpen(!isOpen)}
          >
            ☰
          </button>
        </div>

        {/* Mobile Menu */}
        {isOpen && (
          <div className="md:hidden bg-white p-4 border-t border-gray-200">
            <Link to="/trades" className="block text-gray-700 py-2 hover:text-blue-600">
              Trades
            </Link>
            <Link to="/add_trade" className="block text-gray-700 py-2 hover:text-blue-600">
              Add Trade
            </Link>
            <Link to="/analytics" className="block text-gray-700 py-2 hover:text-blue-600">
              Analytics
            </Link>
          </div>
        )}
      </nav>
    </>
  );
}

export default NavBar;