from sqlalchemy.orm import relationship
from sqlalchemy import BigInteger, Boolean, DateTime, ForeignKey, Text
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class ReportSection(Base):
    __tablename__ = "report_sections"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    report_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("reports.id", ondelete="CASCADE"), nullable=False)
    title: Mapped[str] = mapped_column(Text, nullable=False)
    has_sub_section: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    sequence: Mapped[int] = mapped_column(BigInteger, nullable=False)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)

    sub_sections = relationship("ReportSubsection", backref="section", order_by="ReportSubsection.sequence", lazy="noload")
    components = relationship("ReportComponent", backref="section", order_by="ReportComponent.sequence", lazy="noload")