"""
Tests for restaurant management endpoints.
"""
import uuid

import pytest
from httpx import AsyncClient

from tests.conftest import auth_headers, restaurant_payload

pytestmark = pytest.mark.asyncio


# ──────────────────────── Create restaurant ───────────────────────────────

async def test_create_restaurant_as_owner_succeeds(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    response = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    assert response.status_code == 201
    data = response.json()
    assert data["name"] == "Test Pizza"
    assert data["owner_id"] == owner_id
    assert data["is_active"] is True
    assert "pizza" in data["cuisine_types"]


async def test_create_restaurant_as_customer_forbidden(client: AsyncClient):
    response = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(str(uuid.uuid4()), "customer"),
    )
    assert response.status_code == 403


async def test_create_restaurant_unauthenticated(client: AsyncClient):
    response = await client.post("/api/v1/restaurants", json=restaurant_payload())
    assert response.status_code == 403


# ──────────────────────── Get restaurant ─────────────────────────────────

async def test_get_restaurant_returns_full_details(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    create_resp = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    rid = create_resp.json()["id"]

    response = await client.get(f"/api/v1/restaurants/{rid}")
    assert response.status_code == 200
    data = response.json()
    assert data["id"] == rid
    assert "menu_items" in data


async def test_get_nonexistent_restaurant_404(client: AsyncClient):
    response = await client.get(f"/api/v1/restaurants/{uuid.uuid4()}")
    assert response.status_code == 404


# ──────────────────────── Update restaurant ───────────────────────────────

async def test_owner_can_update_own_restaurant(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    create_resp = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    rid = create_resp.json()["id"]

    update_resp = await client.put(
        f"/api/v1/restaurants/{rid}",
        json={"name": "Updated Name"},
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    assert update_resp.status_code == 200
    assert update_resp.json()["name"] == "Updated Name"


async def test_non_owner_cannot_update_restaurant(client: AsyncClient):
    """
    CRITICAL: A different restaurant_owner must not be able to modify
    another owner's restaurant.
    """
    owner_id = str(uuid.uuid4())
    intruder_id = str(uuid.uuid4())

    create_resp = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    rid = create_resp.json()["id"]

    update_resp = await client.put(
        f"/api/v1/restaurants/{rid}",
        json={"name": "Hacked Name"},
        headers=auth_headers(intruder_id, "restaurant_owner"),
    )
    assert update_resp.status_code == 403


async def test_admin_can_update_any_restaurant(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    create_resp = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    rid = create_resp.json()["id"]

    admin_resp = await client.put(
        f"/api/v1/restaurants/{rid}",
        json={"name": "Admin Updated"},
        headers=auth_headers(str(uuid.uuid4()), "admin"),
    )
    assert admin_resp.status_code == 200
    assert admin_resp.json()["name"] == "Admin Updated"


# ──────────────────────── Toggle restaurant ───────────────────────────────

async def test_toggle_restaurant_flips_is_active(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    create_resp = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    rid = create_resp.json()["id"]
    assert create_resp.json()["is_active"] is True

    toggle_resp = await client.put(
        f"/api/v1/restaurants/{rid}/toggle",
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    assert toggle_resp.status_code == 200
    assert toggle_resp.json()["is_active"] is False

    # Toggle back
    toggle_resp2 = await client.put(
        f"/api/v1/restaurants/{rid}/toggle",
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    assert toggle_resp2.json()["is_active"] is True
