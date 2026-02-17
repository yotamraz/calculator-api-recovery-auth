"""Tests for the health endpoint."""

from fastapi.testclient import TestClient


def test_health_endpoint(client: TestClient):
    """GET /health returns status ok and version 0.1.0."""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data == {"status": "ok", "version": "0.1.0"}
