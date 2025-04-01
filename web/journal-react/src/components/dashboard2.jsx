import { useState, useEffect } from "react";
import { 
  PieChart, 
  Pie, 
  Cell, 
  ResponsiveContainer, 
  Tooltip, 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Legend, 
  BarChart, 
  Bar,
  Sector
} from 'recharts';
import { 
  BarChart2, 
  TrendingUp, 
  TrendingDown, 
  Activity, 
  DollarSign,
  Calendar,
  Tag,
  Filter
} from "lucide-react";

const TICK_VALUES = {
  GC: { tickValue: 10, tickSize: 0.1 }, // Gold futures
  ES: { tickValue: 12.50, tickSize: 0.25 } // E-mini S&P 500 futures
};

const Dashboard = () => {
  const [stats, setStats] = useState(null);
  const [trades, setTrades] = useState([]);
  const [tags, setTags] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filterByTag, setFilterByTag] = useState(null);
  const [dateRange, setDateRange] = useState("all"); // all, week, month, year
  const [equityData, setEquityData] = useState([]);
  const [activeIndex, setActiveIndex] = useState(0);
  const [startingBalance, setStartingBalance] = useState(0);
  
  // Statistics cards data
  const getStatCards = () => {
    if (!stats) return [];
    
    return [
      {
        title: "Total Trades",
        value: stats.total_trades,
        icon: <Activity size={24} />,
        color: "bg-blue-100 text-blue-800"
      },
      {
        title: "Win Rate",
        value: `${(stats.win_rate * 100).toFixed(2)}%`,
        icon: <TrendingUp size={24} />,
        color: "bg-green-100 text-green-800"
      },
      {
        title: "Avg. Profit",
        value: `$${stats.average_profit_loss.toFixed(2)}`,
        icon: <DollarSign size={24} />,
        color: stats.average_profit_loss >= 0 ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"
      },
      {
        title: "Profit Factor",
        value: stats.profit_factor.toFixed(2),
        icon: <BarChart2 size={24} />,
        color: "bg-purple-100 text-purple-800"
      },
      {
        title: "Current Streak",
        value: stats.current_streak,
        icon: stats.current_streak >= 0 ? <TrendingUp size={24} /> : <TrendingDown size={24} />,
        color: stats.current_streak >= 0 ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"
      },
      {
        title: "Max Drawdown",
        value: `${stats.max_drawdown.toFixed(2)}%`,
        icon: <TrendingDown size={24} />,
        color: "bg-orange-100 text-orange-800"
      }
    ];
  };

  // Colors for the win/loss pie chart
  const COLORS = ['#16a34a', '#dc2626', '#94a3b8'];

  // Date range filters
  const dateRangeOptions = [
    { value: "all", label: "All Time" },
    { value: "week", label: "This Week" },
    { value: "month", label: "This Month" },
    { value: "year", label: "This Year" }
  ];

  // Add these to dependencies for useEffect to trigger refetching when filters change
  useEffect(() => {
    fetchDashboardData();
  }, [filterByTag, dateRange]); // Now depends on filters

    const fetchDashboardData = async () => {
      setLoading(true);
      try {
      // Build URL with filters
      let statsUrl = "http://localhost:8080/api/statistics";
      let tradesUrl = "http://localhost:8080/api/trades?limit=10";
      
      // Add date range filter parameters
      if (dateRange !== "all") {
        const dateParams = getDateRangeParams(dateRange);
        statsUrl += `?${dateParams}`;
        tradesUrl += `&${dateParams}`;
      }
      
      // Add tag filter if selected
      if (filterByTag) {
        const tagParam = `tag_id=${filterByTag}`;
        statsUrl += statsUrl.includes('?') ? `&${tagParam}` : `?${tagParam}`;
        tradesUrl += `&${tagParam}`;
      }
      
      // Fetch data with filters applied
      const statsResponse = await fetch(statsUrl);
        if (!statsResponse.ok) {
          throw new Error("Failed to fetch statistics");
        }
        const statsData = await statsResponse.json();
        
      const tradesResponse = await fetch(tradesUrl);
        if (!tradesResponse.ok) {
          throw new Error("Failed to fetch trades");
        }
        const tradesData = await tradesResponse.json();
        
      // Fetch tags (no need to filter tags)
        const tagsResponse = await fetch("http://localhost:8080/api/tags");
        if (!tagsResponse.ok) {
          throw new Error("Failed to fetch tags");
        }
        const tagsData = await tagsResponse.json();
        
        setStats(statsData);
        setTrades(tradesData);
        setTags(tagsData);

      // Generate equity curve data based on filtered trades
      generateEquityCurveData(tradesData);
      } catch (error) {
        console.error("Error fetching dashboard data:", error);
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };
    
  // Helper function to create date parameters based on the selected range
  const getDateRangeParams = (range) => {
    const today = new Date();
    let startDate = new Date(today);
    
    switch(range) {
      case "week":
        startDate.setDate(today.getDate() - 7);
        break;
      case "month":
        startDate.setMonth(today.getMonth() - 1);
        break;
      case "year":
        startDate.setFullYear(today.getFullYear() - 1);
        break;
      default:
        return "";
    }
    
    const formatDate = (date) => {
      return date.toISOString().split('T')[0]; // YYYY-MM-DD format
    };
    
    return `start_date=${formatDate(startDate)}&end_date=${formatDate(today)}`;
  };

  // Handle tag filtering - updated to trigger refetch
  const handleTagSelect = (tagId) => {
    setFilterByTag(filterByTag === tagId ? null : tagId);
    // No need to call fetchDashboardData() here since useEffect will trigger it
  };

  // Handle date range changes - updated to trigger refetch
  const handleDateRangeChange = (range) => {
    setDateRange(range);
    // No need to call fetchDashboardData() here since useEffect will trigger it
  };

  // Generate equity curve data from trades
  const generateEquityCurveData = (trades) => {
    if (!trades || trades.length === 0) return;

    let balance = startingBalance; // Use the dynamic starting balance
    const equityPoints = trades
      .sort((a, b) => new Date(a.trade_date) - new Date(b.trade_date))
      .map(trade => {
        // Get the tick value and size based on the ticker
        const { tickValue, tickSize } = TICK_VALUES[trade.ticker] || { tickValue: 0, tickSize: 1 }; // Default to 0 if ticker not found

        // Calculate profit/loss for this trade
        let profit = 0;
        if (trade.exit_price && trade.entry_price) {
          // Calculate the number of ticks moved
          const ticksMoved = trade.direction === "LONG" 
            ? (trade.exit_price - trade.entry_price) / tickSize
            : (trade.entry_price - trade.exit_price) / tickSize;

          // Calculate profit based on the number of ticks moved
          profit = ticksMoved * tickValue * trade.quantity; // Adjusted profit calculation
        }
        
        if (trade.commissions) {
          profit -= trade.commissions; // Subtract commissions
        }
        
        balance += profit; // Update the balance with the profit/loss
        
        return {
          date: new Date(trade.trade_date).toLocaleDateString(),
          balance: balance.toFixed(2) // Store the current balance
        };
      });

    // If there are no equity points, set a default point to avoid rendering issues
    if (equityPoints.length === 0) {
      equityPoints.push({ date: new Date().toLocaleDateString(), balance: startingBalance.toFixed(2) });
    }

    setEquityData(equityPoints); // Set the equity data for rendering
  };

  // Custom active shape for pie chart
  const renderActiveShape = (props) => {
    const RADIAN = Math.PI / 180;
    const { cx, cy, midAngle, innerRadius, outerRadius, startAngle, endAngle, fill, payload, percent, value } = props;
    const sin = Math.sin(-RADIAN * midAngle);
    const cos = Math.cos(-RADIAN * midAngle);
    const sx = cx + (outerRadius + 10) * cos;
    const sy = cy + (outerRadius + 10) * sin;
    const mx = cx + (outerRadius + 30) * cos;
    const my = cy + (outerRadius + 30) * sin;
    const ex = mx + (cos >= 0 ? 1 : -1) * 22;
    const ey = my;
    const textAnchor = cos >= 0 ? 'start' : 'end';

    return (
      <g>
        <text x={cx} y={cy} dy={8} textAnchor="middle" fill={fill} className="text-lg font-semibold">
          {payload.name}
        </text>
        <Sector
          cx={cx}
          cy={cy}
          innerRadius={innerRadius}
          outerRadius={outerRadius + 6}
          startAngle={startAngle}
          endAngle={endAngle}
          fill={fill}
        />
        <Sector
          cx={cx}
          cy={cy}
          startAngle={startAngle}
          endAngle={endAngle}
          innerRadius={outerRadius + 6}
          outerRadius={outerRadius + 10}
          fill={fill}
        />
        <path d={`M${sx},${sy}L${mx},${my}L${ex},${ey}`} stroke={fill} fill="none" />
        <circle cx={ex} cy={ey} r={2} fill={fill} stroke="none" />
        <text x={ex + (cos >= 0 ? 1 : -1) * 12} y={ey} textAnchor={textAnchor} fill="#333">
          {`${payload.name}: ${value}`}
        </text>
        <text x={ex + (cos >= 0 ? 1 : -1) * 12} y={ey} dy={18} textAnchor={textAnchor} fill="#999">
          {`(${(percent * 100).toFixed(2)}%)`}
        </text>
      </g>
    );
  };

  const onPieEnter = (_, index) => {
    setActiveIndex(index);
  };

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-black"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto p-6">
        <div className="bg-red-100 text-red-800 p-4 rounded-lg">
          <h3 className="font-bold">Error Loading Dashboard</h3>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto mt-6 p-6">
      <h1 className="text-3xl font-bold mb-8">Trading Dashboard</h1>
      
      {/* Filters */}
      <div className="mb-8 bg-white rounded-lg shadow p-4">
        <div className="flex flex-wrap items-center gap-4">
          <div className="flex items-center">
            <Filter size={18} className="mr-2" />
            <span className="font-medium">Filters:</span>
          </div>
          
          {/* Date Range Filter */}
          <div className="flex items-center">
            <Calendar size={16} className="mr-2" />
            <select 
              value={dateRange} 
              onChange={(e) => handleDateRangeChange(e.target.value)}
              className="p-2 border border-gray-300 rounded"
            >
              {dateRangeOptions.map(option => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
          </div>
          
          {/* Tag Filters */}
          <div className="flex items-center flex-wrap gap-2">
            <Tag size={16} className="mr-1" />
            {tags && tags.length > 0 ? (
              tags.map(tag => (
                <button
                  key={tag.id}
                  onClick={() => handleTagSelect(tag.id)}
                  className={`px-3 py-1 rounded-full text-sm flex items-center ${
                    filterByTag === tag.id ? "text-white" : "text-gray-800"
                  }`}
                  style={{
                    backgroundColor: filterByTag === tag.id ? tag.color : "#f3f4f6",
                    color: filterByTag === tag.id ? "#ffffff" : "#1f2937",
                  }}
                >
                  {tag.name}
                </button>
              ))
            ) : (
              <span className="text-gray-500">No tags available</span>
            )}
          </div>
        </div>
      </div>
      
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        {getStatCards().map((card, index) => (
          <div key={index} className={`${card.color} p-4 rounded-lg shadow flex items-center`}>
            <div className="p-3 rounded-full mr-4 bg-white">
              {card.icon}
            </div>
            <div>
              <h3 className="text-sm font-medium">{card.title}</h3>
              <p className="text-2xl font-bold">{card.value}</p>
            </div>
          </div>
        ))}
      </div>
      
      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Win/Loss Pie Chart */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-lg font-bold mb-4">Win/Loss/Break Even Distribution</h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                activeIndex={activeIndex}
                activeShape={renderActiveShape}
                data={[
                  { name: 'Wins', value: stats.winning_trades },
                  { name: 'Losses', value: stats.losing_trades },
                  { name: 'Break Even', value: stats.break_even_trades }
                ]}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
                onMouseEnter={onPieEnter}
              >
                {[stats.winning_trades, stats.losing_trades, stats.break_even_trades].map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
            </PieChart>
          </ResponsiveContainer>
        </div>
        
        {/* Equity Curve */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-lg font-bold mb-4">Equity Curve</h2>
          {equityData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart
                data={equityData}
                margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                style={{ minHeight: '300px' }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="date" 
                  tick={{ fontSize: 12 }}
                  tickFormatter={(value) => value}
                />
                <YAxis />
                <Tooltip formatter={(value) => [`$${value}`, 'Balance']} />
                <Line 
                  type="monotone" 
                  dataKey="balance" 
                  stroke="#2563eb" 
                  activeDot={{ r: 8 }} 
                  strokeWidth={2}
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[300px] flex items-center justify-center text-gray-500">
              No equity data available
            </div>
          )}
        </div>
      </div>
      
      {/* Second Row of Charts */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        {/* Average Win vs Loss */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-lg font-bold mb-4">Average Win vs Loss</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart
              data={[
                { name: 'Avg Win', value: stats.average_winner, fill: '#16a34a' },
                { name: 'Avg Loss', value: Math.abs(stats.average_loser), fill: '#dc2626' },
              ]}
              margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip formatter={(value) => `$${value.toFixed(2)}`} />
              <Bar dataKey="value" name="Amount">
                {/* Use cell to give each bar its own color */}
                <Cell fill="#16a34a" /> {/* Green for win */}
                <Cell fill="#dc2626" /> {/* Red for loss */}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
        
        {/* Largest Win vs Loss */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-lg font-bold mb-4">Largest Win vs Loss</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart
              data={[
                { name: 'Largest Win', value: stats.largest_winner, fill: '#16a34a' },
                { name: 'Largest Loss', value: Math.abs(stats.largest_loser), fill: '#dc2626' },
              ]}
              margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip formatter={(value) => `$${value.toFixed(2)}`} />
              <Bar dataKey="value" name="Amount">
                {/* Use cell to give each bar its own color */}
                <Cell fill="#16a34a" /> {/* Green for win */}
                <Cell fill="#dc2626" /> {/* Red for loss */}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
      
      {/* Recent Trades */}
      <div className="bg-white p-6 rounded-lg shadow mb-8">
        <h2 className="text-lg font-bold mb-4">Recent Trades</h2>
        {trades.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-gray-100">
                  <th className="px-4 py-2 text-left">Date</th>
                  <th className="px-4 py-2 text-left">Ticker</th>
                  <th className="px-4 py-2 text-left">Direction</th>
                  <th className="px-4 py-2 text-right">Entry</th>
                  <th className="px-4 py-2 text-right">Exit</th>
                  <th className="px-4 py-2 text-right">Quantity</th>
                  <th className="px-4 py-2 text-right">P/L</th>
                </tr>
              </thead>
              <tbody>
                {trades.map((trade) => {
                  // Get the tick value and size based on the ticker
                  const { tickValue, tickSize } = TICK_VALUES[trade.ticker] || { tickValue: 0, tickSize: 1 }; // Default to 0 if ticker not found

                  // Calculate P/L for the trade
                  let profitLoss = 0;
                  if (trade.exit_price && trade.entry_price) {
                    const ticksMoved = trade.direction === "LONG" 
                      ? (trade.exit_price - trade.entry_price) / tickSize
                      : (trade.entry_price - trade.exit_price) / tickSize;

                    profitLoss = ticksMoved * tickValue * trade.quantity; // Adjusted profit calculation
                  }

                  const isProfitable = profitLoss >= 0;

                  return (
                    <tr key={trade.id} className="border-b hover:bg-gray-50">
                      <td className="px-4 py-2">
                        {new Date(trade.trade_date).toLocaleDateString()}
                      </td>
                      <td className="px-4 py-2 font-medium">{trade.ticker}</td>
                      <td className="px-4 py-2">
                        <span className={`px-2 py-1 rounded text-xs ${
                          trade.direction === "LONG" ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"
                        }`}>
                          {trade.direction}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-right">${trade.entry_price.toFixed(2)}</td>
                      <td className="px-4 py-2 text-right">${trade.exit_price ? trade.exit_price.toFixed(2) : '-'}</td>
                      <td className="px-4 py-2 text-right">{trade.quantity}</td>
                      <td className={`px-4 py-2 text-right font-medium ${
                        isProfitable ? "text-green-600" : "text-red-600"
                      }`}>
                        {isProfitable ? '+' : ''}${profitLoss.toFixed(2)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="text-gray-500 text-center py-8">No recent trades found</div>
        )}
      </div>
    </div>
  );
};

export default Dashboard;