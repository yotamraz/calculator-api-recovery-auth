"""Shared test fixtures for the calculator API test suite."""

import pytest
from fastapi.testclient import TestClient
from sqlalchemy.pool import StaticPool
from sqlmodel import Session, SQLModel, create_engine

from app.database import get_session
from app.main import create_app
from app.models import Calculation, User  # noqa: F401 — register table models


@pytest.fixture(name="session")
def session_fixture():
    """Create an in-memory SQLite database session for testing."""
    engine = create_engine(
        "sqlite://",
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    SQLModel.metadata.create_all(engine)
    with Session(engine) as session:
        yield session


@pytest.fixture(name="client")
def client_fixture(session: Session):
    """Create a test client with the in-memory database session override."""
    application = create_app()

    def get_session_override():
        yield session

    application.dependency_overrides[get_session] = get_session_override
    with TestClient(application) as client:
        yield client


@pytest.fixture(name="auth_token")
def auth_token_fixture(client: TestClient) -> str:
    """Register a test user and return a valid JWT access token."""
    client.post(
        "/auth/register",
        json={"username": "testuser", "password": "testpass"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "testuser", "password": "testpass"},
    )
    return response.json()["access_token"]


@pytest.fixture(name="auth_headers")
def auth_headers_fixture(auth_token: str) -> dict[str, str]:
    """Return Authorization headers with a valid Bearer token."""
    return {"Authorization": f"Bearer {auth_token}"}
