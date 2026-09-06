from sqlalchemy import Integer, BigInteger, String, Text, Date, DateTime, ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, relationship
from datetime import datetime, date
from typing import Optional, TYPE_CHECKING

from app.models.base import Base

if TYPE_CHECKING:
    from app.models.kecamatan import Kecamatan
    from app.models.kelurahan import Kelurahan


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
    kecamatan_id: Mapped[Optional[int]] = mapped_column(
        BigInteger, ForeignKey("kecamatans.id"), nullable=True
    )
    kelurahan_id: Mapped[Optional[int]] = mapped_column(
        BigInteger, ForeignKey("kelurahans.id"), nullable=True
    )
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

    kecamatan: Mapped[Optional["Kecamatan"]] = relationship(
        "Kecamatan",
        primaryjoin="foreign(Respondent.kecamatan_id) == Kecamatan.id",
        uselist=False,
        lazy="raise",
    )

    kelurahan: Mapped[Optional["Kelurahan"]] = relationship(
        "Kelurahan",
        primaryjoin="foreign(Respondent.kelurahan_id) == Kelurahan.id",
        uselist=False,
        lazy="raise",
    )