from app.service import report_service


async def get_report_detail(req: dict, claims: dict, param: dict, slug: dict):
    report_id = slug.get("id")

    if not report_id:
        return None, "Parameter id wajib diisi", 400

    try:
        report_id = int(report_id)
    except ValueError:
        return None, "Parameter id tidak valid", 400

    result = await report_service.get_report_detail(report_id)

    if not result:
        return None, "Report tidak ditemukan", 404

    return result, "Berhasil mengambil detail report", 200

from app.service import report_service


async def cetak_laporan(req: dict, claims: dict, param: dict, slug: dict):
    laporan_id = slug.get("laporan_id")
    if not laporan_id:
        return None, "Parameter laporan_id wajib diisi", 400

    try:
        laporan_id = int(laporan_id)
    except ValueError:
        return None, "Parameter laporan_id tidak valid", 400

    user_id = claims.get("id")
    if not user_id:
        return None, "User tidak teridentifikasi", 401

    result = await report_service.enqueue_cetak_laporan(user_id, laporan_id)

    return result, "Laporan sedang diproses, kamu akan diberi notifikasi saat selesai", 202

async def test_ws(req: dict, claims: dict, param: dict, slug: dict):
    result = await report_service.test_ws(req, claims, param, slug)

    if not result:
        return None, "Report tidak ditemukan", 404

    return result, "Berhasil mengambil detail report", 200