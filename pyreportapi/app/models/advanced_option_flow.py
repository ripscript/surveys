from sqlalchemy import Integer, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class AdvancedOptionFlow(Base):
    __tablename__ = "advanced_option_flows"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    flow_field_id: Mapped[int] = mapped_column(Integer, nullable=False)
    form_field_id: Mapped[int] = mapped_column(Integer, nullable=False)
    option: Mapped[int] = mapped_column(Integer, nullable=False)
    child_id: Mapped[int] = mapped_column(Integer, nullable=False)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)