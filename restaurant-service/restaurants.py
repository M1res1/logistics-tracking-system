import uuid
from typing import List, Optional

from fastapi import APIRouter, Depends, Query, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.session import get_db
from app.middleware.auth import RequireRole, get_current_user, TokenPayload
from app.schemas.restaurant import (
    RestaurantCreate, RestaurantUpdate, RestaurantResponse,
    RestaurantWithMenu, RestaurantFilterParams,
)
from app.services import restaurant_service as svc

router = APIRouter(prefix="/restaurants", tags=["Restaurants"])


@router.post(
    "",
    response_model=RestaurantResponse,
    status_code=status.HTTP_201_CREATED,
    summary="Create restaurant (owner only)",
)
async def create_restaurant(
    data: RestaurantCreate,
    db: AsyncSession = Depends(get_db),
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
):
    restaurant = await svc.create_restaurant(
        db, owner_id=uuid.UUID(current_user.sub), data=data
    )
    return restaurant


@router.get(
    "",
    response_model=List[RestaurantResponse],
    summary="List restaurants with optional filters",
)
async def list_restaurants(
    lat: Optional[float] = Query(None, ge=-90, le=90),
    lng: Optional[float] = Query(None, ge=-180, le=180),
    radius_km: float = Query(default=5.0, ge=0.1, le=50),
    cuisine_type: Optional[str] = Query(None),
    open_now: bool = Query(default=False),
    page: int = Query(default=1, ge=1),
    limit: int = Query(default=20, ge=1, le=100),
    db: AsyncSession = Depends(get_db),
):
    filters = RestaurantFilterParams(
        lat=lat, lng=lng, radius_km=radius_km,
        cuisine_type=cuisine_type, open_now=open_now,
        page=page, limit=limit,
    )
    restaurants, _ = await svc.list_restaurants(db, filters)
    return restaurants


@router.get(
    "/{restaurant_id}",
    response_model=RestaurantWithMenu,
    summary="Get restaurant with full details and menu",
)
async def get_restaurant(
    restaurant_id: uuid.UUID,
    db: AsyncSession = Depends(get_db),
):
    return await svc.get_restaurant_by_id(db, restaurant_id, with_menu=True)


@router.put(
    "/{restaurant_id}",
    response_model=RestaurantResponse,
    summary="Update restaurant details (owner only)",
)
async def update_restaurant(
    restaurant_id: uuid.UUID,
    data: RestaurantUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
):
    return await svc.update_restaurant(
        db,
        restaurant_id=restaurant_id,
        owner_id=uuid.UUID(current_user.sub),
        data=data,
        is_admin=(current_user.role == "admin"),
    )


@router.put(
    "/{restaurant_id}/toggle",
    response_model=RestaurantResponse,
    summary="Toggle restaurant open/closed",
)
async def toggle_restaurant(
    restaurant_id: uuid.UUID,
    db: AsyncSession = Depends(get_db),
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
):
    return await svc.toggle_restaurant(
        db,
        restaurant_id=restaurant_id,
        owner_id=uuid.UUID(current_user.sub),
        is_admin=(current_user.role == "admin"),
    )
