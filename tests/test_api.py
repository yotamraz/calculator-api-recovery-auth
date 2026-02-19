"""API integration tests for health and calculator operation endpoints."""


class TestHealthEndpoint:
    """Tests for the GET /health endpoint."""

    def test_health_returns_200(self, client):
        response = client.get("/health")
        assert response.status_code == 200

    def test_health_returns_status_ok(self, client):
        data = client.get("/health").json()
        assert data["status"] == "ok"

    def test_health_returns_version(self, client):
        data = client.get("/health").json()
        assert data["version"] == "0.1.0"


class TestAddEndpoint:
    """Tests for the POST /add endpoint."""

    def test_add_returns_200(self, client):
        response = client.post("/add", json={"a": 2, "b": 3})
        assert response.status_code == 200

    def test_add_result(self, client):
        data = client.post("/add", json={"a": 2, "b": 3}).json()
        assert data["result"] == 5.0

    def test_add_negative_numbers(self, client):
        data = client.post("/add", json={"a": -1, "b": -2}).json()
        assert data["result"] == -3.0


class TestSubtractEndpoint:
    """Tests for the POST /subtract endpoint."""

    def test_subtract_returns_200(self, client):
        response = client.post("/subtract", json={"a": 5, "b": 3})
        assert response.status_code == 200

    def test_subtract_result(self, client):
        data = client.post("/subtract", json={"a": 5, "b": 3}).json()
        assert data["result"] == 2.0


class TestMultiplyEndpoint:
    """Tests for the POST /multiply endpoint."""

    def test_multiply_returns_200(self, client):
        response = client.post("/multiply", json={"a": 4, "b": 5})
        assert response.status_code == 200

    def test_multiply_result(self, client):
        data = client.post("/multiply", json={"a": 4, "b": 5}).json()
        assert data["result"] == 20.0


class TestDivideEndpoint:
    """Tests for the POST /divide endpoint."""

    def test_divide_returns_200(self, client):
        response = client.post("/divide", json={"a": 10, "b": 2})
        assert response.status_code == 200

    def test_divide_result(self, client):
        data = client.post("/divide", json={"a": 10, "b": 2}).json()
        assert data["result"] == 5.0

    def test_divide_by_zero_returns_400(self, client):
        response = client.post("/divide", json={"a": 10, "b": 0})
        assert response.status_code == 400

    def test_divide_by_zero_error_message(self, client):
        data = client.post("/divide", json={"a": 10, "b": 0}).json()
        assert data["detail"] == "Cannot divide by zero"
