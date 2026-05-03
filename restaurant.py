import json
import uuid
from datetime import datetime, time
from typing import List

from sqlalchemy import (
    Boolean, Column, DateTime, Float, ForeignKey,
    Integer, String, Time, Text, func, TypeDecorator, types
)
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship, Mapped

from app.db.session import Base


class StringList(TypeDecorator):
    """
    Portable list-of-strings column.
    - PostgreSQL: native ARRAY(String)
    - SQLite / others: JSON text
    """
    impl = types.Text
    cache_ok = True

    def load_dialect_impl(self, dialect):
        if dialect.name == "postgresql":
            from sqlalchemy.dialects.postgresql import ARRAY
            return dialect.type_descriptor(ARRAY(String))
        return dialect.type_descriptor(types.Text())

    def process_bind_param(self, value, dialect):
        if dialect.name == "postgresql":
            return value  # ARRAY handles it natively
        if value is None:
            return "[]"
        return json.dumps(value)

    def process_result_value(self, value, dialect):
        if dialect.name == "postgresql":
            return value  # ARRAY returns a list already
        if value is None:
            return []
        if isinstance(value, list):
            return value
        return json.loads(value)


class Restaurant(Base):
    __tablename__ = "restaurants"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4, index=True)
    owner_id = Column(UUID(as_uuid=True), nullable=False, index=True)

    name = Column(String(255), nullable=False)
    address = Column(String(500), nullable=False)
    lat = Column(Float, nullable=False)
    lng = Column(Float, nullable=False)

    # e.g. ["italian", "pizza", "pasta"]
    cuisine_types = Column(StringList, nullable=False, default=list)

    rating = Column(Float, default=0.0)
    rating_count = Column(Integer, default=0)

    is_active = Column(Boolean, default=True, nullable=False)

    opening_time = Column(Time, nullable=False)   # e.g. 09:00
    closing_time = Column(Time, nullable=False)   # e.g. 22:00

    phone = Column(String(50), nullable=True)
    description = Column(Text, nullable=True)

    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    menu_items: Mapped[List["MenuItem"]] = relationship(
        "MenuItem", back_populates="restaurant", cascade="all, delete-orphan"
    )


class MenuItem(Base):
    __tablename__ = "menu_items"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4, index=True)
    restaurant_id = Column(
        UUID(as_uuid=True),
        ForeignKey("restaurants.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )

    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    price = Column(Float, nullable=False)
    category = Column(String(100), nullable=False)

    is_available = Column(Boolean, default=True, nullable=False)
    prep_time_minutes = Column(Integer, default=15, nullable=False)

    image_url = Column(String(500), nullable=True)

    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    restaurant: Mapped["Restaurant"] = relationship("Restaurant", back_populates="menu_items")
