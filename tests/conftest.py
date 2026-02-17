"""Shared test fixtures for the calculator API test suite."""

import pytest
from fastapi.testclient import TestClient
from sqlmodel import Session, SQLModel, create_engine

from app.database import get_session
from app.main import create_app


@pytest.fixture(name="session")
def session_fixture():
    """Create an in-memory SQLite database session for tests."""
    engine = create_engine("sqlite://", echo=False)
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
