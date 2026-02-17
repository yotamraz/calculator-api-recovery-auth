"""Tests for calculation history CRUD endpoints."""

from fastapi.testclient import TestClient


def test_create_calculation(client: TestClient, auth_headers: dict[str, str]):
    response = client.post(
        "/calculations",
        json={"operation": "add", "a": 3, "b": 4},
        headers=auth_headers,
    )
    assert response.status_code == 201
    data = response.json()
    assert data["operation"] == "add"
    assert data["a"] == 3.0
    assert data["b"] == 4.0
    assert data["result"] == 7.0
    assert "id" in data
    assert "created_at" in data


def test_create_unknown_operation(client: TestClient, auth_headers: dict[str, str]):
    response = client.post(
        "/calculations",
        json={"operation": "mod", "a": 10, "b": 3},
        headers=auth_headers,
    )
    assert response.status_code == 400
    assert "Unknown operation: mod" in response.json()["detail"]


def test_create_division_by_zero(client: TestClient, auth_headers: dict[str, str]):
    response = client.post(
        "/calculations",
        json={"operation": "div", "a": 10, "b": 0},
        headers=auth_headers,
    )
    assert response.status_code == 400
    assert response.json()["detail"] == "Cannot divide by zero"


def test_list_calculations(client: TestClient, auth_headers: dict[str, str]):
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


def test_get_calculation_by_id(client: TestClient, auth_headers: dict[str, str]):
    create_resp = client.post(
        "/calculations",
        json={"operation": "sub", "a": 10, "b": 3},
        headers=auth_headers,
    )
    calc_id = create_resp.json()["id"]
    response = client.get(f"/calculations/{calc_id}", headers=auth_headers)
    assert response.status_code == 200
    data = response.json()
    assert data["operation"] == "sub"
    assert data["a"] == 10.0
    assert data["b"] == 3.0
    assert data["result"] == 7.0


def test_get_missing_calculation(client: TestClient, auth_headers: dict[str, str]):
    response = client.get("/calculations/9999", headers=auth_headers)
    assert response.status_code == 404
    assert response.json()["detail"] == "Calculation not found"


def test_delete_calculation(client: TestClient, auth_headers: dict[str, str]):
    create_resp = client.post(
        "/calculations",
        json={"operation": "mul", "a": 5, "b": 6},
        headers=auth_headers,
    )
    calc_id = create_resp.json()["id"]
    response = client.delete(f"/calculations/{calc_id}", headers=auth_headers)
    assert response.status_code == 204

    # Verify it's gone
    get_resp = client.get(f"/calculations/{calc_id}", headers=auth_headers)
    assert get_resp.status_code == 404


def test_delete_missing_calculation(client: TestClient, auth_headers: dict[str, str]):
    response = client.delete("/calculations/9999", headers=auth_headers)
    assert response.status_code == 404
    assert response.json()["detail"] == "Calculation not found"
