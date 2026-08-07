from sqlalchemy import Integer, BigInteger, String, Text, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class FieldResponse(Base):
    __tablename__ = "field_responses"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    form_field_id: Mapped[int] = mapped_column(Integer, nullable=False)
    form_response_id: Mapped[int] = mapped_column(Integer, nullable=False)
    answer: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    media: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    location_address: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    group_id: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)