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
    
    stmt = text("SELECT id, question FROM form_fields WHERE id = ANY(:field_ids)")
    result = await session.execute(stmt, {"field_ids": field_ids})
    
    labels = {}
    for row in result.all():
        labels[row.id] = row.question if row.question else f"#{row.id}"
    return labels

async def get_form_field_types(session: AsyncSession, field_ids: list[int]) -> dict[int, str]:
    if not field_ids:
        return {}

    stmt = text("SELECT id, template FROM form_fields WHERE id = ANY(:field_ids)")
    result = await session.execute(stmt, {"field_ids": field_ids})

    types = {}
    for row in result.all():
        types[row.id] = row.template
    return types

async def get_table_response_data(session: AsyncSession, report: Report, field_ids: list[int]) -> list[dict]:
    if not field_ids:
        return []

    # Parsing JSON IDs dari tabel reports (Handling jika None atau dict kosong)
    survey_ids = report.survey_id if isinstance(report.survey_id, list) else []
    kecamatan_ids = report.kecamatan_id if isinstance(report.kecamatan_id, list) else []
    kelurahan_ids = report.kelurahan_id if isinstance(report.kelurahan_id, list) else []
    
    # Kumpulkan parameter dasar untuk dikirim ke raw query SQLAlchemy
    params = {"field_ids": tuple(field_ids)} # Harus tuple untuk clause IN / ANY
    
    # Filter survey (karena survey_id adalah array jsonb)
    filter_survey_sql = ""
    if survey_ids:
        filter_survey_sql = "AND survey_respondents.survey_id = ANY(:survey_ids)"
        params["survey_ids"] = tuple(survey_ids)

    sql = ""
    territory_order = []

    # =========================================================
    # LOGIKA KHUSUS TINGKAT WILAYAH = 6 (KOTA) -> Group by Semua Kecamatan
    # =========================================================
    if str(report.tingkat_wilayah) == "6":
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
        stmt_master = text("SELECT sub_district_name FROM kecamatans WHERE deleted_at IS NULL ORDER BY sub_district_name ASC")
        res_master = await session.execute(stmt_master)
        territory_order = [row.sub_district_name for row in res_master.all()]

    # =========================================================
    # LOGIKA KHUSUS TINGKAT WILAYAH = 5 (KECAMATAN) -> Group by Kecamatan Terpilih
    # =========================================================
    elif str(report.tingkat_wilayah) == "5":
        if not kecamatan_ids:
            return [] # Jika tidak ada filter kecamatan, cegah query untuk keamanan
        
        params["kecamatan_ids"] = tuple(kecamatan_ids)
        
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
              AND k.id = ANY(:kecamatan_ids)
            GROUP BY k.sub_district_name, fr.form_field_id
            ORDER BY k.sub_district_name ASC
        """
        stmt_master = text("SELECT sub_district_name FROM kecamatans WHERE deleted_at IS NULL AND id = ANY(:kecamatan_ids) ORDER BY sub_district_name ASC")
        res_master = await session.execute(stmt_master, params)
        territory_order = [row.sub_district_name for row in res_master.all()]

    # =========================================================
    # LOGIKA KHUSUS TINGKAT WILAYAH = 4 (KELURAHAN) -> Group by Kelurahan Terpilih
    # =========================================================
    elif str(report.tingkat_wilayah) == "4":
        if not kelurahan_ids:
            return [] # Jika tidak ada filter kelurahan, cegah query
        
        params["kelurahan_ids"] = tuple(kelurahan_ids)
        
        sql = f"""
            SELECT 
                kel.village_name AS territory_name,
                fr.form_field_id,
                COALESCE(SUM(CAST(NULLIF(fr.answer, '') AS NUMERIC)), 0) as total_value
            FROM kelurahans kel
            LEFT JOIN respondents r 
                ON r.kelurahan_id = kel.id AND r.deleted_at IS NULL
            LEFT JOIN survey_respondents sr 
                ON sr.respondent_id = r.id {filter_survey_sql.replace('survey_respondents.', 'sr.')}
            LEFT JOIN field_responses fr 
                ON fr.form_response_id = sr.id 
                AND fr.form_field_id = ANY(:field_ids)
                AND fr.deleted_at IS NULL
            WHERE kel.deleted_at IS NULL
              AND kel.id = ANY(:kelurahan_ids)
            GROUP BY kel.village_name, fr.form_field_id
            ORDER BY kel.village_name ASC
        """
        stmt_master = text("SELECT village_name FROM kelurahans WHERE deleted_at IS NULL AND id = ANY(:kelurahan_ids) ORDER BY village_name ASC")
        res_master = await session.execute(stmt_master, params)
        territory_order = [row.village_name for row in res_master.all()]
        
    else:
        # Fallback jika wilayah tidak dikenali
        return []

    # Eksekusi Query Gabungan
    result = await session.execute(text(sql), params)
    raw_data = result.all()

    # =========================================================
    # MERAKIT DATA (GROUPING BERDASARKAN WILAYAH)
    # =========================================================
    # Persiapkan Dictionary Map Kosong sesuai wilayah master agar baris yang datanya 0 tetap muncul
    row_map = {t_name: {f_id: 0 for f_id in field_ids} for t_name in territory_order}

    for d in raw_data:
        t_name = d.territory_name
        f_id = d.form_field_id
        val = int(d.total_value) if d.total_value else 0

        if t_name in row_map and f_id is not None:
            row_map[t_name][f_id] = val

    # Penamaan label row sesuai dengan mapping tingkat_wilayah
    child_wilayah_map = {
        "6": "Kecamatan",
        "5": "Kecamatan", 
        "4": "Kelurahan",        
    }
    child_tingkat_wilayah = child_wilayah_map.get(str(report.tingkat_wilayah), "Wilayah")

    # Format akhir menjadi list of dictionary
    final_results = []
    for t_name in territory_order:
        final_results.append({
            "territory_name": t_name,
            "tingkat_wilayah": child_tingkat_wilayah,
            "values": row_map[t_name]
        })

    return final_results


async def calculate_narrative_variable(session, report, calculation_type: str, source_form_field_id: int):
    """
    Hitung satu nilai variable narrative berdasarkan calculation_type.
    """
    # --- Grup: multiple-choice (distribusi opsi jawaban) ---
    if calculation_type in ("most_frequent_option", "most_frequent_count"):
        option_data = await get_multiple_choice_totals(session, report, source_form_field_id)
        if not option_data:
            return "-" if calculation_type == "most_frequent_option" else 0

        top_count = option_data[0]["count"]
        top_options = [o["option"] for o in option_data if o["count"] == top_count]

        if calculation_type == "most_frequent_option":
            if len(top_options) == 1:
                return top_options[0]
            return ", ".join(top_options[:-1]) + f" dan {top_options[-1]}"
        else:
            return top_count

    # --- Grup: number (agregasi per wilayah, reuse get_table_response_data) ---
    data_rows = await get_table_response_data(session, report, [source_form_field_id])

    if not data_rows:
        return "-"

    def clean_territory_name(name: str, prefix: str) -> str:
        """Hilangkan prefix wilayah jika nama sudah mengandungnya, cegah duplikasi."""
        name = (name or "").strip()
        if prefix and name.lower().startswith(prefix.lower()):
            name = name[len(prefix):].strip()
        return name

    def format_territory_names(names: list[str], prefix: str) -> str:
        """
        Gabungkan nama-nama wilayah dengan prefix HANYA SEKALI di depan seluruh frasa.
        Contoh: ["A", "B", "C"] + prefix "Kecamatan" -> "Kecamatan A, B dan C"
        """
        if not names:
            return "-"

        if len(names) == 1:
            joined = names[0]
        elif len(names) == 2:
            joined = f"{names[0]} dan {names[1]}"
        else:
            joined = ", ".join(names[:-1]) + f" dan {names[-1]}"

        return f"{prefix} {joined}".strip() if prefix else joined

    if calculation_type == "max_row_name":
        max_val = max(r["values"].get(source_form_field_id, 0) for r in data_rows)
        prefix = data_rows[0].get("tingkat_wilayah", "")
        best_names = [
            clean_territory_name(r["territory_name"], prefix)
            for r in data_rows
            if r["values"].get(source_form_field_id, 0) == max_val
        ]
        return format_territory_names(best_names, prefix)

    elif calculation_type == "max_row_value":
        return max(r["values"].get(source_form_field_id, 0) for r in data_rows)

    elif calculation_type == "min_row_name":
        min_val = min(r["values"].get(source_form_field_id, 0) for r in data_rows)
        prefix = data_rows[0].get("tingkat_wilayah", "")
        worst_names = [
            clean_territory_name(r["territory_name"], prefix)
            for r in data_rows
            if r["values"].get(source_form_field_id, 0) == min_val
        ]
        return format_territory_names(worst_names, prefix)

    elif calculation_type == "min_row_value":
        return min(r["values"].get(source_form_field_id, 0) for r in data_rows)

    elif calculation_type == "average":
        vals = [r["values"].get(source_form_field_id, 0) for r in data_rows]
        return round(sum(vals) / len(vals), 2) if vals else 0

    elif calculation_type == "total_sum":
        vals = [r["values"].get(source_form_field_id, 0) for r in data_rows]
        return sum(vals)

    return ""

async def get_multiple_choice_totals(session: AsyncSession, report: Report, field_id: int) -> list[dict]:
    """
    Hitung total responden per opsi jawaban (multiple-choice).
    field_responses.answer menyimpan form_answer_fields.id (bukan teks label),
    jadi perlu JOIN ke form_answer_fields untuk dapat label opsi ('Ya'/'Tidak'/dst).
    Semua opsi yang terdaftar di form_answer_fields untuk field ini akan selalu
    muncul di hasil (default count = 0) walau belum ada yang menjawab opsi itu.
    """
    survey_ids = report.survey_id if isinstance(report.survey_id, list) else []
    kecamatan_ids = report.kecamatan_id if isinstance(report.kecamatan_id, list) else []
    kelurahan_ids = report.kelurahan_id if isinstance(report.kelurahan_id, list) else []

    # Ambil master opsi jawaban untuk field ini (biar opsi dengan 0 jawaban tetap muncul)
    stmt_options = text("""
        SELECT id, option
        FROM form_answer_fields
        WHERE form_field_id = :field_id
        ORDER BY sequence ASC
    """)
    res_options = await session.execute(stmt_options, {"field_id": field_id})
    option_master = [(row.id, row.option) for row in res_options.all()]

    if not option_master:
        return []

    params = {"field_id": field_id}
    filter_survey_sql = ""
    if survey_ids:
        filter_survey_sql = "AND sr.survey_id = ANY(:survey_ids)"
        params["survey_ids"] = tuple(survey_ids)

    territory_filter = ""
    if str(report.tingkat_wilayah) == "5":
        if not kecamatan_ids:
            return []
        params["kecamatan_ids"] = tuple(kecamatan_ids)
        territory_filter = "AND r.kecamatan_id = ANY(:kecamatan_ids)"
    elif str(report.tingkat_wilayah) == "4":
        if not kelurahan_ids:
            return []
        params["kelurahan_ids"] = tuple(kelurahan_ids)
        territory_filter = "AND r.kelurahan_id = ANY(:kelurahan_ids)"

    sql = f"""
        SELECT 
            faf.id AS option_id,
            COUNT(*) as total_count
        FROM field_responses fr
        JOIN survey_respondents sr ON sr.id = fr.form_response_id {filter_survey_sql}
        JOIN respondents r ON r.id = sr.respondent_id AND r.deleted_at IS NULL
        JOIN form_answer_fields faf 
            ON faf.id = CAST(NULLIF(fr.answer, '') AS INTEGER)
            AND faf.form_field_id = :field_id
        WHERE fr.form_field_id = :field_id
          AND fr.deleted_at IS NULL
          AND fr.answer IS NOT NULL
          AND fr.answer != ''
          {territory_filter}
        GROUP BY faf.id
    """
    result = await session.execute(text(sql), params)
    count_map = {row.option_id: int(row.total_count) for row in result.all()}

    # Gabungkan ke master opsi supaya opsi yang belum ada yang jawab tetap muncul (0)
    final_results = [
        {"option": label, "count": count_map.get(opt_id, 0)}
        for opt_id, label in option_master
    ]
    final_results.sort(key=lambda x: x["count"], reverse=True)
    return final_results