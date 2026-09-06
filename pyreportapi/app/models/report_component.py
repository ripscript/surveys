from sqlalchemy import BigInteger, String, Text, Boolean, DateTime, ForeignKey
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional, Any

from app.models.base import Base


class ReportComponent(Base):
    __tablename__ = "report_components"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)

    report_section_id: Mapped[Optional[int]] = mapped_column(BigInteger, ForeignKey("report_sections.id", ondelete="CASCADE"), nullable=True)
    report_sub_section_id: Mapped[Optional[int]] = mapped_column(BigInteger, ForeignKey("report_subsections.id", ondelete="CASCADE"), nullable=True)

    type: Mapped[str] = mapped_column(String(50), nullable=False)
    sequence: Mapped[int] = mapped_column(BigInteger, nullable=False)
    title: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    form_field_ids: Mapped[Optional[Any]] = mapped_column(JSONB, nullable=True)
    is_multiple_data: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    chart_type: Mapped[Optional[str]] = mapped_column(String(50), nullable=True)
    map_type: Mapped[Optional[str]] = mapped_column(String(20), nullable=True)
    chart_direction: Mapped[Optional[str]] = mapped_column(String(20), nullable=True)
    table_style: Mapped[Optional[str]] = mapped_column(String(50), nullable=True)
    table_config: Mapped[Optional[Any]] = mapped_column(JSONB, nullable=True)
    narrative_template: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    narrative_logic: Mapped[Optional[Any]] = mapped_column(JSONB, nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)