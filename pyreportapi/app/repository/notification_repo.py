# app/repository/notification_repo.py
import json
import logging
from app import config
from app.utils.grpc_client import hit_backend_grpc

logger = logging.getLogger("py-reportapi")


async def notify_report_status(
    user_id: int,
    laporan_id: int,
    success: bool,
    document_id: str | None = None,
    error_message: str | None = None,
):
    """
    Kirim notifikasi status laporan (success/failed) ke WSAPI,
    supaya diteruskan real-time ke frontend via websocket.
    """
    payload = {
        "user_id": user_id,
        "laporan_id": laporan_id,
        "success": success,
        "document_id": document_id,
    }
    if not success:
        payload["error_message"] = error_message

    _, status_code, message = await hit_backend_grpc(
        context=None,
        target_host=config.WSAPI_URL,
        method="POST",
        path="/wsapi/notify-report-status",
        req_body=payload,
        is_secure=False,
    )

    if status_code != 200:
        # Jangan biarkan kegagalan notif bikin job dianggap gagal juga —
        # job sudah selesai/gagal di DB, notif cuma "best effort"
        logger.error(f"Gagal kirim notifikasi WS untuk laporan {laporan_id}: {message}")