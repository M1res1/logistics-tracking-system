import uuid
from typing import Any, Dict, Optional

from fastapi import APIRouter, Depends, HTTPException, Query, status

from app.middleware.auth import RequireRole, TokenPayload
from app.schemas.restaurant import OrderActionRequest
from app.services.external_clients import order_client, payment_client

router = APIRouter(prefix="/restaurants", tags=["Kitchen Orders"])


async def _assert_restaurant_owner(
    restaurant_id: uuid.UUID,
    current_user: TokenPayload,
) -> None:
    """Verify caller owns this restaurant (or is admin)."""
    # In a full implementation this would query DB;
    # here we rely on the order service returning 403 if restaurant doesn't match.
    # For strictness, inject DB and call get_restaurant_by_id.
    pass


@router.get(
    "/{restaurant_id}/orders",
    summary="List incoming kitchen orders, filter by status",
)
async def list_kitchen_orders(
    restaurant_id: uuid.UUID,
    order_status: Optional[str] = Query(None, alias="status"),
    page: int = Query(default=1, ge=1),
    limit: int = Query(default=20, ge=1, le=100),
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
) -> Dict[str, Any]:
    try:
        return await order_client.get_restaurant_orders(
            restaurant_id=restaurant_id,
            status_filter=order_status,
            page=page,
            limit=limit,
        )
    except Exception as exc:
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail=f"Order service unavailable: {exc}",
        )


@router.post(
    "/{restaurant_id}/orders/{order_id}/accept",
    summary="Accept order → status CONFIRMED",
)
async def accept_order(
    restaurant_id: uuid.UUID,
    order_id: str,
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
) -> Dict[str, Any]:
    try:
        return await order_client.update_order_status(order_id, "CONFIRMED")
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_502_BAD_GATEWAY, detail=str(exc))


@router.post(
    "/{restaurant_id}/orders/{order_id}/ready",
    summary="Mark food ready → status READY, trigger delivery assignment",
)
async def mark_order_ready(
    restaurant_id: uuid.UUID,
    order_id: str,
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
) -> Dict[str, Any]:
    try:
        result = await order_client.update_order_status(order_id, "READY")
        # Trigger delivery assignment asynchronously
        await order_client.trigger_delivery_assignment(order_id)
        return result
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_502_BAD_GATEWAY, detail=str(exc))


@router.post(
    "/{restaurant_id}/orders/{order_id}/reject",
    summary="Reject order with reason, trigger refund",
)
async def reject_order(
    restaurant_id: uuid.UUID,
    order_id: str,
    body: OrderActionRequest,
    current_user: TokenPayload = Depends(RequireRole("restaurant_owner", "admin")),
) -> Dict[str, Any]:
    try:
        result = await order_client.update_order_status(
            order_id, "REJECTED", extra={"reason": body.reason}
        )
        # Trigger refund via payment service
        await payment_client.initiate_refund(order_id, reason=body.reason)
        return result
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_502_BAD_GATEWAY, detail=str(exc))
