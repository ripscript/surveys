# app/repository/file_repo.py
import json
import logging
from app import config
from app.utils.grpc_client import hit_backend_grpc

logger = logging.getLogger("py-reportapi")

async def get_laporan_konten_image_bytes(context, path: str) -> tuple[bytes, str]:
    """
    Ekivalen dengan GetLaporanKontenImageBytes di Golang.
    Hit DOCAPI via gRPC untuk mengambil file bytes.
    Mengembalikan (image_bytes, mime_type)
    """
    slug_data = {"path": path}
    
    data_bytes, status_code, message = await hit_backend_grpc(
        context=context,
        target_host=config.DOCAPI_URL,
        method="GET",
        path="/internal/get-laporan-konten-image-bytes/:path",
        slug_data=slug_data,
        is_secure=False # Ubah True jika butuh token
    )

    if status_code != 200:
        raise Exception(f"Gagal mengambil gambar dari DOCAPI: {message}")

    # Python tidak punya http.DetectContentType bawaan yang akurat dari bytes secara default, 
    # tapi kita bisa import mimetypes atau magic. Atau lebih aman pakai library 'python-magic'
    # Untuk simplifikasi, asumsikan ekstensi dari path:
    import mimetypes
    mime_type, _ = mimetypes.guess_type(path)
    if not mime_type:
        mime_type = "application/octet-stream"

    return data_bytes, mime_type