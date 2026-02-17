"""Tests for the health check endpoint."""


def test_health_returns_ok(client):
    """GET /health returns status ok and the correct version."""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data == {"status": "ok", "version": "0.1.0"}
