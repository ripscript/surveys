from sqlalchemy.orm import relationship
from sqlalchemy import BigInteger, String, DateTime, ForeignKey, Text
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional, Any

from app.models.base import Base


class Report(Base):
    __tablename__ = "reports"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    name: Mapped[str] = mapped_column(String(191), nullable=False)
    respondent_id: Mapped[int] = mapped_column(BigInteger, nullable=False)
    tingkat_wilayah: Mapped[str] = mapped_column(String(255), nullable=False)
    breakdown_level: Mapped[str] = mapped_column(String(20), nullable=False)
    kecamatan_id: Mapped[Optional[Any]] = mapped_column(JSONB, nullable=True)
    kelurahan_id: Mapped[Optional[Any]] = mapped_column(JSONB, nullable=True)
    rw_id: Mapped[Optional[Any]] = mapped_column(JSONB, nullable=True)
    survey_id: Mapped[Any] = mapped_column(JSONB, nullable=False)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True))
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True))

    cover = relationship("ReportCover", backref="report", uselist=False, lazy="noload")
    sections = relationship("ReportSection", backref="report", order_by="ReportSection.sequence", lazy="noload")