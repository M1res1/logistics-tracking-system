import uuid
from typing import List

from fastapi import APIRouter, Depends, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.session import get_db
from app.middleware.auth import RequireRole, TokenPayload
from app.schemas.restaurant import MenuItemCreate, MenuItemUpdate, MenuItemResponse, MenuByCategory
from app.services import restaurant_service as svc

router = APIRouter(prefix="/restaurants", tags=["Menu"])


@router.get(
    "/{restaurant_id}/menu",
    response_model=List[MenuByCategory],
    summary="Get full menu grouped by category",
)
async def get_menu(
    restaurant_id: uuid.UUID,
    db: AsyncSession = Depends(get_db),
):
    return await svc.get_menu_by_category(db, restaurant_id)


@router.post(
    "/{restaurant_id}/menu-items",
    response_model=MenuItemResponse,
    status_code=status.HTTP_201_CREATED,
    summary="Add menu item (owner only)",
)
async def add_menu_item(
    restaurant_id: uuid.UUID,
    data: MenuItemCreate,
    db: AsyncSession = Depends(get_db),
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
):
    return await svc.add_menu_item(
        db,
        restaurant_id=restaurant_id,
        owner_id=uuid.UUID(current_user.sub),
        data=data,
        is_admin=(current_user.role == "admin"),
    )


@router.put(
    "/{restaurant_id}/menu-items/{item_id}",
    response_model=MenuItemResponse,
    summary="Update menu item (owner only)",
)
async def update_menu_item(
    restaurant_id: uuid.UUID,
    item_id: uuid.UUID,
    data: MenuItemUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
):
    return await svc.update_menu_item(
        db,
        restaurant_id=restaurant_id,
        item_id=item_id,
        owner_id=uuid.UUID(current_user.sub),
        data=data,
        is_admin=(current_user.role == "admin"),
    )


@router.delete(
    "/{restaurant_id}/menu-items/{item_id}",
    response_model=MenuItemResponse,
    summary="Soft-delete menu item (owner only)",
)
async def delete_menu_item(
    restaurant_id: uuid.UUID,
    item_id: uuid.UUID,
    db: AsyncSession = Depends(get_db),
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
):
    return await svc.soft_delete_menu_item(
        db,
        restaurant_id=restaurant_id,
        item_id=item_id,
        owner_id=uuid.UUID(current_user.sub),
        is_admin=(current_user.role == "admin"),
    )
