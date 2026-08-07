from sqlalchemy import Integer, BigInteger, String, Text, Date, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime, date
from typing import Optional

from app.models.base import Base


class Respondent(Base):
    __tablename__ = "respondents"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    name: Mapped[str] = mapped_column(String(70), nullable=False)
    phone_number: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    email: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    blk_id: Mapped[int] = mapped_column(BigInteger, nullable=False)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    nik: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    kecamatan_id: Mapped[Optional[int]] = mapped_column(BigInteger, nullable=True)
    kelurahan_id: Mapped[Optional[int]] = mapped_column(BigInteger, nullable=True)
    rw_id: Mapped[Optional[int]] = mapped_column(BigInteger, nullable=True)
    rt_id: Mapped[Optional[int]] = mapped_column(BigInteger, nullable=True)
    role_id: Mapped[int] = mapped_column(BigInteger, nullable=False)
    tanggal_lahir: Mapped[Optional[date]] = mapped_column(Date, nullable=True)
    alamat: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    tempat_lahir: Mapped[Optional[str]] = mapped_column(String(100), nullable=True)
    is_blocked: Mapped[str] = mapped_column(String(255), nullable=False, default="false")
    username: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    avatar: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)