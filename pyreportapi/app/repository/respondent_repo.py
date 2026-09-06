import json
from sqlalchemy import select, func, and_, or_, cast, Numeric, text
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from app.models.respondent import Respondent

async def find_by_id(session: AsyncSession, id: int):
    stmt = (
        select(Respondent)
        .where(Respondent.id == id)
    )
    result = await session.execute(stmt)
    return result.scalar_one_or_none()
