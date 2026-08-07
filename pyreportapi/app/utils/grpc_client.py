import json
import logging
from typing import Dict, Any, Tuple, Optional

import grpc
from core import sic_pb2, sic_pb2_grpc

logger = logging.getLogger("py-reportapi")

async def hit_backend_grpc(
    context,          # Context dari request yang masuk (untuk ekstrak token JWT)
    target_host: str, # host:port dari service target (misal: "localhost:8081")
    method: str,      # "GET", "POST", "PUT", dll
    path: str,        # "/internal/get-laporan-konten-image-bytes/:path"
    slug_data: Optional[Dict[str, Any]] = None,
    req_body: Optional[Dict[str, Any]] = None,
    is_secure: bool = False
) -> Tuple[bytes, int, str]:
    """
    Hit service internal lain via gRPC (Bypass API Gateway).
    Mengembalikan: (data_bytes, status_code, message)
    """
    try:
        # 1. Ekstrak Authorization Header dari context yang masuk (jika butuh passing token)
        metadata = dict(context.invocation_metadata()) if context else {}
        auth_token = metadata.get("authorization", "")
        
        # Metadata yang akan dikirim ke service target
        outbound_metadata = []
        if auth_token:
            outbound_metadata.append(("authorization", auth_token))

        # 2. Siapkan Request Payload
        request = sic_pb2.ProxyRequest(
            path=path,
            method=method,
            isSecure=is_secure
        )

        if slug_data:
            request.slug = json.dumps(slug_data).encode('utf-8')
            
        if req_body:
            request.data = json.dumps(req_body).encode('utf-8')

        # 3. Buka koneksi gRPC dan eksekusi
        # Gunakan insekur channel untuk komunikasi internal backend
        async with grpc.aio.insecure_channel(target_host) as channel:
            stub = sic_pb2_grpc.ProxyStub(channel)
            
            # Panggil method SendData dari protobuf
            response = await stub.SendData(request, metadata=outbound_metadata)
            
            if not response.success:
                logger.error(f"gRPC Call Failed to {target_host}{path}: {response.message}")
                return b"", response.code, response.message

            return response.data, response.code, response.message

    except grpc.aio.AioRpcError as e:
        logger.error(f"gRPC RPC Error ke {target_host}: {e.details()}")
        return b"", 500, f"Gagal terhubung ke service target: {e.details()}"
    except Exception as e:
        logger.error(f"gRPC Exception ke {target_host}: {str(e)}")
        return b"", 500, f"Internal Error saat gRPC call: {str(e)}"