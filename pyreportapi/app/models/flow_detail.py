from sqlalchemy import Integer, BigInteger, String, Text, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class FlowDetail(Base):
    __tablename__ = "flow_details"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    form_id: Mapped[int] = mapped_column(Integer, nullable=False)
    name: Mapped[str] = mapped_column(String(191), nullable=False)
    description: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    code: Mapped[str] = mapped_column(String(191), nullable=False)
    version: Mapped[int] = mapped_column(Integer, nullable=False)
    status_section: Mapped[str] = mapped_column(String(1), nullable=False)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    opening_id: Mapped[int] = mapped_column(Integer, nullable=False)
    closing_id: Mapped[int] = mapped_column(Integer, nullable=False)
    created_by: Mapped[int] = mapped_column(BigInteger, nullable=False)