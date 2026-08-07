import asyncio
import json
import logging
from urllib.parse import parse_qs

import grpc

from core import sic_pb2, sic_pb2_grpc
from app.config import PORT
from app.response import send_data, send_error, send_file
from app.jwt_auth import validate_token
from app.router import get_handler

logging.basicConfig(level=logging.INFO)
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


async def serve():
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
    await server.wait_for_termination()


if __name__ == "__main__":
    asyncio.run(serve())