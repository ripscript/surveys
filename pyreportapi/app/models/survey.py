from sqlalchemy import Integer, BigInteger, String, Text, DateTime, Boolean
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class Survey(Base):
    __tablename__ = "surveys"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    name: Mapped[str] = mapped_column(Text, nullable=False)
    flow_detail_id: Mapped[Optional[int]] = mapped_column(BigInteger, nullable=True)
    start_date: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    end_date: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    type: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    status: Mapped[str] = mapped_column(String(191), nullable=False)
    created_by: Mapped[int] = mapped_column(BigInteger, nullable=False)
    deskripsi: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    approval_survey: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    alasan_reject: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    is_repeated: Mapped[Optional[bool]] = mapped_column(Boolean, nullable=True, default=False)