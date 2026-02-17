"""Tests for CRUD endpoints on calculation history."""

from fastapi.testclient import TestClient


def test_create_calculation(auth_client: TestClient):
    """POST /calculations with a valid operation returns 201."""
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


def test_create_calculation_unknown_operation(auth_client: TestClient):
    """POST /calculations with an unknown operation returns 400."""
    response = auth_client.post(
        "/calculations",
        json={"operation": "modulo", "a": 5, "b": 3},
    )
    assert response.status_code == 400
    assert "Unknown operation" in response.json()["detail"]


def test_create_calculation_division_by_zero(auth_client: TestClient):
    """POST /calculations with div and b=0 returns 400."""
    response = auth_client.post(
        "/calculations",
        json={"operation": "div", "a": 5, "b": 0},
    )
    assert response.status_code == 400
    assert response.json()["detail"] == "Cannot divide by zero"


def test_list_calculations(auth_client: TestClient):
    """GET /calculations returns a list ordered by created_at descending."""
    auth_client.post("/calculations", json={"operation": "add", "a": 1, "b": 2})
    auth_client.post("/calculations", json={"operation": "sub", "a": 10, "b": 3})

    response = auth_client.get("/calculations")
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    assert len(data) >= 2


def test_get_calculation_by_id(auth_client: TestClient):
    """GET /calculations/{id} returns the calculation."""
    create_resp = auth_client.post(
        "/calculations",
        json={"operation": "mul", "a": 4, "b": 5},
    )
    calc_id = create_resp.json()["id"]

    response = auth_client.get(f"/calculations/{calc_id}")
    assert response.status_code == 200
    data = response.json()
    assert data["id"] == calc_id
    assert data["result"] == 20.0


def test_get_calculation_not_found(auth_client: TestClient):
    """GET /calculations/{id} for a non-existent ID returns 404."""
    response = auth_client.get("/calculations/99999")
    assert response.status_code == 404
    assert response.json()["detail"] == "Calculation not found"


def test_delete_calculation(auth_client: TestClient):
    """DELETE /calculations/{id} removes the calculation and returns 204."""
    create_resp = auth_client.post(
        "/calculations",
        json={"operation": "add", "a": 1, "b": 1},
    )
    calc_id = create_resp.json()["id"]

    response = auth_client.delete(f"/calculations/{calc_id}")
    assert response.status_code == 204

    # Confirm it's gone
    get_resp = auth_client.get(f"/calculations/{calc_id}")
    assert get_resp.status_code == 404


def test_delete_calculation_not_found(auth_client: TestClient):
    """DELETE /calculations/{id} for a non-existent ID returns 404."""
    response = auth_client.delete("/calculations/99999")
    assert response.status_code == 404
    assert response.json()["detail"] == "Calculation not found"
