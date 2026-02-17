"""FastAPI application creation, lifespan, and router wiring."""

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI

from .database import init_db
from .models import HealthResponse
from .routes import auth as auth_routes
from .routes import calculator as calculator_routes
from .routes import calculations as calculations_routes


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    """Create database tables on startup."""
    init_db()
    yield


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(title="Calculator API", version="0.1.0", lifespan=lifespan)

    app.include_router(auth_routes.router)
    app.include_router(calculator_routes.router)
    app.include_router(calculations_routes.router)

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
