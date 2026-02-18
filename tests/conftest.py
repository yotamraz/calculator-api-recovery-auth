"""Shared test fixtures for the calculator API test suite."""

import pytest
from fastapi.testclient import TestClient
from sqlmodel import Session, SQLModel, create_engine

from app.database import get_session
from app.main import create_app


@pytest.fixture(name="session")
def session_fixture():
    """Create an in-memory SQLite database session for testing."""
    engine = create_engine("sqlite://", echo=False)
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
