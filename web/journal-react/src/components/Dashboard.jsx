import { useState, useEffect } from "react";
import { 
  BarChart, 
  PieChart, 
  LineChart, 
  TrendingUp, 
  TrendingDown, 
  Activity, 
  DollarSign,
  Tag 
} from "lucide-react";

const Dashboard = () => {
  const [stats, setStats] = useState(null);
  const [trades, setTrades] = useState([]);
  const [tags, setTags] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filterByTag, setFilterByTag] = useState(null);
  const [dateRange, setDateRange] = useState("all"); // all, week, month, year

  useEffect(() => {
    const fetchDashboardData = async () => {
      setLoading(true);
      try {
        // fetch statistics
        const statsResponse = await fetch("http://localhost:8080/api/statistics");
        if (!statsResponse.ok) {
          throw new Error("Failed to fetch statistics");
        }
        const statsData = await statsResponse.json();
        
        // fetch recent trades
        const tradesResponse = await fetch("http://localhost:8080/api/trades?limit=10");
        if (!tradesResponse.ok) {
          throw new Error("Failed to fetch trades");
        }
        const tradesData = await tradesResponse.json();
        
        // fetch tags
        const tagsResponse = await fetch("http://localhost:8080/api/tags");
        if (!tagsResponse.ok) {
          throw new Error("Failed to fetch tags");
        }
        const tagsData = await tagsResponse.json();
        
        setStats(statsData);
        setTrades(tradesData);
        setTags(tagsData);
      } catch (error) {
        console.error("Error fetching dashboard data:", error);
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };
    
    fetchDashboardData();
  }, []);

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto mt-10 p-6 bg-white rounded-lg shadow">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto mt-10 p-6 bg-white rounded-lg shadow">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          <p>Error loading dashboard: {error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto mt-10 pb-10">
      <h1 className="text-3xl font-bold mb-6">Trading Dashboard</h1>
      
      {/* Key Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {stats && (
          <>
            <StatCard 
              title="Win Rate" 
              value={`${(stats.win_rate * 100).toFixed(1)}%`} 
              icon={<Activity className="h-8 w-8 text-blue-500" />}
              color="blue"
            />
            <StatCard 
              title="Profit Factor" 
              value={stats.profit_factor.toFixed(2)} 
              icon={<DollarSign className="h-8 w-8 text-green-500" />}
              color="green"
            />
            <StatCard 
              title="Total Trades" 
              value={stats.total_trades} 
              icon={<BarChart className="h-8 w-8 text-purple-500" />}
              color="purple"
            />
            <StatCard 
              title="Current Streak" 
              value={stats.current_streak < 0 
                ? `${stats.current_streak}` // Already has minus sign if negative
                : `+${stats.current_streak}` // Add plus sign for winning streak
              }
              icon={stats.current_streak >= 0 ? 
                <TrendingUp className="h-8 w-8 text-green-500" /> : 
                <TrendingDown className="h-8 w-8 text-red-500" />
              }
              color={stats.current_streak >= 0 ? "green" : "red"}
            />
          </>
        )}
      </div>
      
      {/* Second Row - Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Win/Loss Chart */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 flex items-center">
            <PieChart className="h-5 w-5 mr-2" />
            Win/Loss Ratio
          </h2>
          {stats && (
            <div className="flex items-center justify-center h-64">
              {/* Placeholder for Win/Loss Pie Chart */}
              <div className="flex w-full max-w-xs">
                <div className="relative h-48 w-48 mx-auto">
                  <div className="h-full w-full rounded-full overflow-hidden">
                    <div className="h-full bg-green-500" 
                         style={{width: `${(stats.winning_trades / stats.total_trades) * 100}%`}}></div>
                  </div>
                  <div className="absolute inset-0 flex flex-col items-center justify-center">
                    <span className="text-2xl font-bold">
                      {stats.winning_trades} / {stats.losing_trades}
                    </span>
                    <span className="text-gray-500">Win / Loss</span>
                  </div>
                </div>
                
                <div className="flex flex-col justify-center p-4">
                  <div className="flex items-center mb-2">
                    <div className="w-4 h-4 bg-green-500 mr-2"></div>
                    <span>Wins: {stats.winning_trades}</span>
                  </div>
                  <div className="flex items-center">
                    <div className="w-4 h-4 bg-red-500 mr-2"></div>
                    <span>Losses: {stats.losing_trades}</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
        
        {/* Profit/Loss Chart */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 flex items-center">
            <LineChart className="h-5 w-5 mr-2" />
            Profit/Loss Overview
          </h2>
          <div className="flex items-center justify-center h-64">
            {/* Placeholder for Profit/Loss Chart - In a real implementation, use a proper charting library */}
            <div className="w-full h-48 bg-gray-100 flex items-center justify-center rounded-lg">
              <span className="text-gray-500">Equity curve would be rendered here</span>
            </div>
          </div>
        </div>
      </div>
      
      {/* Third Row - Tags and Recent Trades */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Tags Section */}
        <div className="bg-white rounded-lg shadow p-6 lg:col-span-1">
          <h2 className="text-xl font-semibold mb-4 flex items-center">
            <Tag className="h-5 w-5 mr-2" />
            Trade Tags
          </h2>
          <div className="flex flex-wrap gap-2 mb-4">
            {tags && tags.map(tag => (
              <button
                key={tag.id}
                onClick={() => setFilterByTag(filterByTag === tag.id ? null : tag.id)}
                className={`px-3 py-1 rounded-full text-sm flex items-center ${
                  filterByTag === tag.id ? "bg-gray-700 text-white" : "bg-gray-200 text-gray-800"
                }`}
                style={{
                  backgroundColor: filterByTag === tag.id ? tag.color : "#f3f4f6",
                  color: filterByTag === tag.id ? "#ffffff" : "#1f2937",
                }}
              >
                {tag.name}
              </button>
            ))}
          </div>
          {tags && (
            <div className="mt-4">
              <h3 className="font-semibold mb-2">Tag Performance</h3>
              <div className="space-y-2">
                {/* We would need additional data from the backend to show per-tag performance */}
                <p className="text-sm text-gray-500">
                  Tag performance statistics will be displayed here.
                </p>
              </div>
            </div>
          )}
        </div>
        
        {/* Recent Trades */}
        <div className="bg-white rounded-lg shadow p-6 lg:col-span-2">
          <h2 className="text-xl font-semibold mb-4">Recent Trades</h2>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Date
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Ticker
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Direction
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    P/L
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {trades.length > 0 ? (
                  trades.map(trade => (
                    <tr key={trade.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(trade.trade_date).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {trade.ticker}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <span className={`px-2 py-1 rounded ${
                          trade.direction === "LONG" ? "bg-green-100 text-green-800" : 
                                                      "bg-red-100 text-red-800"
                        }`}>
                          {trade.direction}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        <span className={
                          trade.exit_price > trade.entry_price && trade.direction === "LONG" || 
                          trade.exit_price < trade.entry_price && trade.direction === "SHORT" 
                            ? "text-green-600" : "text-red-600"
                        }>
                          {calculateProfit(trade)}
                        </span>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan="4" className="px-6 py-4 text-center text-sm text-gray-500">
                      No trades found
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};

// Statistic Card Component
const StatCard = ({ title, value, icon, color }) => {
  const colorClass = {
    blue: "border-blue-500",
    green: "border-green-500",
    purple: "border-purple-500",
    red: "border-red-500"
  }[color];

  return (
    <div className={`bg-white rounded-lg shadow p-6 border-l-4 ${colorClass}`}>
      <div className="flex justify-between items-center">
        <div>
          <p className="text-gray-500 text-sm font-medium">{title}</p>
          <p className="text-2xl font-bold mt-1">{value}</p>
        </div>
        <div>
          {icon}
        </div>
      </div>
    </div>
  );
};

// Helper function to calculate profit/loss
const calculateProfit = (trade) => {
  const entry = parseFloat(trade.entry_price);
  const exit = parseFloat(trade.exit_price);
  const qty = parseFloat(trade.quantity);
  const comm = trade.commissions ? parseFloat(trade.commissions) : 0;

  let profitLoss;
  if (trade.direction === "LONG") {
    profitLoss = (exit - entry) * qty;
  } else {
    profitLoss = (entry - exit) * qty;
  }

  return (profitLoss - comm).toFixed(2);
};

export default Dashboard; 