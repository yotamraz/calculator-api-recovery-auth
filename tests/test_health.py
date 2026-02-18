"""Tests for the /health endpoint."""

from fastapi.testclient import TestClient


def test_health_returns_ok(client: TestClient):
    """GET /health returns 200 with status ok and version."""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert data["version"] == "0.1.0"


def test_health_response_structure(client: TestClient):
    """GET /health response contains exactly the expected keys."""
    response = client.get("/health")
    data = response.json()
    assert set(data.keys()) == {"status", "version"}
