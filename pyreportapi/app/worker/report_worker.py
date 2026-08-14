# app/worker/report_worker.py
import asyncio
import logging
import time
import signal


from app.database import MasterSession
from app.repository import report_queue_repo, file_repo
from app.service import report_service
from app.config import STALE_TIMEOUT_MINUTES, POLL_INTERVAL_SECONDS, RECOVERY_CHECK_INTERVAL_SECONDS
from app.shutdown import shutdown_event
from app.repository import notification_repo

logger = logging.getLogger("py-reportapi-worker")

POLL_INTERVAL_SECONDS = 2

def handle_shutdown(*args):
    logger.info("Menerima sinyal shutdown, akan berhenti setelah job saat ini selesai...")
    shutdown_event.set()

signal.signal(signal.SIGTERM, handle_shutdown)
signal.signal(signal.SIGINT, handle_shutdown)

async def run_worker_loop():
    logger.info("Report worker dimulai")
    last_recovery_check = 0

    while not shutdown_event.is_set():
        now = time.time()
        if now - last_recovery_check > RECOVERY_CHECK_INTERVAL_SECONDS:
            try:
                async with MasterSession() as session:
                    stale = await report_queue_repo.requeue_stale_jobs(session, STALE_TIMEOUT_MINUTES)
                    if stale:
                        logger.warning(f"Requeue {len(stale)} stale job(s): {stale}")
            except Exception as e:
                logger.error(f"Gagal cek stale job: {e}")
            last_recovery_check = now

        try:
            job = await claim_and_process()
            if not job:
                await asyncio.sleep(POLL_INTERVAL_SECONDS)
        except Exception as e:
            logger.error(f"Worker loop error: {e}")
            await asyncio.sleep(POLL_INTERVAL_SECONDS)

    logger.info("Worker berhenti dengan aman")

async def claim_and_process():
    async with MasterSession() as session:
        job = await report_queue_repo.claim_next_pending_job(session)

    if not job:
        return None

    job_id, user_id, laporan_id = job.id, job.user_id, job.laporan_id
    logger.info(f"Memproses job {job_id} (laporan_id={laporan_id})")

    try:
        pdf_bytes = await report_service.build_report_pdf(laporan_id)
        logger.info(f"PDF berhasil dibuat, size: {len(pdf_bytes)} bytes")

        document_id = await file_repo.upload_to_document_service(pdf_bytes, laporan_id)

        async with MasterSession() as session:
            await report_queue_repo.mark_completed(session, job_id, document_id)
        
        await notification_repo.notify_report_status(
            user_id=user_id,
            laporan_id=laporan_id,
            success=True,
            document_id=document_id,
        )

    except Exception as e:
        logger.error(f"Job {job_id} gagal: {e}")
        async with MasterSession() as session:
            await report_queue_repo.mark_failed(session, job_id, str(e))

        await notification_repo.notify_report_status(
            user_id=user_id,
            laporan_id=laporan_id,
            success=False,
            error_message=str(e),
        )

    return job