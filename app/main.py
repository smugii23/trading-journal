from fastapi import FastAPI
import uvicorn
from app.routers import analysis

app = FastAPI(
    title="Trading Journal Analysis API",
    description="Provides endpoints for analyzing trading performance (Tier 2+).",
    version="0.2.0"
)

app.include_router(analysis.router)

@app.get("/")
async def root():
    """Basic health check endpoint."""
    return {"message": "Trading Analysis API is running."}

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
