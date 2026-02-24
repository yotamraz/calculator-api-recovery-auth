"""FastAPI application factory, lifespan, and router registration."""

from contextlib import asynccontextmanager
from typing import AsyncGenerator

import uvicorn
from fastapi import FastAPI

from .database import init_db
from .models import HealthResponse


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Create database tables on startup."""
    init_db()
    yield


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(title="Calculator API", version="0.1.0", lifespan=lifespan)

    # --- Health Check ---

    @app.get("/health", response_model=HealthResponse)
    def health_check() -> HealthResponse:
        """Return the health status of the service."""
        return HealthResponse(status="ok", version=app.version)

    # Router registration stubs — routers will be added in future milestones:
    # app.include_router(auth_routes.router, prefix="/auth", tags=["auth"])
    # app.include_router(calculator_routes.router, tags=["calculator"])
    # app.include_router(calculations_routes.router, prefix="/calculations", tags=["calculations"])

    return app


app = create_app()


def main() -> None:
    """Run the server."""
    uvicorn.run(app, host="0.0.0.0", port=8000)


if __name__ == "__main__":
    main()
