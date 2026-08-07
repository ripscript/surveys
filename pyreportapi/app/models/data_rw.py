from sqlalchemy import Integer, BigInteger, String, Text, Date, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from datetime import datetime, date
from typing import Optional

from app.models.base import Base


class DataRw(Base):
    __tablename__ = "data__rws"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    kelurahan_id: Mapped[int] = mapped_column(BigInteger, nullable=False)
    nama_rw: Mapped[str] = mapped_column(String(191), nullable=False)
    nik: Mapped[Optional[str]] = mapped_column(String(20), nullable=True)
    nama_pejabat: Mapped[Optional[str]] = mapped_column(String(100), nullable=True)
    lahir: Mapped[Optional[date]] = mapped_column(Date, nullable=True)
    alamat: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    hp: Mapped[Optional[str]] = mapped_column(String(15), nullable=True)
    lat: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    long: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)
    created_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    updated_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime, nullable=True)
    kode_wilayah: Mapped[Optional[str]] = mapped_column(String(191), nullable=True)