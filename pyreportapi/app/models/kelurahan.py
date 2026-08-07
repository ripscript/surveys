from sqlalchemy import Integer, BigInteger, String, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class Kelurahan(Base):
    __tablename__ = "kelurahans"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    sub_district_id: Mapped[int] = mapped_column(BigInteger, nullable=False)
    village_name: Mapped[str] = mapped_column(String(50), nullable=False)
    village_name_slug: Mapped[str] = mapped_column(String(70), nullable=False)
    village_postal_code: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    kode_wilayah: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    lat: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    long: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)