import asyncio
import json
import logging
from urllib.parse import parse_qs

from app.shutdown import shutdown_event
import signal

import grpc

from core import sic_pb2, sic_pb2_grpc
from app.config import PORT, GRPC_STOP_GRACE_SECONDS
from app.response import send_data, send_error, send_file
from app.jwt_auth import validate_token
from app.router import get_handler
from app.worker.report_worker import run_worker_loop

from app.logging_config import setup_logging
setup_logging()
logger = logging.getLogger("py-reportapi")


class ProxyServicer(sic_pb2_grpc.ProxyServicer):
    async def SendData(self, request, context):
        try:
            path = request.path
            method = request.method
            is_secure = request.isSecure

            success, message, code, claims, token = validate_token(context, is_secure)
            if not success:
                return send_error(message, code)

            handler = get_handler(path, method)
            if handler is None:
                return send_error("Path grpc tidak ditemukan", 404)

            req_data = {}
            if request.data:
                try:
                    req_data = json.loads(request.data.decode())
                except Exception:
                    req_data = {}

            slug = {}
            if request.slug:
                try:
                    slug = json.loads(request.slug.decode())
                except Exception:
                    slug = {}

            param = {}
            if request.param:
                param = parse_qs(request.param.decode())

            if asyncio.iscoroutinefunction(handler):
                result_data, result_message, result_code = await handler(req_data, claims, param, slug)
            else:
                result_data, result_message, result_code = handler(req_data, claims, param, slug)

            # Cek apakah handler mengembalikan file (PDF, dsb), bukan JSON biasa
            if isinstance(result_data, dict) and result_data.get("__is_file__"):
                return send_file(
                    file_bytes=result_data["bytes"],
                    content_type=result_data["content_type"],
                    filename=result_data["filename"],
                    code=result_code,
                    token=token,
                )

            return send_data(result_data, result_message, result_code, token=token)

        except Exception as e:
            logger.exception("Terjadi kesalahan saat memproses request")
            return send_error(f"Terjadi kendala pada service yang sedang anda akses: {e}", 500)

def handle_shutdown(*args):
    logger.info("Menerima sinyal shutdown...")
    shutdown_event.set()

signal.signal(signal.SIGTERM, handle_shutdown)
signal.signal(signal.SIGINT, handle_shutdown)

async def serve_grpc():
    server = grpc.aio.server(
        options=[
            ("grpc.max_send_message_length", 100 * 1024 * 1024),
            ("grpc.max_receive_message_length", 100 * 1024 * 1024),
        ],
    )
    sic_pb2_grpc.add_ProxyServicer_to_server(ProxyServicer(), server)
    server.add_insecure_port(f"[::]:{PORT}")
    await server.start()
    logger.info(f"gRPC async server (py-reportapi) is running on port {PORT}")

    await shutdown_event.wait()

    logger.info(f"Menghentikan gRPC server (grace={GRPC_STOP_GRACE_SECONDS}s)...")
    await server.stop(grace=GRPC_STOP_GRACE_SECONDS)
    logger.info("gRPC server berhenti")


async def main():
    """
    Jalankan gRPC server dan report worker secara bersamaan
    di satu event loop yang sama (1 proses).
    """
    await asyncio.gather(
        serve_grpc(),
        run_worker_loop(),
    )


if __name__ == "__main__":
    asyncio.run(main())