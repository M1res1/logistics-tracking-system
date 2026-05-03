import math
import uuid
from datetime import datetime, time
from typing import List, Optional, Tuple

from fastapi import HTTPException, status
from sqlalchemy import select, and_, or_, func, cast
from sqlalchemy.dialects.postgresql import ARRAY
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.models.restaurant import Restaurant, MenuItem
from app.schemas.restaurant import (
    RestaurantCreate, RestaurantUpdate,
    MenuItemCreate, MenuItemUpdate,
    RestaurantFilterParams, MenuByCategory,
)


# ─────────────────────────── Haversine helper ────────────────────────────

def haversine_km(lat1: float, lng1: float, lat2: float, lng2: float) -> float:
    """Calculate distance in km between two lat/lng points."""
    R = 6371.0
    φ1, φ2 = math.radians(lat1), math.radians(lat2)
    Δφ = math.radians(lat2 - lat1)
    Δλ = math.radians(lng2 - lng1)
    a = math.sin(Δφ / 2) ** 2 + math.cos(φ1) * math.cos(φ2) * math.sin(Δλ / 2) ** 2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))


def _is_open_now(restaurant: Restaurant) -> bool:
    now = datetime.now().time()
    o, c = restaurant.opening_time, restaurant.closing_time
    if o <= c:
        return o <= now <= c
    # overnight (e.g. 22:00 – 02:00)
    return now >= o or now <= c


# ──────────────────────────── Restaurant CRUD ────────────────────────────

async def create_restaurant(
    db: AsyncSession, owner_id: uuid.UUID, data: RestaurantCreate
) -> Restaurant:
    restaurant = Restaurant(
        owner_id=owner_id,
        **data.model_dump(),
    )
    db.add(restaurant)
    await db.flush()
    await db.refresh(restaurant)
    return restaurant


async def get_restaurant_by_id(
    db: AsyncSession, restaurant_id: uuid.UUID, with_menu: bool = False
) -> Restaurant:
    query = select(Restaurant).where(Restaurant.id == restaurant_id)
    if with_menu:
        query = query.options(selectinload(Restaurant.menu_items))
    result = await db.execute(query)
    restaurant = result.scalar_one_or_none()
    if not restaurant:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Restaurant not found")
    return restaurant


async def list_restaurants(
    db: AsyncSession, filters: RestaurantFilterParams
) -> Tuple[List[Restaurant], int]:
    query = select(Restaurant)

    # Filter by cuisine type (applied in Python for cross-DB compat)
    result = await db.execute(query)
    restaurants = list(result.scalars().all())

    if filters.cuisine_type:
        cuisine = filters.cuisine_type.strip().lower()
        restaurants = [r for r in restaurants if cuisine in (r.cuisine_types or [])]

    # Apply haversine in Python (for small datasets; use PostGIS for large scale)
    if filters.lat is not None and filters.lng is not None:
        restaurants = [
            r for r in restaurants
            if haversine_km(filters.lat, filters.lng, r.lat, r.lng) <= filters.radius_km
        ]

    if filters.open_now:
        restaurants = [r for r in restaurants if _is_open_now(r)]

    total = len(restaurants)
    offset = (filters.page - 1) * filters.limit
    return restaurants[offset: offset + filters.limit], total


async def update_restaurant(
    db: AsyncSession,
    restaurant_id: uuid.UUID,
    owner_id: uuid.UUID,
    data: RestaurantUpdate,
    is_admin: bool = False,
) -> Restaurant:
    restaurant = await get_restaurant_by_id(db, restaurant_id)
    _assert_owner(restaurant, owner_id, is_admin)

    for field, value in data.model_dump(exclude_unset=True).items():
        setattr(restaurant, field, value)

    await db.flush()
    await db.refresh(restaurant)
    return restaurant


async def toggle_restaurant(
    db: AsyncSession,
    restaurant_id: uuid.UUID,
    owner_id: uuid.UUID,
    is_admin: bool = False,
) -> Restaurant:
    restaurant = await get_restaurant_by_id(db, restaurant_id)
    _assert_owner(restaurant, owner_id, is_admin)
    restaurant.is_active = not restaurant.is_active
    await db.flush()
    await db.refresh(restaurant)
    return restaurant


# ───────────────────────────── Menu CRUD ─────────────────────────────────

async def get_menu_by_category(
    db: AsyncSession, restaurant_id: uuid.UUID
) -> List[MenuByCategory]:
    result = await db.execute(
        select(MenuItem)
        .where(
            and_(
                MenuItem.restaurant_id == restaurant_id,
                MenuItem.is_available == True,
            )
        )
        .order_by(MenuItem.category, MenuItem.name)
    )
    items = list(result.scalars().all())

    grouped: dict[str, list] = {}
    for item in items:
        grouped.setdefault(item.category, []).append(item)

    return [MenuByCategory(category=cat, items=items) for cat, items in grouped.items()]


async def add_menu_item(
    db: AsyncSession,
    restaurant_id: uuid.UUID,
    owner_id: uuid.UUID,
    data: MenuItemCreate,
    is_admin: bool = False,
) -> MenuItem:
    restaurant = await get_restaurant_by_id(db, restaurant_id)
    _assert_owner(restaurant, owner_id, is_admin)

    item = MenuItem(restaurant_id=restaurant_id, **data.model_dump())
    db.add(item)
    await db.flush()
    await db.refresh(item)
    return item


async def update_menu_item(
    db: AsyncSession,
    restaurant_id: uuid.UUID,
    item_id: uuid.UUID,
    owner_id: uuid.UUID,
    data: MenuItemUpdate,
    is_admin: bool = False,
) -> MenuItem:
    restaurant = await get_restaurant_by_id(db, restaurant_id)
    _assert_owner(restaurant, owner_id, is_admin)

    item = await _get_menu_item(db, restaurant_id, item_id)

    for field, value in data.model_dump(exclude_unset=True).items():
        setattr(item, field, value)

    await db.flush()
    await db.refresh(item)
    return item


async def soft_delete_menu_item(
    db: AsyncSession,
    restaurant_id: uuid.UUID,
    item_id: uuid.UUID,
    owner_id: uuid.UUID,
    is_admin: bool = False,
) -> MenuItem:
    restaurant = await get_restaurant_by_id(db, restaurant_id)
    _assert_owner(restaurant, owner_id, is_admin)

    item = await _get_menu_item(db, restaurant_id, item_id)
    item.is_available = False
    await db.flush()
    await db.refresh(item)
    return item


# ─────────────────────────── Private helpers ─────────────────────────────

def _assert_owner(restaurant: Restaurant, user_id: uuid.UUID, is_admin: bool) -> None:
    if not is_admin and restaurant.owner_id != user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You do not own this restaurant",
        )


async def _get_menu_item(
    db: AsyncSession, restaurant_id: uuid.UUID, item_id: uuid.UUID
) -> MenuItem:
    result = await db.execute(
        select(MenuItem).where(
            and_(MenuItem.id == item_id, MenuItem.restaurant_id == restaurant_id)
        )
    )
    item = result.scalar_one_or_none()
    if not item:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Menu item not found")
    return item
