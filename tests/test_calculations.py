"""Tests for calculations CRUD endpoints."""


class TestCreateCalculation:
    """Tests for POST /calculations."""

    def test_create_valid_operation(self, auth_client):
        response = auth_client.post(
            "/calculations",
            json={"operation": "add", "a": 5, "b": 3},
        )
        assert response.status_code == 201
        data = response.json()
        assert data["operation"] == "add"
        assert data["a"] == 5.0
        assert data["b"] == 3.0
        assert data["result"] == 8.0
        assert "id" in data
        assert "created_at" in data

    def test_create_invalid_operation(self, auth_client):
        response = auth_client.post(
            "/calculations",
            json={"operation": "mod", "a": 5, "b": 3},
        )
        assert response.status_code == 400
        assert "Unknown operation" in response.json()["detail"]

    def test_create_unauthenticated(self, client):
        response = client.post(
            "/calculations",
            json={"operation": "add", "a": 1, "b": 2},
        )
        assert response.status_code == 401


class TestListCalculations:
    """Tests for GET /calculations."""

    def test_list_empty(self, auth_client):
        response = auth_client.get("/calculations")
        assert response.status_code == 200
        assert response.json() == []

    def test_list_after_create(self, auth_client):
        auth_client.post(
            "/calculations",
            json={"operation": "mul", "a": 7, "b": 6},
        )
        response = auth_client.get("/calculations")
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 1
        assert data[0]["result"] == 42.0

    def test_list_ordered_by_created_at_desc(self, auth_client):
        auth_client.post(
            "/calculations",
            json={"operation": "add", "a": 1, "b": 1},
        )
        auth_client.post(
            "/calculations",
            json={"operation": "sub", "a": 10, "b": 5},
        )
        response = auth_client.get("/calculations")
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 2
        # Most recent first
        assert data[0]["result"] == 5.0
        assert data[1]["result"] == 2.0

    def test_list_unauthenticated(self, client):
        response = client.get("/calculations")
        assert response.status_code == 401


class TestGetCalculation:
    """Tests for GET /calculations/{id}."""

    def test_get_existing(self, auth_client):
        create_resp = auth_client.post(
            "/calculations",
            json={"operation": "add", "a": 3, "b": 4},
        )
        calc_id = create_resp.json()["id"]

        response = auth_client.get(f"/calculations/{calc_id}")
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == calc_id
        assert data["result"] == 7.0

    def test_get_not_found(self, auth_client):
        response = auth_client.get("/calculations/9999")
        assert response.status_code == 404
        assert response.json()["detail"] == "Calculation not found"

    def test_get_unauthenticated(self, client):
        response = client.get("/calculations/1")
        assert response.status_code == 401


class TestDeleteCalculation:
    """Tests for DELETE /calculations/{id}."""

    def test_delete_existing(self, auth_client):
        create_resp = auth_client.post(
            "/calculations",
            json={"operation": "sub", "a": 10, "b": 3},
        )
        calc_id = create_resp.json()["id"]

        response = auth_client.delete(f"/calculations/{calc_id}")
        assert response.status_code == 204

        # Verify it's gone
        get_resp = auth_client.get(f"/calculations/{calc_id}")
        assert get_resp.status_code == 404

    def test_delete_not_found(self, auth_client):
        response = auth_client.delete("/calculations/9999")
        assert response.status_code == 404
        assert response.json()["detail"] == "Calculation not found"

    def test_delete_unauthenticated(self, client):
        response = client.delete("/calculations/1")
        assert response.status_code == 401
