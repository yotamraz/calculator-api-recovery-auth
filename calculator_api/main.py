"""FastAPI application creation, lifespan, and router inclusion."""

from contextlib import asynccontextmanager
from typing import AsyncGenerator

import uvicorn
from fastapi import FastAPI

from .database import init_db
from .models import HealthResponse
from .routes import auth as auth_routes
from .routes import calculations as calc_routes


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Create database tables on startup."""
    init_db()
    yield


app = FastAPI(title="Calculator API", version="0.1.0", lifespan=lifespan)


# --- Health Check ---


@app.get("/health", response_model=HealthResponse)
def health_check() -> HealthResponse:
    """Return the health status of the service."""
    return HealthResponse(status="ok", version=app.version)


# --- Include Routers ---

app.include_router(auth_routes.router, prefix="/auth", tags=["auth"])
app.include_router(calc_routes.router, tags=["calculations"])


def main() -> None:
    """Run the server."""
    uvicorn.run(app, host="0.0.0.0", port=8000)


if __name__ == "__main__":
    main()
