"""Shared test fixtures for the Calculator API test suite."""

from collections.abc import Generator

import pytest
from fastapi.testclient import TestClient
from sqlmodel import Session, SQLModel, create_engine

from calculator_api.database import get_db
from calculator_api.main import create_app

# In-memory SQLite engine for test isolation
_test_engine = create_engine("sqlite://", echo=False)


def _get_test_db() -> Generator[Session, None, None]:
    """Yield an in-memory test database session."""
    with Session(_test_engine) as session:
        yield session


@pytest.fixture(name="app")
def app_fixture():
    """Create a FastAPI test app with in-memory SQLite database."""
    # Create all tables in the test database
    SQLModel.metadata.create_all(_test_engine)

    application = create_app()
    application.dependency_overrides[get_db] = _get_test_db

    yield application

    # Clean up: drop all tables after the test
    SQLModel.metadata.drop_all(_test_engine)
    application.dependency_overrides.clear()


@pytest.fixture(name="client")
def client_fixture(app):
    """Provide a TestClient bound to the test app."""
    with TestClient(app) as client:
        yield client
