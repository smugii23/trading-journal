import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';

function TradeView() {
  const { id } = useParams();
  const [trade, setTrade] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [modalOpen, setModalOpen] = useState(false);

  useEffect(() => {
    const fetchTrade = async () => {
      try {
        const response = await fetch(`/api/trades/${id}`);
        
        if (!response.ok) {
          throw new Error('Failed to fetch trade');
        }
        
        const data = await response.json();
        setTrade(data);
      } catch (error) {
        console.error('Error fetching trade:', error);
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };
    
    fetchTrade();
  }, [id]);

  if (loading) {
    return <div className="text-center py-10">Loading trade details...</div>;
  }
  
  if (error) {
    return <div className="text-center py-10 text-red-500">Error: {error}</div>;
  }
  
  if (!trade) {
    return <div className="text-center py-10">Trade not found</div>;
  }

  // Calculate profit/loss
  const profitLoss = trade.metrics ? trade.metrics.profit_loss : null;
  const isProfitable = profitLoss > 0;

  return (
    <div className="max-w-6xl mx-auto mt-10 p-6 bg-white rounded-lg shadow">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">{trade.ticker} - {trade.direction.toUpperCase()}</h2>
        <div className={`px-4 py-2 rounded-lg font-semibold ${isProfitable ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {profitLoss ? `${isProfitable ? '+' : ''}${profitLoss.toFixed(2)}` : 'N/A'}
        </div>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Trade Details */}
        <div className="md:col-span-2 space-y-4">
          <h3 className="text-xl font-semibold">Trade Details</h3>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-gray-600">Date</p>
              <p className="font-medium">{new Date(trade.trade_date).toLocaleDateString()}</p>
            </div>
            <div>
              <p className="text-gray-600">Direction</p>
              <p className="font-medium">{trade.direction.toUpperCase()}</p>
            </div>
            <div>
              <p className="text-gray-600">Entry Price</p>
              <p className="font-medium">{trade.entry_price}</p>
            </div>
            <div>
              <p className="text-gray-600">Exit Price</p>
              <p className="font-medium">{trade.exit_price || 'N/A'}</p>
            </div>
            <div>
              <p className="text-gray-600">Quantity</p>
              <p className="font-medium">{trade.quantity}</p>
            </div>
            <div>
              <p className="text-gray-600">Stop Loss</p>
              <p className="font-medium">{trade.stop_loss || 'N/A'}</p>
            </div>
          </div>
          
          {/* Trade Metrics */}
          {trade.metrics && (
            <div className="mt-6">
              <h3 className="text-xl font-semibold mb-2">Trade Metrics</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-gray-600">Profit/Loss</p>
                  <p className={`font-medium ${isProfitable ? 'text-green-600' : 'text-red-600'}`}>
                    {isProfitable ? '+' : ''}{trade.metrics.profit_loss.toFixed(2)}
                  </p>
                </div>
                <div>
                  <p className="text-gray-600">P/L %</p>
                  <p className={`font-medium ${isProfitable ? 'text-green-600' : 'text-red-600'}`}>
                    {isProfitable ? '+' : ''}{trade.metrics.profit_loss_percent.toFixed(2)}%
                  </p>
                </div>
                <div>
                  <p className="text-gray-600">Risk/Reward</p>
                  <p className="font-medium">{trade.metrics.risk_reward_ratio ? trade.metrics.risk_reward_ratio.toFixed(2) : 'N/A'}</p>
                </div>
                <div>
                  <p className="text-gray-600">R-Multiple</p>
                  <p className="font-medium">{trade.metrics.r_multiple ? trade.metrics.r_multiple.toFixed(2) : 'N/A'}</p>
                </div>
              </div>
            </div>
          )}
          
          {/* Notes */}
          {trade.notes && (
            <div className="mt-6">
              <h3 className="text-xl font-semibold mb-2">Notes</h3>
              <div className="p-4 bg-gray-50 rounded-lg whitespace-pre-wrap">
                {trade.notes}
              </div>
            </div>
          )}
        </div>
        
        {/* Screenshot */}
        <div>
          <h3 className="text-xl font-semibold mb-2">Screenshot</h3>
          {trade.screenshot_url ? (
            <div 
              className="cursor-pointer" 
              onClick={() => setModalOpen(true)}
            >
              <img 
                src={trade.screenshot_url} 
                alt="Trade Screenshot" 
                className="w-full rounded-lg border border-gray-200 hover:opacity-90 transition-opacity"
              />
              <p className="text-center text-sm text-gray-500 mt-1">Click to enlarge</p>
            </div>
          ) : (
            <div className="border border-dashed border-gray-300 rounded-lg flex items-center justify-center p-8">
              <p className="text-gray-500">No screenshot available</p>
            </div>
          )}
        </div>
      </div>
      
      {/* Screenshot Modal */}
      {modalOpen && trade.screenshot_url && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50"
          onClick={() => setModalOpen(false)}
        >
          <div className="max-w-4xl max-h-screen p-4">
            <img 
              src={trade.screenshot_url} 
              alt="Trade Screenshot" 
              className="max-w-full max-h-full object-contain"
            />
          </div>
        </div>
      )}
    </div>
  );
}

export default TradeView;