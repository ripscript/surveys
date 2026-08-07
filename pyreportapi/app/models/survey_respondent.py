from sqlalchemy import Integer, SmallInteger, String, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class SurveyRespondent(Base):
    __tablename__ = "survey_respondents"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    respondent_id: Mapped[int] = mapped_column(Integer, nullable=False)
    survey_id: Mapped[int] = mapped_column(Integer, nullable=False)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    status: Mapped[Optional[int]] = mapped_column(SmallInteger, nullable=True)
    status_approval: Mapped[Optional[str]] = mapped_column(String(20), nullable=True)
    is_edit_pertanyaan: Mapped[str] = mapped_column(String(255), nullable=False, default="false")