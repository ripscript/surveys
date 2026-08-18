# app/worker/report_worker.py
import asyncio
import logging
import signal

from app.database import MasterSession
from app.repository import report_queue_repo, file_repo
from app.service import report_service
from app.config import (
    STALE_TIMEOUT_MINUTES, 
    POLL_INTERVAL_SECONDS, 
    RECOVERY_CHECK_INTERVAL_SECONDS, 
    MAX_CONCURRENT_WORKERS
)
from app.shutdown import shutdown_event
from app.repository import notification_repo

logger = logging.getLogger("py-reportapi-worker")

def handle_shutdown(*args):
    logger.info("Menerima sinyal shutdown, akan berhenti setelah job saat ini selesai...")
    shutdown_event.set()

signal.signal(signal.SIGTERM, handle_shutdown)
signal.signal(signal.SIGINT, handle_shutdown)

async def stale_job_recovery_loop():
    """Task khusus untuk mengecek dan mereset job yang macet."""
    logger.info("Recovery checker berjalan...")
    while not shutdown_event.is_set():
        try:
            async with MasterSession() as session:
                stale = await report_queue_repo.requeue_stale_jobs(session, STALE_TIMEOUT_MINUTES)
                if stale:
                    logger.warning(f"Requeue {len(stale)} stale job(s): {stale}")
        except Exception as e:
            logger.error(f"Gagal cek stale job: {e}")
        
        # Tunggu sesuai interval, tapi akan langsung tembus jika ada sinyal shutdown
        try:
            await asyncio.wait_for(shutdown_event.wait(), timeout=RECOVERY_CHECK_INTERVAL_SECONDS)
        except asyncio.TimeoutError:
            pass

async def worker_task(worker_id: int):
    """Task kasir individu yang terus menerus mengambil job dari antrian."""
    logger.info(f"Worker {worker_id} siap menerima antrian")
    while not shutdown_event.is_set():
        try:
            job = await claim_and_process(worker_id)
            if not job:
                # Jika antrian kosong, tidur sejenak lalu cek lagi
                await asyncio.sleep(POLL_INTERVAL_SECONDS)
        except Exception as e:
            logger.error(f"[Worker {worker_id}] Error pada worker loop: {e}")
            await asyncio.sleep(POLL_INTERVAL_SECONDS)
            
    logger.info(f"Worker {worker_id} berhenti")

async def run_worker_loop():
    """Menjalankan recovery task dan N buah concurrent worker."""
    logger.info(f"Report worker dimulai dengan {MAX_CONCURRENT_WORKERS} concurrent worker(s)")
    
    # 1. Daftarkan task untuk recovery job macet
    tasks = [stale_job_recovery_loop()]
    
    # 2. Daftarkan N task untuk memproses laporan secara paralel
    for i in range(1, MAX_CONCURRENT_WORKERS + 1):
        tasks.append(worker_task(i))
        
    # 3. Jalankan semuanya secara bersamaan dalam event loop
    await asyncio.gather(*tasks)
    
    logger.info("Semua worker loop berhenti dengan aman")

async def claim_and_process(worker_id: int):
    async with MasterSession() as session:
        # Asumsi: query ini menggunakan FOR UPDATE SKIP LOCKED di repository-nya
        # agar job yang sama tidak ditarik ganda oleh worker yang berbeda
        job = await report_queue_repo.claim_next_pending_job(session)

    if not job:
        return None

    job_id, user_id, laporan_id = job.id, job.user_id, job.laporan_id
    logger.info(f"[Worker {worker_id}] Memproses job {job_id} (laporan_id={laporan_id})")

    try:
        pdf_bytes = await report_service.build_report_pdf(laporan_id)
        logger.info(f"[Worker {worker_id}] PDF untuk job {job_id} berhasil dibuat, size: {len(pdf_bytes)} bytes")

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
        logger.error(f"[Worker {worker_id}] Job {job_id} gagal: {e}")
        async with MasterSession() as session:
            await report_queue_repo.mark_failed(session, job_id, str(e))

        await notification_repo.notify_report_status(
            user_id=user_id,
            laporan_id=laporan_id,
            success=False,
            error_message=str(e),
        )

    return job