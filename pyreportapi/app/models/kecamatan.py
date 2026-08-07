from sqlalchemy import Integer, String, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class Kecamatan(Base):
    __tablename__ = "kecamatans"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    sub_district_name: Mapped[str] = mapped_column(String(191), nullable=False)
    sub_district_slug: Mapped[str] = mapped_column(String(191), nullable=False)
    kode_wilayah: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    lat: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    long: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)