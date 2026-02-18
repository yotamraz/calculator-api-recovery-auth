"""Tests for calculation history CRUD endpoints."""

from fastapi.testclient import TestClient


class TestCreateCalculation:
    """Tests for POST /calculations."""

    def test_requires_auth(self, client: TestClient):
        response = client.post("/calculations", json={"operation": "add", "a": 1, "b": 2})
        assert response.status_code == 401

    def test_create_add(self, client: TestClient, auth_headers: dict):
        response = client.post(
            "/calculations",
            json={"operation": "add", "a": 5, "b": 3},
            headers=auth_headers,
        )
        assert response.status_code == 201
        data = response.json()
        assert data["operation"] == "add"
        assert data["a"] == 5.0
        assert data["b"] == 3.0
        assert data["result"] == 8.0
        assert "id" in data
        assert "created_at" in data

    def test_create_sub(self, client: TestClient, auth_headers: dict):
        response = client.post(
            "/calculations",
            json={"operation": "sub", "a": 10, "b": 4},
            headers=auth_headers,
        )
        assert response.status_code == 201
        assert response.json()["result"] == 6.0

    def test_create_mul(self, client: TestClient, auth_headers: dict):
        response = client.post(
            "/calculations",
            json={"operation": "mul", "a": 7, "b": 6},
            headers=auth_headers,
        )
        assert response.status_code == 201
        assert response.json()["result"] == 42.0

    def test_create_div(self, client: TestClient, auth_headers: dict):
        response = client.post(
            "/calculations",
            json={"operation": "div", "a": 10, "b": 4},
            headers=auth_headers,
        )
        assert response.status_code == 201
        assert response.json()["result"] == 2.5

    def test_unknown_operation(self, client: TestClient, auth_headers: dict):
        response = client.post(
            "/calculations",
            json={"operation": "modulo", "a": 5, "b": 3},
            headers=auth_headers,
        )
        assert response.status_code == 400
        assert "Unknown operation" in response.json()["detail"]

    def test_division_by_zero(self, client: TestClient, auth_headers: dict):
        response = client.post(
            "/calculations",
            json={"operation": "div", "a": 1, "b": 0},
            headers=auth_headers,
        )
        assert response.status_code == 400
        assert response.json()["detail"] == "Cannot divide by zero"


class TestListCalculations:
    """Tests for GET /calculations."""

    def test_requires_auth(self, client: TestClient):
        response = client.get("/calculations")
        assert response.status_code == 401

    def test_empty_list(self, client: TestClient, auth_headers: dict):
        response = client.get("/calculations", headers=auth_headers)
        assert response.status_code == 200
        assert response.json() == []

    def test_list_includes_created(self, client: TestClient, auth_headers: dict):
        client.post(
            "/calculations",
            json={"operation": "add", "a": 1, "b": 2},
            headers=auth_headers,
        )
        client.post(
            "/calculations",
            json={"operation": "mul", "a": 3, "b": 4},
            headers=auth_headers,
        )
        response = client.get("/calculations", headers=auth_headers)
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 2


class TestGetCalculation:
    """Tests for GET /calculations/{id}."""

    def test_requires_auth(self, client: TestClient):
        response = client.get("/calculations/1")
        assert response.status_code == 401

    def test_get_existing(self, client: TestClient, auth_headers: dict):
        create_resp = client.post(
            "/calculations",
            json={"operation": "add", "a": 10, "b": 20},
            headers=auth_headers,
        )
        calc_id = create_resp.json()["id"]
        response = client.get(f"/calculations/{calc_id}", headers=auth_headers)
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == calc_id
        assert data["operation"] == "add"
        assert data["result"] == 30.0

    def test_get_nonexistent_returns_404(self, client: TestClient, auth_headers: dict):
        response = client.get("/calculations/9999", headers=auth_headers)
        assert response.status_code == 404
        assert response.json()["detail"] == "Calculation not found"


class TestDeleteCalculation:
    """Tests for DELETE /calculations/{id}."""

    def test_requires_auth(self, client: TestClient):
        response = client.delete("/calculations/1")
        assert response.status_code == 401

    def test_delete_existing(self, client: TestClient, auth_headers: dict):
        create_resp = client.post(
            "/calculations",
            json={"operation": "sub", "a": 20, "b": 5},
            headers=auth_headers,
        )
        calc_id = create_resp.json()["id"]
        response = client.delete(f"/calculations/{calc_id}", headers=auth_headers)
        assert response.status_code == 204

        # Verify it's actually deleted
        get_resp = client.get(f"/calculations/{calc_id}", headers=auth_headers)
        assert get_resp.status_code == 404

    def test_delete_nonexistent_returns_404(self, client: TestClient, auth_headers: dict):
        response = client.delete("/calculations/9999", headers=auth_headers)
        assert response.status_code == 404
        assert response.json()["detail"] == "Calculation not found"
