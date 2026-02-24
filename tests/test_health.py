"""Integration tests for the /health endpoint."""


class TestHealthCheck:
    """Tests for the health check endpoint."""

    def test_health_returns_200(self, client):
        response = client.get("/health")
        assert response.status_code == 200

    def test_health_returns_expected_payload(self, client):
        response = client.get("/health")
        data = response.json()
        assert data["status"] == "ok"
        assert data["version"] == "0.1.0"

    def test_health_no_auth_required(self, client):
        """Health endpoint must be accessible without authentication."""
        response = client.get("/health")
        assert response.status_code == 200
