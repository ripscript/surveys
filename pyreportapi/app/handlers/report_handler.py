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

async def cetak_laporan(req: dict, claims: dict, param: dict, slug: dict):
    laporan_id = slug.get("laporan_id")

    if not laporan_id:
        return None, "Parameter laporan_id wajib diisi", 400

    try:
        laporan_id = int(laporan_id)
    except ValueError:
        return None, "Parameter laporan_id tidak valid", 400

    pdf_bytes = await report_service.build_report_pdf(laporan_id)

    # tandai sebagai file response, bukan JSON biasa
    return {
        "__is_file__": True,
        "bytes": pdf_bytes,
        "content_type": "application/pdf",
        "filename": "Laporan.pdf",
    }, "OK", 200