"""Shared test fixtures for the calculator API test suite."""

import pytest
from fastapi.testclient import TestClient
from sqlmodel import Session, SQLModel, create_engine
from sqlmodel.pool import StaticPool

from app.database import get_session
from app.main import create_app


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


@pytest.fixture(name="app")
def app_fixture(session: Session):
    """Create a FastAPI app with test database session override."""
    application = create_app()

    def get_session_override():
        yield session

    application.dependency_overrides[get_session] = get_session_override
    yield application
    application.dependency_overrides.clear()


@pytest.fixture(name="client")
def client_fixture(app):
    """Create a TestClient for the app."""
    return TestClient(app)


@pytest.fixture(name="auth_headers")
def auth_headers_fixture(client: TestClient) -> dict[str, str]:
    """Register a test user and return Authorization headers with a valid token."""
    client.post(
        "/auth/register",
        json={"username": "testuser", "password": "testpass123"},
    )
    response = client.post(
        "/auth/token",
        data={"username": "testuser", "password": "testpass123"},
    )
    token = response.json()["access_token"]
    return {"Authorization": f"Bearer {token}"}
