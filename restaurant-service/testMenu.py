"""
Tests for menu item management + ownership enforcement.
"""
import uuid

import pytest
from httpx import AsyncClient

from tests.conftest import auth_headers, restaurant_payload

pytestmark = pytest.mark.asyncio


async def _create_restaurant(client: AsyncClient, owner_id: str) -> str:
    resp = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    assert resp.status_code == 201
    return resp.json()["id"]


def menu_item_payload(**overrides) -> dict:
    base = {
        "name": "Margherita Pizza",
        "description": "Classic tomato and mozzarella",
        "price": 12.99,
        "category": "Pizza",
        "prep_time_minutes": 20,
    }
    base.update(overrides)
    return base


# ──────────────────────── Add menu item ───────────────────────────────────

async def test_owner_can_add_menu_item(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)

    resp = await client.post(
        f"/api/v1/restaurants/{rid}/menu-items",
        json=menu_item_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == "Margherita Pizza"
    assert data["price"] == 12.99
    assert data["is_available"] is True
    assert data["restaurant_id"] == rid


async def test_non_owner_cannot_add_menu_item(client: AsyncClient):
    """
    CRITICAL: non-owner restaurant_owner cannot add item to another restaurant's menu.
    """
    owner_id = str(uuid.uuid4())
    intruder_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)

    resp = await client.post(
        f"/api/v1/restaurants/{rid}/menu-items",
        json=menu_item_payload(),
        headers=auth_headers(intruder_id, "restaurant_owner"),
    )
    assert resp.status_code == 403


async def test_customer_cannot_add_menu_item(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)

    resp = await client.post(
        f"/api/v1/restaurants/{rid}/menu-items",
        json=menu_item_payload(),
        headers=auth_headers(str(uuid.uuid4()), "customer"),
    )
    assert resp.status_code == 403


# ──────────────────────── Get menu ────────────────────────────────────────

async def test_get_menu_grouped_by_category(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)
    headers = auth_headers(owner_id, "restaurant_owner")

    # Add items in two categories
    await client.post(f"/api/v1/restaurants/{rid}/menu-items",
                      json=menu_item_payload(name="Pepperoni", category="Pizza"), headers=headers)
    await client.post(f"/api/v1/restaurants/{rid}/menu-items",
                      json=menu_item_payload(name="Tiramisu", category="Dessert", price=6.0), headers=headers)
    await client.post(f"/api/v1/restaurants/{rid}/menu-items",
                      json=menu_item_payload(name="Margherita", category="Pizza"), headers=headers)

    resp = await client.get(f"/api/v1/restaurants/{rid}/menu")
    assert resp.status_code == 200
    categories = {group["category"]: group["items"] for group in resp.json()}
    assert "Pizza" in categories
    assert "Dessert" in categories
    assert len(categories["Pizza"]) == 2
    assert len(categories["Dessert"]) == 1


async def test_unavailable_items_excluded_from_menu(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)
    headers = auth_headers(owner_id, "restaurant_owner")

    add_resp = await client.post(
        f"/api/v1/restaurants/{rid}/menu-items",
        json=menu_item_payload(name="Hidden Item"),
        headers=headers,
    )
    item_id = add_resp.json()["id"]

    # Soft delete (is_available = False)
    await client.delete(
        f"/api/v1/restaurants/{rid}/menu-items/{item_id}",
        headers=headers,
    )

    menu_resp = await client.get(f"/api/v1/restaurants/{rid}/menu")
    all_items = [item for group in menu_resp.json() for item in group["items"]]
    assert not any(i["id"] == item_id for i in all_items)


# ──────────────────────── Update menu item ────────────────────────────────

async def test_owner_can_update_menu_item(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)
    headers = auth_headers(owner_id, "restaurant_owner")

    add_resp = await client.post(f"/api/v1/restaurants/{rid}/menu-items",
                                  json=menu_item_payload(), headers=headers)
    item_id = add_resp.json()["id"]

    update_resp = await client.put(
        f"/api/v1/restaurants/{rid}/menu-items/{item_id}",
        json={"price": 15.99, "description": "New description"},
        headers=headers,
    )
    assert update_resp.status_code == 200
    assert update_resp.json()["price"] == 15.99
    assert update_resp.json()["description"] == "New description"


async def test_non_owner_cannot_update_another_restaurants_menu(client: AsyncClient):
    """
    CRITICAL: non-owner must be blocked from modifying menu items.
    """
    owner_id = str(uuid.uuid4())
    intruder_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)

    add_resp = await client.post(
        f"/api/v1/restaurants/{rid}/menu-items",
        json=menu_item_payload(),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    item_id = add_resp.json()["id"]

    update_resp = await client.put(
        f"/api/v1/restaurants/{rid}/menu-items/{item_id}",
        json={"price": 1.00},
        headers=auth_headers(intruder_id, "restaurant_owner"),
    )
    assert update_resp.status_code == 403


# ──────────────────────── Soft delete ────────────────────────────────────

async def test_soft_delete_sets_is_available_false(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    rid = await _create_restaurant(client, owner_id)
    headers = auth_headers(owner_id, "restaurant_owner")

    add_resp = await client.post(f"/api/v1/restaurants/{rid}/menu-items",
                                  json=menu_item_payload(), headers=headers)
    item_id = add_resp.json()["id"]

    del_resp = await client.delete(
        f"/api/v1/restaurants/{rid}/menu-items/{item_id}",
        headers=headers,
    )
    assert del_resp.status_code == 200
    assert del_resp.json()["is_available"] is False
