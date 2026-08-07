from sqlalchemy import Integer, BigInteger, String, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime
from typing import Optional

from app.models.base import Base


class DataRt(Base):
    __tablename__ = "data__rts"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    rw_id: Mapped[int] = mapped_column(BigInteger, nullable=False)
    nama_rt: Mapped[str] = mapped_column(String(4), nullable=False)
    lat: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    long: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    kode_wilayah: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)