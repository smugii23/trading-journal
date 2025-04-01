import { useState, useEffect } from "react";
import { Camera, AlertTriangle, CheckCircle, Tag, Plus, Check, Edit2, Trash } from "lucide-react";

const predefinedColors = [
  // Blues
  "#2563eb", // Blue
  "#0891b2", // Cyan
  "#0ea5e9", // Sky Blue
  
  // Greens
  "#16a34a", // Green
  "#059669", // Emerald
  "#65a30d", // Lime

  // Warm colors
  "#dc2626", // Red
  "#ea580c", // Orange
  "#d97706", // Amber
  "#ca8a04", // Yellow

  // Purples/Pinks
  "#9333ea", // Purple
  "#c026d3", // Fuchsia
  "#db2777", // Pink

  // Neutrals
  "#525252", // Gray
  "#44403c", // Stone
];

const FUTURES_CONTRACTS = {
  "ES": { tickValue: 12.50, minTickSize: 0.25 },     // E-mini S&P 500
  "GC": { tickValue: 10.0, minTickSize: 0.1 },      // Gold
};

const AddTrade = () => {
  // default state for the trade form
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

  // state variables to manage the component
  const [previewUrl, setPreviewUrl] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [notification, setNotification] = useState({ show: false, type: "", message: "" });
  const [profit, setProfit] = useState(null);
  const [selectedTags, setSelectedTags] = useState([]);
  const [tags, setTags] = useState([]);
  const [newTag, setNewTag] = useState({
    name: "",
    category: "",
    color: "#000000",
  });
  const [showTagForm, setShowTagForm] = useState(false);
  const [showResetModal, setShowResetModal] = useState(false);
  const [errors, setErrors] = useState({});
  const [isLoading, setIsLoading] = useState(false);

  // function to validate form fields before submission
  const validateForm = () => {
    const errors = {};
    if (!trade.ticker) errors.ticker = "Ticker is required";
    if (!trade.entry_price) errors.entry_price = "Entry price is required";
    if (!trade.quantity) errors.quantity = "Quantity is required";
    if (!trade.trade_date) errors.trade_date = "Trade date is required";
    if (!trade.entry_time) errors.entry_time = "Entry time is required";
    setErrors(errors);
    return Object.keys(errors).length === 0;
  };

  // fetch tags from the database
  useEffect(() => {
    let mounted = true; // prevent state updates after unmounting

    const fetchTags = async () => {
      if (!mounted) return;
      
      setIsLoading(true);
      try {
        // fetch tags from the database
        const response = await fetch("http://localhost:8080/api/tags");
        if (!response.ok) {
          throw new Error("Failed to fetch tags");
        }
        // if the tags are fetched successfully, set the tags state
        if (mounted) {
          const data = await response.json();
          setTags(Array.isArray(data) ? data : []); // ensure it's always an array
        }
      } catch (error) {
        if (mounted) {
          console.error("Error fetching tags:", error);
          setTags([]); //  set tags to an empty array on error
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    fetchTags();

    // cleanup function to prevent state updates after unmounting
    return () => {
      mounted = false;
    };
  }, []); // empty dependency array means this runs once on mount

  // handle tag selection
  const handleTagSelect = (tagId) => {
    // add or remove tag from the selected tags
    setSelectedTags((prev) => {
      if (prev.includes(tagId)) {
        // remove tag from the selected tags
        return prev.filter((id) => id !== tagId);
      } else {
        // add tag to the selected tags
        return [...prev, tagId];
      }
    });
  };

  // handle new tag submission
  const handleTagSubmit = async (e) => {
    // prevent default form submission
    e.preventDefault();
    if (!newTag.name) {
      setNotification({
        show: true,
        type: "error",
        message: "Tag name is required",
      });
      return;
    }

    try {
      // log the payload we're sending
      console.log("Sending tag data:", JSON.stringify(newTag, null, 2));
      
      // send the tag data to the database
      const response = await fetch("http://localhost:8080/api/tags", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(newTag),
      });

      // if there's an error, try to get the detailed error message
      if (!response.ok) {
        const errorText = await response.text();
        console.error("Server error response:", errorText);
        throw new Error(`Failed to create tag: ${errorText || response.statusText}`);
      }

      // parse the response text
      let createdTag;
      try {
        createdTag = JSON.parse(await response.text());
      } catch (parseError) {
        console.error("Error parsing tag response:", parseError);
        throw new Error("Invalid response format");
      }

      // make sure tags is always an array before spreading
      const currentTags = Array.isArray(tags) ? tags : [];
      setTags([...currentTags, createdTag]);
      
      // also make sure that selectedTags is an array
      const currentSelectedTags = Array.isArray(selectedTags) ? selectedTags : [];
      setSelectedTags([...currentSelectedTags, createdTag.id]);
      
      setNewTag({ name: "", category: "", color: "#000000" });
      setShowTagForm(false);
      
      setNotification({
        show: true,
        type: "success",
        message: "Tag created successfully",
      });
    } catch (error) {
      console.error("Error creating tag:", error);
      setNotification({
        show: true,
        type: "error",
        message: error.message || "Failed to create tag",
      });
    }
  };

// function to check if a ticker is a futures contract
const isFuturesContract = (ticker) => {
  return !!FUTURES_CONTRACTS[ticker?.toUpperCase()];
};

// calculate profit/loss
const calculateProfit = () => {
  if (trade.entry_price && trade.exit_price && trade.quantity) {
    const entry = parseFloat(trade.entry_price);
    const exit = parseFloat(trade.exit_price);
    const qty = parseFloat(trade.quantity);
    const comm = parseFloat(trade.commissions) || 0;
    const ticker = trade.ticker?.toUpperCase();
    
    // check if this is a futures contract
    if (isFuturesContract(ticker)) {
      const contract = FUTURES_CONTRACTS[ticker];
      
      // calculate price difference based on direction
      let priceDiff;
      if (trade.direction === "long") {
        priceDiff = exit - entry;
      } else {
        priceDiff = entry - exit;
      }
      
      // calculate number of ticks (rounded to nearest valid tick)
      const numTicks = Math.round(priceDiff / contract.minTickSize);
      
      // calculate profit loss based on number of ticks and tick value
      const profitLoss = numTicks * contract.tickValue * qty;
      
      return profitLoss - comm;
    } else {
      // for non-futures contracts, just use price difference * quantity
      let profitLoss;
      if (trade.direction === "long") {
        profitLoss = (exit - entry) * qty;
      } else {
        profitLoss = (entry - exit) * qty;
      }
      
      return profitLoss - comm;
    }
  }
  return null;
};

// recalculate profit when relevant fields change
useEffect(() => {
  const newProfit = calculateProfit();
  setProfit(newProfit);
}, [trade.entry_price, trade.exit_price, trade.quantity, trade.direction, trade.commissions, trade.ticker]);

// handle form field changes
const handleChange = (e) => {
  const { name, value } = e.target;
  setTrade((prevTrade) => ({ ...prevTrade, [name]: value }));
};

  // handle file upload
  const handleFileChange = (e) => {
    const file = e.target.files[0];
    // if there is a file, check the size and type
    if (file) {
      if (file.size > 5 * 1024 * 1024) {
        setNotification({
          show: true,
          type: "error",
          message: "File size exceeds 5MB limit",
        });
        return;
      }
      // make sure the file is an image
      if (!file.type.startsWith("image/")) {
        setNotification({
          show: true,
          type: "error",
          message: "Only image files are allowed",
        });
        return;
      }
      // set the screenshot to the file
      setTrade({ ...trade, screenshot: file });
      // create a preview url for the file
      setPreviewUrl(URL.createObjectURL(file));
    }
  };

  // handle form submission
  const handleSubmit = async (e) => {
    // prevent default form submission
    e.preventDefault();
    // validate the form fields
    if (!validateForm()) {
      setNotification({
        show: true,
        type: "error",
        message: "Please fix the errors in the form.",
      });
      return;
    }
    // set the isUploading state to true
    setIsUploading(true);

    try {
      // create a trade date
      const tradeDate = new Date(`${trade.trade_date}T${trade.entry_time}:00Z`).toISOString();
      const entryTime = new Date(`${trade.trade_date}T${trade.entry_time}:00Z`).toISOString();
      const exitTime = new Date(`${trade.trade_date}T${trade.exit_time}:00Z`).toISOString();
  
      const formData = new FormData();
      
      // add all trade data
      formData.append("ticker", trade.ticker);
      formData.append("direction", trade.direction.toUpperCase());
      formData.append("entry_price", trade.entry_price);
      formData.append("exit_price", trade.exit_price);
      formData.append("quantity", trade.quantity);
      formData.append("trade_date", tradeDate);
      formData.append("entry_time", entryTime);
      formData.append("exit_time", exitTime);
      formData.append("stop_loss", trade.stop_loss);
      formData.append("take_profit", trade.take_profit);
      formData.append("commissions", trade.commissions);
      formData.append("highest_price", trade.highest_price);
      formData.append("lowest_price", trade.lowest_price);
      formData.append("notes", trade.notes);
  
      // handle screenshot specially
      if (trade.screenshot) {
        formData.append("screenshot", trade.screenshot, trade.screenshot.name);
      }
  
      // send the trade data to the database
      const response = await fetch("http://localhost:8080/api/trades", {
        method: "POST",
        body: formData,
      });
  
      if (!response.ok) {
        const errorText = await response.text();
        console.error("Backend error:", errorText);
        throw new Error("Failed to submit trade");
      }
  
      const result = await response.json();
      
      // add selected tags to the newly created trade
      if (selectedTags.length > 0) {
        // get the trade id from the response
        const tradeId = result.id;
        // loop through the selected tags and add them to the trade
        for (const tagId of selectedTags) {
          try {
            await fetch(`http://localhost:8080/api/trades/${tradeId}/tags/${tagId}`, {
              method: "POST",
            });
          } catch (tagError) {
            console.error(`Error adding tag ${tagId} to trade:`, tagError);
          }
        }
      }

      // set the notification to show a success message
      setNotification({
        show: true,
        type: "success",
        message: "Trade added successfully!",
      });

      // reset the form
      resetForm();
    } catch (error) {
      console.error("Error submitting trade:", error);
      setNotification({
        show: true,
        type: "error",
        message: error.message || "Error submitting trade. Please try again.",
      });
    } finally {
      setIsUploading(false);
      setTimeout(() => setNotification({ show: false, type: "", message: "" }), 3000);
    }
  };

  // reset form
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
    setErrors({});
    setSelectedTags([]);
  };

  // confirmation modal for reset
  const handleResetClick = () => setShowResetModal(true);
  const handleResetConfirm = () => {
    resetForm();
    setShowResetModal(false);
  };

  return (
    <div className="max-w-5xl mx-auto mt-10 p-6 bg-white rounded-lg shadow">
      <h2 className="text-2xl font-bold mb-4">Add a New Trade</h2>

      {/* Notification */}
      {notification.show && (
        <div
          // set the notification color based on the type
          className={`mb-4 p-3 rounded flex items-center ${
            notification.type === "success" ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"
          }`}
        >
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
          <label htmlFor="ticker" className="block text-gray-700">
            Ticker
          </label>
          <input
            id="ticker"
            type="text"
            name="ticker"
            value={trade.ticker}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded mt-1"
            placeholder="Enter symbol (e.g., ES, NQ)"
            required
          />
          {errors.ticker && <p className="text-red-500 text-sm mt-1">{errors.ticker}</p>}
        </div>

        {/* Direction */}
        <div className="mb-4">
          <label htmlFor="direction" className="block text-gray-700">
            Direction
          </label>
          <select
            id="direction"
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
            <label htmlFor="entry_price" className="block text-gray-700">
              Entry Price
            </label>
            <input
              id="entry_price"
              type="number"
              name="entry_price"
              value={trade.entry_price}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              required
              step="any"
            />
            {errors.entry_price && <p className="text-red-500 text-sm mt-1">{errors.entry_price}</p>}
          </div>
          <div>
            <label htmlFor="exit_price" className="block text-gray-700">
              Exit Price
            </label>
            <input
              id="exit_price"
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
          <label htmlFor="quantity" className="block text-gray-700">
            Quantity
          </label>
          <input
            id="quantity"
            type="number"
            name="quantity"
            value={trade.quantity}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded mt-1"
            required
            min="1"
            step="any"
          />
          {errors.quantity && <p className="text-red-500 text-sm mt-1">{errors.quantity}</p>}
        </div>

        {/* Trade Date, Entry Time, Exit Time */}
        <div className="grid grid-cols-3 gap-4 mb-4">
          <div>
            <label htmlFor="trade_date" className="block text-gray-700">
              Trade Date
            </label>
            <input
              id="trade_date"
              type="date"
              name="trade_date"
              value={trade.trade_date}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              required
            />
            {errors.trade_date && <p className="text-red-500 text-sm mt-1">{errors.trade_date}</p>}
          </div>
          <div>
            <label htmlFor="entry_time" className="block text-gray-700">
              Entry Time
            </label>
            <input
              id="entry_time"
              type="time"
              name="entry_time"
              value={trade.entry_time}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              required
            />
            {errors.entry_time && <p className="text-red-500 text-sm mt-1">{errors.entry_time}</p>}
          </div>
          <div>
            <label htmlFor="exit_time" className="block text-gray-700">
              Exit Time
            </label>
            <input
              id="exit_time"
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
            <label htmlFor="stop_loss" className="block text-gray-700">
              Stop Loss
            </label>
            <input
              id="stop_loss"
              type="number"
              name="stop_loss"
              value={trade.stop_loss}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
          <div>
            <label htmlFor="take_profit" className="block text-gray-700">
              Take Profit
            </label>
            <input
              id="take_profit"
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
          <label htmlFor="commissions" className="block text-gray-700">
            Commissions
          </label>
          <div className="relative">
            <input
              id="commissions"
              type="number"
              name="commissions"
              value={trade.commissions}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
            <div className="absolute top-0 right-0 mt-6 mr-2">
              <div className="group relative">
                <svg className="h-4 w-4 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
                  <path
                    fillRule="evenodd"
                    d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                    clipRule="evenodd"
                  />
                </svg>
                <div className="hidden group-hover:block absolute right-0 mt-2 w-48 bg-black text-white text-sm p-2 rounded">
                  Total commissions paid for this trade.
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Highest & Lowest Price */}
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label htmlFor="highest_price" className="block text-gray-700">
              Highest Price
            </label>
            <input
              id="highest_price"
              type="number"
              name="highest_price"
              value={trade.highest_price}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
          <div>
            <label htmlFor="lowest_price" className="block text-gray-700">
              Lowest Price
            </label>
            <input
              id="lowest_price"
              type="number"
              name="lowest_price"
              value={trade.lowest_price}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded mt-1"
              step="any"
            />
          </div>
        </div>
        

        {/* Profit/Loss Calculator */}
        {profit !== null && (
          <div className="mb-4 p-3 rounded border border-gray-300">
            <h3 className="font-bold">Estimated P&L:</h3>
            <p className={profit >= 0 ? "text-green-600 font-bold" : "text-red-600 font-bold"}>
              {profit >= 0 ? "+" : ""}
              {profit.toFixed(2)}
            </p>
          </div>
        )}

        {/* Tags Section */}
        <div className="mb-4">
          <label className="block text-gray-700 mb-2">Tags</label>
          <div className="flex flex-wrap gap-2 mb-2">
            {isLoading ? (
              <span className="text-gray-500">Loading tags...</span>
            ) : (tags?.length === 0) ? (
              <span className="text-gray-500">No tags available</span>
            ) : (
              tags?.map((tag) => (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => handleTagSelect(tag.id)}
                  className={`px-3 py-1 rounded-full text-sm flex items-center ${
                    selectedTags.includes(tag.id)
                      ? "bg-gray-700 text-white"
                      : "bg-gray-200 text-gray-800"
                  }`}
                  style={{
                    backgroundColor: selectedTags.includes(tag.id) ? tag.color : "#f3f4f6",
                    color: selectedTags.includes(tag.id) ? "#ffffff" : "#1f2937",
                  }}
                >
                  {tag.name}
                  {selectedTags.includes(tag.id) && (
                    <Check className="ml-1" size={14} />
                  )}
                </button>
              )) || <span className="text-gray-500">No tags available</span>
            )}
            <button
              type="button"
              onClick={() => setShowTagForm(!showTagForm)}
              className="px-3 py-1 rounded-full bg-gray-200 text-gray-800 text-sm flex items-center"
            >
              <Plus size={14} className="mr-1" /> New Tag
            </button>
          </div>

          {/* New Tag Form */}
          {showTagForm && (
            <div className="p-3 border border-gray-300 rounded mb-3">
              <h4 className="font-bold mb-2">Create New Tag</h4>
              <div className="grid grid-cols-3 gap-2 mb-2">
                <div>
                  <label htmlFor="tagName" className="block text-sm text-gray-700">
                    Name
                  </label>
                  <input
                    id="tagName"
                    type="text"
                    value={newTag.name}
                    onChange={(e) => setNewTag({ ...newTag, name: e.target.value })}
                    className="w-full p-2 border border-gray-300 rounded mt-1 text-sm"
                    placeholder="Tag name"
                  />
                </div>
                <div>
                  <label htmlFor="tagCategory" className="block text-sm text-gray-700">
                    Category
                  </label>
                  <input
                    id="tagCategory"
                    type="text"
                    value={newTag.category}
                    onChange={(e) => setNewTag({ ...newTag, category: e.target.value })}
                    className="w-full p-2 border border-gray-300 rounded mt-1 text-sm"
                    placeholder="Category (optional)"
                  />
                </div>
                <div>
                  <label htmlFor="tagColor" className="block text-sm text-gray-700">
                    Color
                  </label>
                  <div className="mt-1 grid grid-cols-5 gap-2">
                    {predefinedColors.map((color) => (
                      <button
                        key={color}
                        type="button"
                        onClick={() => setNewTag({ ...newTag, color })}
                        className={`w-full h-8 rounded-md border ${
                          newTag.color === color ? 'ring-2 ring-offset-2 ring-black' : 'border-gray-300'
                        }`}
                        style={{ backgroundColor: color }}
                        aria-label={`Select color ${color}`}
                      />
                    ))}
                  </div>
                </div>
              </div>
              <div className="flex justify-end">
                <button
                  type="button"
                  onClick={() => setShowTagForm(false)}
                  className="px-3 py-1 bg-gray-200 text-gray-800 rounded mr-2 text-sm"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleTagSubmit}
                  className="px-3 py-1 bg-black text-white rounded text-sm"
                >
                  Add Tag
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Notes */}
        <div className="mb-4">
          <label htmlFor="notes" className="block text-gray-700">
            Notes
          </label>
          <textarea
            id="notes"
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
          <label htmlFor="screenshot" className="block text-gray-700">
            Trade Screenshot
          </label>
          <div className="mt-1 flex items-center space-x-4">
            <div className="flex-1">
              <input
                id="screenshot"
                type="file"
                accept="image/*"
                onChange={handleFileChange}
                className="w-full p-2 border border-gray-300 rounded"
              />
            </div>
          </div>
          {previewUrl && (
            <div className="mt-2">
              <div className="relative group">
                <img
                  src={previewUrl}
                  alt="Screenshot preview"
                  className="max-h-64 object-contain rounded border border-gray-300"
                />
                <button
                  type="button"
                  className="absolute top-2 right-2 bg-red-500 text-white p-1 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                  onClick={() => {
                    setTrade({ ...trade, screenshot: null });
                    setPreviewUrl(null);
                  }}
                >
                  <Trash size={16} />
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="grid grid-cols-2 gap-4">
          <button
            type="button"
            onClick={handleResetClick}
            className="w-full bg-gray-200 text-gray-800 p-2 rounded mt-4 hover:bg-gray-300 transition-colors"
          >
            Reset Form
          </button>
          <button
            type="submit"
            className="w-full bg-black text-white p-2 rounded mt-4 hover:bg-gray-800 transition-colors disabled:bg-gray-400"
            disabled={isUploading}
          >
            {isUploading ? (
              <div className="flex items-center justify-center">
                <svg className="animate-spin h-5 w-5 mr-2" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                Submitting...
              </div>
            ) : (
              "Submit Trade"
            )}
          </button>
        </div>
      </form>

      {/* Reset Confirmation Modal */}
      {showResetModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg">
            <h3 className="text-lg font-bold mb-4">Are you sure?</h3>
            <p className="mb-4">This will reset all fields and clear your input.</p>
            <div className="flex justify-end space-x-4">
              <button
                onClick={() => setShowResetModal(false)}
                className="bg-gray-200 text-gray-800 px-4 py-2 rounded hover:bg-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={handleResetConfirm}
                className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
              >
                Reset
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AddTrade;