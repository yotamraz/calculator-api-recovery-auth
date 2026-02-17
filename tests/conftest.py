"""Shared test fixtures."""

import pytest
from fastapi.testclient import TestClient
from sqlalchemy.pool import StaticPool
from sqlmodel import Session, SQLModel, create_engine

from app.database import get_session
from app.main import create_app


@pytest.fixture(name="session")
def session_fixture():
    """Create an in-memory SQLite database session for tests."""
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
    """Create a TestClient with the test database session."""
    app = create_app()

    def get_session_override():
        return session

    app.dependency_overrides[get_session] = get_session_override
    client = TestClient(app)
    yield client
    app.dependency_overrides.clear()


@pytest.fixture(name="test_user")
def test_user_fixture(client: TestClient) -> dict:
    """Register a test user and return the user data."""
    response = client.post(
        "/auth/register",
        json={"username": "testuser", "password": "testpass123"},
    )
    assert response.status_code == 201
    return response.json()


@pytest.fixture(name="auth_token")
def auth_token_fixture(client: TestClient, test_user: dict) -> str:
    """Obtain a JWT token for the test user."""
    response = client.post(
        "/auth/token",
        data={"username": "testuser", "password": "testpass123"},
    )
    assert response.status_code == 200
    return response.json()["access_token"]


@pytest.fixture(name="auth_headers")
def auth_headers_fixture(auth_token: str) -> dict:
    """Return authorization headers with a valid Bearer token."""
    return {"Authorization": f"Bearer {auth_token}"}
