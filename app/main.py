"""FastAPI application factory, health endpoint, and startup logic."""

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI
from pydantic import BaseModel

from .database import init_db
from .routes import auth as auth_routes


class HealthResponse(BaseModel):
    """Response model for the health check endpoint."""

    status: str
    version: str


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""

    @asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        """Create database tables on startup."""
        init_db()
        yield

    application = FastAPI(title="Calculator API", version="0.1.0", lifespan=lifespan)

    application.include_router(auth_routes.router)

    @application.get("/health", response_model=HealthResponse)
    def health_check() -> HealthResponse:
        """Return the health status of the service."""
        return HealthResponse(status="ok", version=application.version)

    return application


app = create_app()


def run() -> None:
    """Run the application with uvicorn."""
    import uvicorn

    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)
