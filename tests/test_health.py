"""Tests for the health check endpoint."""

from fastapi.testclient import TestClient

from calculator_api.main import app


def test_health_endpoint(client: TestClient):
    """GET /health returns status ok and version."""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert data["version"] == "0.1.0"
