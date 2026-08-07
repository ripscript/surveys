from sqlalchemy.orm import relationship
from sqlalchemy import BigInteger, String, DateTime, ForeignKey, Text
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class ReportSubsection(Base):
    __tablename__ = "report_subsections"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    report_section_id: Mapped[int] = mapped_column(BigInteger, ForeignKey("report_sections.id", ondelete="CASCADE"), nullable=False)
    title: Mapped[str] = mapped_column(Text, nullable=False)
    sequence: Mapped[int] = mapped_column(BigInteger, nullable=False)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)

    components = relationship("ReportComponent", backref="sub_section", order_by="ReportComponent.sequence", lazy="noload")