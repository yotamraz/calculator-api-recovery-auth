"""Shared fixtures for tests."""

import pytest
from fastapi.testclient import TestClient
from sqlalchemy.pool import StaticPool
from sqlmodel import Session, SQLModel, create_engine

from app.database import get_session
from app.main import create_app


@pytest.fixture(name="session")
def session_fixture():
    """Create an in-memory SQLite database session for testing."""
    engine = create_engine(
        "sqlite://",
        echo=False,
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    SQLModel.metadata.create_all(engine)
    with Session(engine) as session:
        yield session


@pytest.fixture(name="client")
def client_fixture(session: Session):
    """Create a test client with overridden DB session."""
    app = create_app()

    def get_session_override():
        yield session

    app.dependency_overrides[get_session] = get_session_override
    with TestClient(app) as client:
        yield client
    app.dependency_overrides.clear()


@pytest.fixture(name="auth_token")
def auth_token_fixture(client: TestClient) -> str:
    """Register a test user and return a valid JWT token."""
    client.post(
        "/auth/register",
        json={"username": "testuser", "password": "testpassword"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "testuser", "password": "testpassword"},
    )
    return response.json()["access_token"]


@pytest.fixture(name="auth_client")
def auth_client_fixture(client: TestClient, auth_token: str) -> TestClient:
    """Return the test client with Authorization header pre-set."""
    client.headers["Authorization"] = f"Bearer {auth_token}"
    return client
