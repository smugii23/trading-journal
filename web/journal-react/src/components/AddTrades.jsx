import { useState, useEffect } from "react";
import { Camera, AlertTriangle, CheckCircle, Tag, Plus, Check, Edit2, Trash } from "lucide-react";

const AddTrade = () => {
  const [trade, setTrade] = useState({
    ticker: "",
    direction: "long",
    entry_price: "",
    exit_price: "",
    quantity: "",
    trade_date: "",
    entry_time: "",
    exit_time: "",
    stop_loss: "",
    take_profit: "",
    commissions: "",
    highest_price: "",
    lowest_price: "",
    notes: "",
    screenshot: null,
  });

  const [previewUrl, setPreviewUrl] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [captureMode, setCaptureMode] = useState(false);
  const [notification, setNotification] = useState({ show: false, type: "", message: "" });
  const [profit, setProfit] = useState(null);
  const [selectedTags, setSelectedTags] = useState([]);

  // Calculate profit/loss when relevant fields change
  const calculateProfit = () => {
    if (trade.entry_price && trade.exit_price && trade.quantity) {
      const entry = parseFloat(trade.entry_price);
      const exit = parseFloat(trade.exit_price);
      const qty = parseFloat(trade.quantity);
      
      // Make sure to handle empty strings and invalid inputs
      const commStr = trade.commissions ? trade.commissions.toString().trim() : "";
      const comm = commStr === "" ? 0 : parseFloat(commStr) || 0;
      
      // For debugging
      console.log("Entry:", entry, "Exit:", exit, "Qty:", qty, "Comm:", comm);
      
      let profitLoss;
      if (trade.direction === "long") {
        profitLoss = (exit - entry) * qty;
      } else { 
        profitLoss = (entry - exit) * qty;
      }
      
      const finalProfit = profitLoss - comm;
      console.log("Profit/Loss:", profitLoss, "Final after comm:", finalProfit);
      return finalProfit;
    }
    return null;
  };
  
  const handleChange = (e) => {
    const { name, value } = e.target;
    setTrade(prevTrade => {
      const newTrade = { ...prevTrade, [name]: value };
      
      // Immediately recalculate profit on state update
      if (['entry_price', 'exit_price', 'quantity', 'direction', 'commissions'].includes(name)) {
        // Use the updated trade object for calculation
        const entry = parseFloat(newTrade.entry_price) || 0;
        const exit = parseFloat(newTrade.exit_price) || 0;
        const qty = parseFloat(newTrade.quantity) || 0;
        const commStr = newTrade.commissions ? newTrade.commissions.toString().trim() : "";
        const comm = commStr === "" ? 0 : parseFloat(commStr) || 0;
        
        let newProfit = null;
        if (newTrade.entry_price && newTrade.exit_price && newTrade.quantity) {
          let profitLoss;
          if (newTrade.direction === "long") {
            profitLoss = (exit - entry) * qty;
          } else { 
            profitLoss = (entry - exit) * qty;
          }
          newProfit = profitLoss - comm;
        }
        
        // Update profit state with the new calculation
        setTimeout(() => setProfit(newProfit), 0);
      }
      
      return newTrade;
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsUploading(true);
    
    try {
      // Create a FormData object for file upload
      const formData = new FormData();
      
      // Add all trade data to formData
      Object.keys(trade).forEach(key => {
        if (key === 'screenshot' && trade[key]) {
          formData.append(key, trade[key]);
        } else {
          formData.append(key, trade[key] || '');
        }
      });
      
      // Add selected tags to formData
      formData.append('tags', JSON.stringify(selectedTags));
      
      const response = await fetch('/api/trades', {
        method: 'POST',
        body: formData
      });
      
      if (!response.ok) {
        throw new Error('Failed to submit trade');
      }
      
      const result = await response.json();
      
      // Show success notification
      setNotification({
        show: true,
        type: "success",
        message: "Trade added successfully!"
      });
      
      // Reset form
      resetForm();
      
      // Hide notification after 3 seconds
      setTimeout(() => setNotification({ show: false, type: "", message: "" }), 3000);
      
    } catch (error) {
      console.error('Error submitting trade:', error);
      
      // Show error notification
      setNotification({
        show: true,
        type: "error",
        message: "Error submitting trade. Please try again."
      });
      
      // Hide notification after 3 seconds
      setTimeout(() => setNotification({ show: false, type: "", message: "" }), 3000);
    } finally {
      setIsUploading(false);
    }
  };

  const resetForm = () => {
    setTrade({
      ticker: "",
      direction: "long",
      entry_price: "",
      exit_price: "",
      quantity: "",
      trade_date: "",
      entry_time: "",
      exit_time: "",
      stop_loss: "",
      take_profit: "",
      commissions: "",
      highest_price: "",
      lowest_price: "",
      notes: "",
      screenshot: null,
    });
    setPreviewUrl(null);
    setProfit(null);
    setSelectedTags([]);
  };

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      if (file.size > 5 * 1024 * 1024) {
        setNotification({
          show: true,
          type: "error",
          message: "File size exceeds 5MB limit"
        });
        setTimeout(() => setNotification({ show: false, type: "", message: "" }), 3000);
        return;
      }
      
      setTrade({ ...trade, screenshot: file });
      
      // Create a preview URL
      const fileReader = new FileReader();
      fileReader.onload = () => {
        setPreviewUrl(fileReader.result);
      };
      fileReader.readAsDataURL(file);
    }
  };

  // Handle tag selection changes
  const handleTagsChange = (newSelectedTags) => {
    setSelectedTags(newSelectedTags);
  };

  return (
    <div className="max-w-5xl mx-auto mt-10 p-6 bg-white rounded-lg shadow">
      <h2 className="text-2xl font-bold mb-4">Add a New Trade</h2>
      
      {/* Notification */}
      {notification.show && (
        <div className={`mb-4 p-3 rounded flex items-center ${
          notification.type === "success" ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"
        }`}>
          {notification.type === "success" ? (
            <CheckCircle className="mr-2" size={20} />
          ) : (
            <AlertTriangle className="mr-2" size={20} />
          )}
          {notification.message}
        </div>
      )}
      
      <form onSubmit={handleSubmit}>
        {/* Ticker */}
        <div className="mb-4">
          <label className="block text-gray-700">Ticker</label>
          <input
            type="text"
            name="ticker"
            value={trade.ticker}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded mt-1"
            placeholder="Enter symbol (e.g., ES, NQ)"
            required
          />
        </div>

        {/* Direction */}
        <div className="mb-4">
          <label className="block text-gray-700">Direction</label>
          <select
            name="direction"
            value={trade.direction}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded mt-1"
            required
          >
            <option value="long">Long</option>
            <option value="short">Short</option>
          </select>
        </div>

        {/* Entry & Exit Price */}
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label className="block text-gray-700">Entry Price</label>
            <input
              type="number"
              name="entry_price"
              value={trade.entry_price}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              required
              step="any"
            />
          </div>
          <div>
            <label className="block text-gray-700">Exit Price</label>
            <input
              type="number"
              name="exit_price"
              value={trade.exit_price}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
        </div>

        {/* Quantity */}
        <div className="mb-4">
          <label className="block text-gray-700">Quantity</label>
          <input
            type="number"
            name="quantity"
            value={trade.quantity}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded mt-1"
            required
            min="1"
            step="any"
          />
        </div>

        {/* Trade Date, Entry Time, Exit Time */}
        <div className="grid grid-cols-3 gap-4 mb-4">
          <div>
            <label className="block text-gray-700">Trade Date</label>
            <input
              type="date"
              name="trade_date"
              value={trade.trade_date}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              required
            />
          </div>
          <div>
            <label className="block text-gray-700">Entry Time</label>
            <input
              type="time"
              name="entry_time"
              value={trade.entry_time}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              required
            />
          </div>
          <div>
            <label className="block text-gray-700">Exit Time</label>
            <input
              type="time"
              name="exit_time"
              value={trade.exit_time}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
            />
          </div>
        </div>

        {/* Stop Loss & Take Profit */}
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label className="block text-gray-700">Stop Loss</label>
            <input
              type="number"
              name="stop_loss"
              value={trade.stop_loss}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
          <div>
            <label className="block text-gray-700">Take Profit</label>
            <input
              type="number"
              name="take_profit"
              value={trade.take_profit}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
        </div>

        {/* Commissions */}
        <div className="mb-4">
          <label className="block text-gray-700">Commissions</label>
          <input
            type="number"
            name="commissions"
            value={trade.commissions}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded mt-1"
            step="any"
          />
        </div>

        {/* Highest & Lowest Price */}
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label className="block text-gray-700">Highest Price</label>
            <input
              type="number"
              name="highest_price"
              value={trade.highest_price}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
          <div>
            <label className="block text-gray-700">Lowest Price</label>
            <input
              type="number"
              name="lowest_price"
              value={trade.lowest_price}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
        </div>

        {/* Tag Manager */}
        <TagManager onTagsChange={handleTagsChange} selectedTags={selectedTags} />

        {/* Profit/Loss Calculator */}
        {profit !== null && (
          <div className="mb-4 p-3 rounded border border-gray-300">
            <h3 className="font-bold">Estimated P&L:</h3>
            <p className={profit >= 0 ? "text-green-600 font-bold" : "text-red-600 font-bold"}>
              {profit >= 0 ? "+" : ""}{profit.toFixed(2)}
            </p>
          </div>
        )}

        {/* Notes */}
        <div className="mb-4">
          <label className="block text-gray-700">Notes</label>
          <textarea
            name="notes"
            value={trade.notes}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded mt-1"
            placeholder="Optional notes about this trade"
            rows="4"
          ></textarea>
        </div>

        {/* Screenshot Upload */}
        <div className="mb-4">
          <label className="block text-gray-700">Trade Screenshot</label>
          <div className="mt-1 flex items-center space-x-4">
            <div className="flex-1">
              <input
                type="file"
                accept="image/*"
                onChange={handleFileChange}
                className="w-full p-2 border border-gray-300 rounded"
              />
            </div>
          </div>
          
          {/* Preview the image if available */}
          {previewUrl && (
            <div className="mt-2">
              <div className="relative group">
                <img 
                  src={previewUrl} 
                  alt="Screenshot preview" 
                  className="max-h-64 object-contain rounded border border-gray-300" 
                />
                
                {/* Remove image button */}
                <button
                  type="button"
                  className="absolute top-2 right-2 bg-red-500 text-white p-1 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                  onClick={() => {
                    setTrade({...trade, screenshot: null});
                    setPreviewUrl(null);
                  }}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="grid grid-cols-2 gap-4">
          <button 
            type="button" 
            onClick={resetForm}
            className="w-full bg-gray-200 text-gray-800 p-2 rounded mt-4 hover:bg-gray-300 transition-colors"
          >
            Reset Form
          </button>
          
          <button 
            type="submit" 
            className="w-full bg-black text-white p-2 rounded mt-4 hover:bg-gray-800 transition-colors disabled:bg-gray-400"
            disabled={isUploading}
          >
            {isUploading ? 'Submitting...' : 'Submit Trade'}
          </button>
        </div>
      </form>
    </div>
  );
};

// Separate TagManager component
const TagManager = ({ onTagsChange, selectedTags = [] }) => {
    const [tags, setTags] = useState([]);
    const [isLoading, setIsLoading] = useState(true);
    const [showTagForm, setShowTagForm] = useState(false);
    const [newTag, setNewTag] = useState({
      name: "",
      category: "general",
      color: "#6366f1" // default indigo color
    });
    const [editingTag, setEditingTag] = useState(null);
    const [error, setError] = useState("");
  
    // Fetch user's tags on component mount
    useEffect(() => {
      fetchTags();
    }, []);
  
    const fetchTags = async () => {
      setIsLoading(true);
      try {
        const response = await fetch('/api/tags');
        if (!response.ok) {
          throw new Error('Failed to fetch tags');
        }
        const data = await response.json();
        setTags(data);
      } catch (error) {
        console.error('Error fetching tags:', error);
        setError("Failed to load tags. Please try again.");
      } finally {
        setIsLoading(false);
      }
    };
  
    const handleTagSelect = (tagId) => {
      if (selectedTags.includes(tagId)) {
        onTagsChange(selectedTags.filter(id => id !== tagId));
      } else {
        onTagsChange([...selectedTags, tagId]);
      }
    };
  
    const handleSubmitTag = async (e) => {
      e.preventDefault();
      setError("");
  
      if (!newTag.name.trim()) {
        setError("Tag name is required");
        return;
      }
  
      try {
        if (editingTag) {
          // Update existing tag
          const response = await fetch(`/api/tags/${editingTag.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newTag)
          });
          
          if (!response.ok) throw new Error('Failed to update tag');
          
          setTags(tags.map(tag => 
            tag.id === editingTag.id ? { ...tag, ...newTag } : tag
          ));
        } else {
          // Create new tag
          const response = await fetch('/api/tags', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newTag)
          });
          
          if (!response.ok) throw new Error('Failed to create tag');
          
          const createdTag = await response.json();
          setTags([...tags, createdTag]);
        }
        
        setNewTag({ name: "", category: "general", color: "#6366f1" });
        setShowTagForm(false);
        setEditingTag(null);
      } catch (error) {
        console.error('Error with tag operation:', error);
        setError(error.message);
      }
    };
  
    const handleEditTag = (tag) => {
      setEditingTag(tag);
      setNewTag({
        name: tag.name,
        category: tag.category,
        color: tag.color
      });
      setShowTagForm(true);
    };
  
    const handleDeleteTag = async (tagId) => {
      if (!window.confirm("Are you sure you want to delete this tag?")) return;
      
      try {
        const response = await fetch(`/api/tags/${tagId}`, {
          method: 'DELETE'
        });
        
        if (!response.ok) throw new Error('Failed to delete tag');
        
        setTags(tags.filter(tag => tag.id !== tagId));
        
        if (selectedTags.includes(tagId)) {
          onTagsChange(selectedTags.filter(id => id !== tagId));
        }
      } catch (error) {
        console.error('Error deleting tag:', error);
        setError("Failed to delete tag. Please try again.");
      }
    };

    const tagCategories = [
      { value: "general", label: "General" },
      { value: "strategy", label: "Strategy" },
      { value: "market", label: "Market" },
      { value: "performance", label: "Performance" },
      { value: "emotion", label: "Emotion" }
    ];

    const colors = [
      "#ef4444", // red
      "#f97316", // orange
      "#f59e0b", // amber
      "#84cc16", // lime
      "#10b981", // emerald
      "#06b6d4", // cyan
      "#3b82f6", // blue
      "#6366f1", // indigo
      "#8b5cf6", // violet
      "#d946ef", // fuchsia
      "#ec4899", // pink
      "#64748b"  // slate
    ];

    const cancelForm = () => {
      setShowTagForm(false);
      setEditingTag(null);
      setNewTag({ name: "", category: "general", color: "#6366f1" });
      setError("");
    };

    // Group tags by category for display
    const tagsByCategory = tags.reduce((acc, tag) => {
      if (!acc[tag.category]) {
        acc[tag.category] = [];
      }
      acc[tag.category].push(tag);
      return acc;
    }, {});

    return (
      <div className="mb-6 p-4 border border-gray-300 rounded-lg">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-bold text-lg flex items-center">
            <Tag className="mr-2" size={18} />
            Tags
          </h3>
          {!showTagForm && (
            <button
              type="button"
              onClick={() => setShowTagForm(true)}
              className="flex items-center text-sm bg-gray-100 hover:bg-gray-200 px-2 py-1 rounded transition-colors"
            >
              <Plus size={16} className="mr-1" /> Add Tag
            </button>
          )}
        </div>

        {error && (
          <div className="mb-4 p-2 bg-red-100 text-red-700 rounded text-sm">
            {error}
          </div>
        )}

        {isLoading ? (
          <div className="text-center py-4">Loading tags...</div>
        ) : (
          <>
            {/* Tag Creator/Editor Form */}
            {showTagForm && (
              <div className="mb-4 p-3 bg-gray-50 rounded border border-gray-200">
                <h4 className="font-medium mb-2">{editingTag ? "Edit Tag" : "Create New Tag"}</h4>
                <form onSubmit={handleSubmitTag}>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
                    <div>
                      <label className="block text-sm text-gray-600 mb-1">Tag Name</label>
                      <input
                        type="text"
                        value={newTag.name}
                        onChange={(e) => setNewTag({...newTag, name: e.target.value})}
                        className="w-full p-2 border border-gray-300 rounded"
                        placeholder="Enter tag name"
                      />
                    </div>
                    <div>
                      <label className="block text-sm text-gray-600 mb-1">Category</label>
                      <select
                        value={newTag.category}
                        onChange={(e) => setNewTag({...newTag, category: e.target.value})}
                        className="w-full p-2 border border-gray-300 rounded"
                      >
                        {tagCategories.map(category => (
                          <option key={category.value} value={category.value}>
                            {category.label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                  
                  <div className="mb-3">
                    <label className="block text-sm text-gray-600 mb-1">Color</label>
                    <div className="flex flex-wrap gap-2">
                      {colors.map(color => (
                        <div
                          key={color}
                          onClick={() => setNewTag({...newTag, color})}
                          className={`w-6 h-6 rounded-full cursor-pointer ${
                            newTag.color === color ? 'ring-2 ring-offset-2 ring-black' : ''
                          }`}
                          style={{ backgroundColor: color }}
                        />
                      ))}
                    </div>
                  </div>
                  
                  <div className="flex justify-end gap-2">
                    <button
                      type="button"
                      onClick={cancelForm}
                      className="px-3 py-1 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      className="px-3 py-1 bg-black text-white rounded hover:bg-gray-800"
                    >
                      {editingTag ? "Update" : "Create"}
                    </button>
                  </div>
                </form>
              </div>
            )}

            {/* Tags Display */}
            {tags.length === 0 ? (
              <div className="text-gray-500 text-center py-2">
                No tags created yet. Create your first tag to start organizing your trades.
              </div>
            ) : (
              <div>
                {Object.entries(tagsByCategory).map(([category, categoryTags]) => (
                  <div key={category} className="mb-3">
                    <h4 className="text-sm text-gray-600 mb-1 capitalize">{category}</h4>
                    <div className="flex flex-wrap gap-2">
                      {categoryTags.map(tag => (
                        <div 
                          key={tag.id} 
                          className="group relative flex items-center"
                        >
                          <div
                            onClick={() => handleTagSelect(tag.id)}
                            style={{ backgroundColor: tag.color + '20', borderColor: tag.color }}
                            className={`
                              text-sm px-2 py-1 rounded-md cursor-pointer 
                              border transition-all flex items-center
                              ${selectedTags.includes(tag.id) ? 'ring-2 ring-offset-1' : ''}
                            `}
                          >
                            {selectedTags.includes(tag.id) && <Check size={14} className="mr-1" />}
                            {tag.name}
                          </div>
                          
                          {/* Tag management buttons on hover */}
                          <div className="absolute right-0 top-0 -mt-2 -mr-2 hidden group-hover:flex bg-white rounded-full border border-gray-200 shadow-sm">
                            <button
                              type="button"
                              onClick={() => handleEditTag(tag)}
                              className="p-1 text-gray-600 hover:text-blue-600 rounded-full"
                              title="Edit tag"
                            >
                              <Edit2 size={12} />
                            </button>
                            <button
                              type="button"
                              onClick={() => handleDeleteTag(tag.id)}
                              className="p-1 text-gray-600 hover:text-red-600 rounded-full"
                              title="Delete tag"
                            >
                              <Trash size={12} />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    );
};



export default AddTrade;