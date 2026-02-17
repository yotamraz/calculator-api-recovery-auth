"""Tests for calculations CRUD endpoints."""

from fastapi.testclient import TestClient


def test_create_calculation(client: TestClient, auth_headers: dict):
    """POST /calculations with a valid operation returns 201."""
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


def test_create_calculation_unknown_operation(client: TestClient, auth_headers: dict):
    """POST /calculations with unknown operation returns 400."""
    response = client.post(
        "/calculations",
        json={"operation": "modulo", "a": 5, "b": 3},
        headers=auth_headers,
    )
    assert response.status_code == 400
    assert "Unknown operation" in response.json()["detail"]


def test_create_calculation_division_by_zero(client: TestClient, auth_headers: dict):
    """POST /calculations with div and b=0 returns 400."""
    response = client.post(
        "/calculations",
        json={"operation": "div", "a": 1, "b": 0},
        headers=auth_headers,
    )
    assert response.status_code == 400
    assert response.json()["detail"] == "Cannot divide by zero"


def test_list_calculations(client: TestClient, auth_headers: dict):
    """GET /calculations returns created items ordered by created_at desc."""
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
    assert len(data) >= 2
    # Most recently created should be first (desc order)
    assert data[0]["operation"] == "mul"
    assert data[1]["operation"] == "add"


def test_get_calculation_by_id(client: TestClient, auth_headers: dict):
    """GET /calculations/{id} returns the calculation."""
    create_resp = client.post(
        "/calculations",
        json={"operation": "sub", "a": 10, "b": 3},
        headers=auth_headers,
    )
    calc_id = create_resp.json()["id"]
    response = client.get(f"/calculations/{calc_id}", headers=auth_headers)
    assert response.status_code == 200
    data = response.json()
    assert data["id"] == calc_id
    assert data["operation"] == "sub"
    assert data["result"] == 7.0


def test_get_calculation_not_found(client: TestClient, auth_headers: dict):
    """GET /calculations/{id} with nonexistent ID returns 404."""
    response = client.get("/calculations/99999", headers=auth_headers)
    assert response.status_code == 404
    assert response.json()["detail"] == "Calculation not found"


def test_delete_calculation(client: TestClient, auth_headers: dict):
    """DELETE /calculations/{id} returns 204."""
    create_resp = client.post(
        "/calculations",
        json={"operation": "mul", "a": 2, "b": 5},
        headers=auth_headers,
    )
    calc_id = create_resp.json()["id"]
    response = client.delete(f"/calculations/{calc_id}", headers=auth_headers)
    assert response.status_code == 204
    # Verify it's actually deleted
    get_resp = client.get(f"/calculations/{calc_id}", headers=auth_headers)
    assert get_resp.status_code == 404


def test_delete_calculation_not_found(client: TestClient, auth_headers: dict):
    """DELETE /calculations/{id} with nonexistent ID returns 404."""
    response = client.delete("/calculations/99999", headers=auth_headers)
    assert response.status_code == 404
    assert response.json()["detail"] == "Calculation not found"


def test_create_calculation_requires_auth(client: TestClient):
    """POST /calculations without token returns 401."""
    response = client.post(
        "/calculations",
        json={"operation": "add", "a": 1, "b": 2},
    )
    assert response.status_code == 401


def test_list_calculations_requires_auth(client: TestClient):
    """GET /calculations without token returns 401."""
    response = client.get("/calculations")
    assert response.status_code == 401


def test_get_calculation_requires_auth(client: TestClient):
    """GET /calculations/{id} without token returns 401."""
    response = client.get("/calculations/1")
    assert response.status_code == 401


def test_delete_calculation_requires_auth(client: TestClient):
    """DELETE /calculations/{id} without token returns 401."""
    response = client.delete("/calculations/1")
    assert response.status_code == 401
