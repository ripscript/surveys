from sqlalchemy import BigInteger, String, DateTime, ForeignKey, Text
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional, Any

from app.models.base import Base


class ReportCover(Base):
    __tablename__ = "report_covers"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    report_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("reports.id", ondelete="CASCADE"), nullable=False)
    img_depan: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    img_belakang: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    page: Mapped[Optional[Any]] = mapped_column(JSONB, nullable=True)
    text_depan: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    text_belakang: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    kata_pengantar: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)