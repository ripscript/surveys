import json
from sqlalchemy import select, func, and_, or_, cast, Numeric, text
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from app.models.report import Report
from app.models.report_cover import ReportCover
from app.models.report_section import ReportSection
from app.models.report_subsection import ReportSubsection

async def find_by_id(session: AsyncSession, report_id: int):
    stmt = (
        select(Report)
        .options(
            # Preload Cover
            selectinload(Report.cover),
            # Preload Section -> Components (Jika tidak ada sub section)
            selectinload(Report.sections).selectinload(ReportSection.components),
            # Preload Section -> Sub Sections -> Components (Jika ada sub section)
            selectinload(Report.sections).selectinload(ReportSection.sub_sections).selectinload(ReportSubsection.components)
        )
        .where(Report.id == report_id)
    )
    result = await session.execute(stmt)
    return result.scalar_one_or_none()

async def get_report_cover_by_id(session: AsyncSession, report_id: int):
    stmt = select(ReportCover).where(ReportCover.report_id == report_id)
    result = await session.execute(stmt)
    return result.scalar_one_or_none()

async def get_report_sections_by_report_id(session: AsyncSession, report_id: int):
    stmt = select(ReportSection).where(ReportSection.report_id == report_id)
    result = await session.execute(stmt)
    return result.scalars().all()

async def get_report_sub_sections_by_report_section_id(session: AsyncSession, report_section_id: int):
    stmt = select(ReportSubsection).where(ReportSubsection.report_section_id == report_section_id)
    result = await session.execute(stmt)
    return result.scalars().all()

async def get_form_field_labels(session: AsyncSession, field_ids: list[int]) -> dict[int, str]:
    if not field_ids:
        return {}
    
    # Asumsikan Anda punya model FormField, atau gunakan text() raw query
    from sqlalchemy import text
    stmt = text("SELECT id, question FROM form_fields WHERE id = ANY(:field_ids)")
    result = await session.execute(stmt, {"field_ids": field_ids})
    
    labels = {}
    for row in result.all():
        labels[row.id] = row.question if row.question else f"#{row.id}"
    return labels

async def get_table_response_data(session: AsyncSession, report: Report, field_ids: list[int]) -> list[dict]:
    if not field_ids:
        return []

    # Parsing JSON IDs (Handling jika None atau dict kosong)
    survey_ids = report.survey_id if isinstance(report.survey_id, list) else []
    
    # Kumpulkan parameter untuk dikirim ke raw query SQLAlchemy
    params = {"field_ids": tuple(field_ids)} # Harus tuple untuk clause IN / ANY
    
    # Filter survey (karena survey_id adalah array di struct golang Anda)
    filter_survey_sql = ""
    if survey_ids:
        filter_survey_sql = "AND survey_respondents.survey_id = ANY(:survey_ids)"
        params["survey_ids"] = survey_ids

    # =========================================================
    # LOGIKA KHUSUS TINGKAT WILAYAH = 6 (KOTA -> KECAMATAN)
    # =========================================================
    if str(report.tingkat_wilayah) == "6":
        # Gunakan tabel master kecamatans sebagai base (LEFT JOIN ke data)
        # Tujuannya agar SELURUH kecamatan di Kota Bandung muncul, 
        # meskipun valuenya 0 / tidak ada jawaban.
        
        sql = f"""
            SELECT 
                k.sub_district_name AS territory_name,
                fr.form_field_id,
                COALESCE(SUM(CAST(NULLIF(fr.answer, '') AS NUMERIC)), 0) as total_value
            FROM kecamatans k
            LEFT JOIN respondents r 
                ON r.kecamatan_id = k.id AND r.deleted_at IS NULL
            LEFT JOIN survey_respondents sr 
                ON sr.respondent_id = r.id {filter_survey_sql.replace('survey_respondents.', 'sr.')}
            LEFT JOIN field_responses fr 
                ON fr.form_response_id = sr.id 
                AND fr.form_field_id = ANY(:field_ids)
                AND fr.deleted_at IS NULL
            WHERE k.deleted_at IS NULL
            GROUP BY k.sub_district_name, fr.form_field_id
            ORDER BY k.sub_district_name ASC
        """
        
    # TODO: Tambahkan kondisi elif untuk wilayah 5 (Kecamatan -> Kelurahan)
    # TODO: Tambahkan kondisi elif untuk wilayah 4 (Kelurahan -> RW)
    else:
        # Fallback jika wilayah tidak dikenali
        return []

    # Eksekusi Query menggunakan text()
    result = await session.execute(text(sql), params)
    raw_data = result.all()

    # =========================================================
    # MERAKIT DATA (GROUPING BERDASARKAN WILAYAH)
    # =========================================================
    
    # Ambil daftar semua wilayah (Kecamatan) terlebih dahulu agar urutan terjamin
    territory_order = []
    if str(report.tingkat_wilayah) == "6":
        # Ambil semua kecamatan asli dari DB agar kalau datanya 0 semua, baris kecamatan tetap ada
        stmt_master = text("SELECT sub_district_name FROM kecamatans WHERE deleted_at IS NULL ORDER BY sub_district_name ASC")
        res_master = await session.execute(stmt_master)
        territory_order = [row.sub_district_name for row in res_master.all()]

    row_map = {t_name: {f_id: 0 for f_id in field_ids} for t_name in territory_order}

    for d in raw_data:
        t_name = d.territory_name
        f_id = d.form_field_id
        val = int(d.total_value) if d.total_value else 0

        # Jika form_field_id terisi (artinya ada data dari LEFT JOIN), simpan nilainya
        if t_name in row_map and f_id is not None:
            row_map[t_name][f_id] = val

    # Format akhir menjadi list of dictionary
    final_results = []
    for t_name in territory_order:
        final_results.append({
            "territory_name": t_name,
            "values": row_map[t_name]
        })

    return final_results


async def calculate_narrative_variable(session, report, calculation_type: str, source_form_field_id: int):
    """
    Hitung satu nilai variable narrative berdasarkan calculation_type.
    Reuse get_table_response_data agar konsisten dengan data tabel.
    """
    data_rows = await get_table_response_data(session, report, [source_form_field_id])
    # data_rows: [{"territory_name": ..., "values": {field_id: val}}, ...]

    if not data_rows:
        return "-"

    if calculation_type == "max_row_name":
        best = max(data_rows, key=lambda r: r["values"].get(source_form_field_id, 0))
        return best["territory_name"]

    elif calculation_type == "max_value":
        best = max(data_rows, key=lambda r: r["values"].get(source_form_field_id, 0))
        return best["values"].get(source_form_field_id, 0)

    elif calculation_type == "min_row_name":
        worst = min(data_rows, key=lambda r: r["values"].get(source_form_field_id, 0))
        return worst["territory_name"]

    elif calculation_type == "min_value":
        worst = min(data_rows, key=lambda r: r["values"].get(source_form_field_id, 0))
        return worst["values"].get(source_form_field_id, 0)

    elif calculation_type == "average":
        vals = [r["values"].get(source_form_field_id, 0) for r in data_rows]
        return round(sum(vals) / len(vals), 2) if vals else 0

    elif calculation_type == "sum":
        vals = [r["values"].get(source_form_field_id, 0) for r in data_rows]
        return sum(vals)

    return ""