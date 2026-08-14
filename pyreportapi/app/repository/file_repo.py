# app/repository/file_repo.py
import base64
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

async def upload_to_document_service(pdf_bytes: bytes, laporan_id: int, context=None) -> str:
    b64_data = base64.b64encode(pdf_bytes).decode("utf-8")
    data_uri = f"data:application/pdf;base64,{b64_data}"

    req_body = {"datauri": data_uri}

    data_bytes, status_code, message = await hit_backend_grpc(
        context=context,
        target_host=config.DOCAPI_URL,
        method="POST",
        path="/upload-temp-laporan",
        req_body=req_body,
        is_secure=False,
    )

    if status_code != 200:
        raise Exception(f"Gagal upload PDF ke document service: {message}")

    try:
        result_data = json.loads(data_bytes.decode("utf-8")) if data_bytes else None
    except Exception:
        result_data = None

    document_id = result_data.get("data") if isinstance(result_data, dict) else result_data
    if not document_id:
        raise Exception(f"Response upload tidak berisi filename/document_id: {result_data}")

    logger.info(f"Laporan {laporan_id} berhasil diupload sebagai {document_id}")
    return document_id