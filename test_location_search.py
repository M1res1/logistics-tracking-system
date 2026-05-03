"""
Tests for location-based restaurant search.
"""
import uuid
from datetime import datetime

import pytest
from httpx import AsyncClient

from tests.conftest import auth_headers, restaurant_payload
from app.services.restaurant_service import haversine_km

pytestmark = pytest.mark.asyncio


# ──────────────────────── Unit: haversine ────────────────────────────────

def test_haversine_same_point():
    assert haversine_km(41.2995, 69.2401, 41.2995, 69.2401) == pytest.approx(0.0, abs=0.001)


def test_haversine_known_distance():
    # Tashkent → Samarkand ≈ 270 km
    dist = haversine_km(41.2995, 69.2401, 39.6270, 66.9750)
    assert 260 < dist < 280


def test_haversine_short_distance():
    # ~1.57 km apart
    dist = haversine_km(41.2995, 69.2401, 41.3136, 69.2401)
    assert 1.5 < dist < 1.7


# ──────────────────────── Integration: radius filter ─────────────────────

async def _create_restaurant_at(client: AsyncClient, lat: float, lng: float, name: str) -> str:
    owner_id = str(uuid.uuid4())
    resp = await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(lat=lat, lng=lng, name=name),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    assert resp.status_code == 201, resp.text
    return resp.json()["id"]


async def test_location_search_returns_within_radius(client: AsyncClient):
    """
    Create restaurants at different distances from search centre.
    Only the one within radius_km should be returned.
    """
    # Centre point: 41.2995, 69.2401 (Tashkent)
    # Restaurant A: ~0.5 km away → should appear
    # Restaurant B: ~30 km away → should NOT appear in 5 km radius
    rid_near = await _create_restaurant_at(client, 41.2995, 69.2401, "Near Restaurant")
    rid_far  = await _create_restaurant_at(client, 41.5500, 69.2401, "Far Restaurant")  # ~28 km

    resp = await client.get(
        "/api/v1/restaurants",
        params={"lat": 41.2995, "lng": 69.2401, "radius_km": 5.0},
    )
    assert resp.status_code == 200
    ids = {r["id"] for r in resp.json()}

    assert rid_near in ids, "Near restaurant should be in results"
    assert rid_far not in ids, "Far restaurant should NOT be in 5 km radius"


async def test_location_search_wider_radius_includes_more(client: AsyncClient):
    rid_near = await _create_restaurant_at(client, 41.2995, 69.2401, "Near")
    rid_far  = await _create_restaurant_at(client, 41.5500, 69.2401, "Far")

    resp = await client.get(
        "/api/v1/restaurants",
        params={"lat": 41.2995, "lng": 69.2401, "radius_km": 50.0},
    )
    ids = {r["id"] for r in resp.json()}
    assert rid_near in ids
    assert rid_far in ids


async def test_cuisine_type_filter(client: AsyncClient):
    owner_id = str(uuid.uuid4())
    await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(name="Sushi Place", cuisine_types=["sushi", "japanese"]),
        headers=auth_headers(owner_id, "restaurant_owner"),
    )
    await client.post(
        "/api/v1/restaurants",
        json=restaurant_payload(name="Burger Joint", cuisine_types=["burgers", "american"]),
        headers=auth_headers(str(uuid.uuid4()), "restaurant_owner"),
    )

    resp = await client.get("/api/v1/restaurants", params={"cuisine_type": "sushi"})
    assert resp.status_code == 200
    names = [r["name"] for r in resp.json()]
    assert "Sushi Place" in names
    assert "Burger Joint" not in names


async def test_no_location_params_returns_all(client: AsyncClient):
    for i in range(3):
        await _create_restaurant_at(
            client, 41.2995 + i * 0.1, 69.2401, f"Restaurant {i}"
        )

    resp = await client.get("/api/v1/restaurants")
    assert resp.status_code == 200
    assert len(resp.json()) == 3
