from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession
from datetime import datetime, timezone


async def enqueue(session: AsyncSession, user_id: int, laporan_id: int) -> int:
    result = await session.execute(
        text("""
            INSERT INTO report_queue (user_id, laporan_id, status, created_at)
            VALUES (:user_id, :laporan_id, 'pending', now())
            RETURNING id
        """),
        {"user_id": user_id, "laporan_id": laporan_id},
    )
    job_id = result.scalar_one()
    await session.commit()
    return job_id


async def claim_next_pending_job(session: AsyncSession):
    """
    FOR UPDATE SKIP LOCKED memastikan kalau ada >1 worker jalan bersamaan,
    tidak ada 2 worker yang ambil job yang sama.
    """
    result = await session.execute(
        text("""
            UPDATE report_queue
            SET status = 'processing', started_at = now()
            WHERE id = (
                SELECT id FROM report_queue
                WHERE status = 'pending'
                ORDER BY created_at
                FOR UPDATE SKIP LOCKED
                LIMIT 1
            )
            RETURNING id, user_id, laporan_id
        """)
    )
    row = result.fetchone()
    await session.commit()
    return row


async def mark_completed(session: AsyncSession, job_id: int, document_id: str):
    await session.execute(
        text("""
            UPDATE report_queue
            SET status = 'completed', document_id = :document_id, completed_at = now()
            WHERE id = :id
        """),
        {"id": job_id, "document_id": document_id},
    )
    await session.commit()


async def mark_failed(session: AsyncSession, job_id: int, error_message: str):
    await session.execute(
        text("""
            UPDATE report_queue
            SET status = 'failed', error_message = :error_message, completed_at = now()
            WHERE id = :id
        """),
        {"id": job_id, "error_message": error_message},
    )
    await session.commit()

async def requeue_stale_jobs(session: AsyncSession, timeout_minutes: int = 10):
    """
    Job yang macet di 'processing' lebih dari timeout_minutes
    dianggap gagal (kemungkinan pod-nya mati) -> reset ke pending
    """
    result = await session.execute(
        text("""
            UPDATE report_queue
            SET status = 'pending', started_at = NULL
            WHERE status = 'processing'
              AND started_at < now() - make_interval(mins => :timeout_minutes)
            RETURNING id
        """),
        {"timeout_minutes": timeout_minutes},
    )
    stale_ids = [row.id for row in result.fetchall()]
    await session.commit()
    return stale_ids