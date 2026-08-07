from sqlalchemy import Integer, String, Text, Boolean, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class FormField(Base):
    __tablename__ = "form_fields"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    form_id: Mapped[int] = mapped_column(Integer, nullable=False)
    template: Mapped[str] = mapped_column(String(191), nullable=False)
    attribute: Mapped[str] = mapped_column(String(191), nullable=False)
    deskripsi: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    question: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    required: Mapped[Optional[bool]] = mapped_column(Boolean, nullable=True)
    options: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    filled: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    is_verified: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    sequence: Mapped[int] = mapped_column(Integer, nullable=False)
    image_quantity: Mapped[Optional[str]] = mapped_column(String(10), nullable=True)