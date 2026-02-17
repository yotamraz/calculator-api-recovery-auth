"""FastAPI application creation, lifespan, and router wiring."""

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI
from pydantic import BaseModel

from .database import init_db


class HealthResponse(BaseModel):
    """Response model for the health check endpoint."""

    status: str
    version: str


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    """Create database tables on startup."""
    init_db()
    yield


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(title="Calculator API", version="0.1.0", lifespan=lifespan)

    @app.get("/health", response_model=HealthResponse)
    def health_check() -> HealthResponse:
        """Return the health status of the service."""
        return HealthResponse(status="ok", version=app.version)

    return app


app = create_app()


def run() -> None:
    """Run the server with uvicorn."""
    import uvicorn

    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)
