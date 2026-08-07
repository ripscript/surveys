from sqlalchemy import Integer, BigInteger, Boolean, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class FlowField(Base):
    __tablename__ = "flow_fields"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    form_field_id: Mapped[int] = mapped_column(Integer, nullable=False)
    form_answer_field_id: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)
    child_id: Mapped[int] = mapped_column(Integer, nullable=False)
    section_id: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    flow_detail_id: Mapped[int] = mapped_column(Integer, nullable=False)
    sequence: Mapped[int] = mapped_column(Integer, nullable=False)
    breakdown: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    is_advanced_option: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    group_child_id: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    group_id: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)