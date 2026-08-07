import json
from core import sic_pb2


def send_data(data=None, message: str = "Berhasil", code: int = 200, token: str = ""):
    try:
        payload = json.dumps(data if data is not None else {}, default=str).encode()
    except Exception as e:
        return send_error(f"Terjadi kesalahan saat encode data: {e}", 500)

    return sic_pb2.ProxyResponse(
        data=payload,
        success=True,
        message=message,
        code=code,
        token=token,
    )


def send_file(file_bytes: bytes, content_type: str, filename: str, code: int = 200, token: str = ""):
    """
    Mengikuti pola Go: message = "Data File,<content-type>,<filename>"
    data = raw bytes file (bukan JSON)
    """
    message = f"Data File,{content_type},{filename}"
    return sic_pb2.ProxyResponse(
        data=file_bytes,
        success=True,
        message=message,
        code=code,
        token=token,
    )


def send_error(message: str, code: int = 500, token: str = ""):
    return sic_pb2.ProxyResponse(
        data=b"",
        success=False,
        message=message,
        code=code,
        token=token,
    )