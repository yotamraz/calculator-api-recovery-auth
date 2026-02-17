"""Shared test fixtures for the calculator API test suite."""

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
        echo=False,
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    SQLModel.metadata.create_all(engine)
    with Session(engine) as session:
        yield session


@pytest.fixture(name="app")
def app_fixture(session: Session):
    """Create a FastAPI app with the test database session override."""
    app = create_app()

    def get_session_override():
        yield session

    app.dependency_overrides[get_session] = get_session_override
    yield app
    app.dependency_overrides.clear()


@pytest.fixture(name="client")
def client_fixture(app):
    """Create a TestClient for the FastAPI app."""
    with TestClient(app) as client:
        yield client


@pytest.fixture(name="registered_user")
def registered_user_fixture(client):
    """Register a test user and return the credentials dict."""
    credentials = {"username": "testuser", "password": "testpassword123"}
    response = client.post("/auth/register", json=credentials)
    assert response.status_code == 201
    return credentials


@pytest.fixture(name="auth_token")
def auth_token_fixture(client, registered_user):
    """Obtain a JWT token for the registered test user."""
    response = client.post(
        "/auth/token",
        data={
            "username": registered_user["username"],
            "password": registered_user["password"],
        },
    )
    assert response.status_code == 200
    return response.json()["access_token"]


@pytest.fixture(name="auth_client")
def auth_client_fixture(client, auth_token):
    """Return the test client with Authorization header set."""
    client.headers["Authorization"] = f"Bearer {auth_token}"
    return client
