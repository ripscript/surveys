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

    survey_ids = report.survey_id if isinstance(report.survey_id, list) else []
    kecamatan_ids = report.kecamatan_id if isinstance(report.kecamatan_id, list) else []
    kelurahan_ids = report.kelurahan_id if isinstance(report.kelurahan_id, list) else []
    rw_ids = report.rw_id if isinstance(report.rw_id, list) else []

    params = {"field_ids": tuple(field_ids)}
    filter_survey_sql = ""
    if survey_ids:
        filter_survey_sql = "AND survey_respondents.survey_id = ANY(:survey_ids)"
        params["survey_ids"] = tuple(survey_ids)

    breakdown = report.breakdown_level
    territory_order = []

    if breakdown == "kecamatan":
        join_table, name_col, resp_col = "kecamatans", "sub_district_name", "kecamatan_id"
        filter_ids = kecamatan_ids
    elif breakdown == "kelurahan":
        if not kelurahan_ids:
            return []
        join_table, name_col, resp_col = "kelurahans", "village_name", "kelurahan_id"
        filter_ids = kelurahan_ids
    elif breakdown == "rw":
        if not rw_ids:
            return []
        join_table, name_col, resp_col = "data__rws", "nama_rw", "rw_id"
        filter_ids = rw_ids
    else:
        return []

    order_expr = (
        f"(m.{name_col} ~ '^[0-9]+$') DESC, "
        f"CASE WHEN m.{name_col} ~ '^[0-9]+$' "
        f"THEN CAST(m.{name_col} AS INTEGER) END ASC, "
        f"m.{name_col} ASC"
    )

    filter_extra_sql = ""
    if filter_ids:
        params["filter_ids"] = tuple(filter_ids)
        filter_extra_sql = f"AND m.id = ANY(:filter_ids)"

    sql = f"""
        SELECT 
            m.{name_col} AS territory_name,
            fr.form_field_id,
            COALESCE(SUM(CAST(NULLIF(fr.answer, '') AS NUMERIC)), 0) as total_value
        FROM {join_table} m
        LEFT JOIN respondents r ON r.{resp_col} = m.id AND r.deleted_at IS NULL
        LEFT JOIN survey_respondents sr ON sr.respondent_id = r.id {filter_survey_sql.replace('survey_respondents.', 'sr.')}
        LEFT JOIN field_responses fr 
            ON fr.form_response_id = sr.id 
            AND fr.form_field_id = ANY(:field_ids)
            AND fr.deleted_at IS NULL
        WHERE m.deleted_at IS NULL {filter_extra_sql}
        GROUP BY m.{name_col}, fr.form_field_id
        ORDER BY {order_expr}
    """
    stmt_master = text(f"SELECT {name_col} FROM {join_table} m WHERE m.deleted_at IS NULL {filter_extra_sql} ORDER BY {order_expr}")

    res_master = await session.execute(stmt_master, params)
    territory_order = [getattr(row, name_col) for row in res_master.all()]

    result = await session.execute(text(sql), params)
    raw_data = result.all()

    row_map = {t_name: {f_id: 0 for f_id in field_ids} for t_name in territory_order}
    for d in raw_data:
        t_name, f_id = d.territory_name, d.form_field_id
        val = int(d.total_value) if d.total_value else 0
        if t_name in row_map and f_id is not None:
            row_map[t_name][f_id] = val

    label_map = {"kecamatan": "Kecamatan", "kelurahan": "Kelurahan", "rw": "RW"}
    child_tingkat_wilayah = label_map.get(breakdown, "Wilayah")

    return [
        {"territory_name": t_name, "tingkat_wilayah": child_tingkat_wilayah, "values": row_map[t_name]}
        for t_name in territory_order
    ]


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
    """
    survey_ids = report.survey_id if isinstance(report.survey_id, list) else []

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

    territory = _get_territory_filter(report, alias="r")
    if territory is None:
        return []
    territory_filter, territory_params = territory

    params = {"field_id": field_id, **territory_params}
    filter_survey_sql = ""
    if survey_ids:
        filter_survey_sql = "AND sr.survey_id = ANY(:survey_ids)"
        params["survey_ids"] = tuple(survey_ids)

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

    final_results = [
        {"option": label, "count": count_map.get(opt_id, 0)}
        for opt_id, label in option_master
    ]
    final_results.sort(key=lambda x: x["count"], reverse=True)
    return final_results

async def get_respondent_text_data(session: AsyncSession, report: Report, field_ids: list[int]) -> list[dict]:
    """
    Untuk table_style = 'respondent_text' / 'respondent_text_grouped'.
    1 respondent = 1 baris, lengkap dengan nama Kecamatan/Kelurahan/RW/RT,
    dan jawaban mentah dari field_responses.answer.
    """
    if not field_ids:
        return []

    survey_ids = report.survey_id if isinstance(report.survey_id, list) else []

    territory = _get_territory_filter(report, alias="r")
    if territory is None:
        return []
    territory_filter, territory_params = territory

    params = {"field_ids": tuple(field_ids), **territory_params}
    filter_survey_sql = ""
    if survey_ids:
        filter_survey_sql = "AND sr.survey_id = ANY(:survey_ids)"
        params["survey_ids"] = tuple(survey_ids)

    sql = f"""
        SELECT
            sr.id AS survey_respondent_id,
            k.sub_district_name AS kecamatan_name,
            kel.village_name AS kelurahan_name,
            rw.nama_rw AS rw_name,
            rt.nama_rt AS rt_name,
            fr.form_field_id,
            fr.answer
        FROM survey_respondents sr
        JOIN respondents r ON r.id = sr.respondent_id AND r.deleted_at IS NULL
        LEFT JOIN kecamatans k ON k.id = r.kecamatan_id AND k.deleted_at IS NULL
        LEFT JOIN kelurahans kel ON kel.id = r.kelurahan_id AND kel.deleted_at IS NULL
        LEFT JOIN data__rws rw ON rw.id = r.rw_id AND rw.deleted_at IS NULL
        LEFT JOIN data__rts rt ON rt.id = r.rt_id AND rt.deleted_at IS NULL
        LEFT JOIN field_responses fr
            ON fr.form_response_id = sr.id
            AND fr.form_field_id = ANY(:field_ids)
            AND fr.deleted_at IS NULL
        WHERE sr.status_approval = 'validated_lurah'
            {filter_survey_sql}
            {territory_filter}
        ORDER BY sr.id ASC
    """

    result = await session.execute(text(sql), params)
    raw_data = result.all()

    respondent_map = {}
    order = []

    for d in raw_data:
        sr_id = d.survey_respondent_id
        if sr_id not in respondent_map:
            respondent_map[sr_id] = {
                "kecamatan_name": d.kecamatan_name,
                "kelurahan_name": d.kelurahan_name,
                "rw_name": d.rw_name,
                "rt_name": d.rt_name,
                "answers": {},
            }
            order.append(sr_id)

        if d.form_field_id is not None:
            respondent_map[sr_id]["answers"][d.form_field_id] = d.answer

    return [respondent_map[sr_id] for sr_id in order]

async def get_maps_response_data(session: AsyncSession, report: Report, field_id: int) -> list[tuple[float, float]]:
    """
    Untuk chart_type = 'map'. Flatten semua titik lat/lng dari field_responses
    bertipe 'maps' untuk satu field_id, sesuai cakupan laporan.
    """
    survey_ids = report.survey_id if isinstance(report.survey_id, list) else []

    territory = _get_territory_filter(report, alias="r")
    if territory is None:
        return []
    territory_filter, territory_params = territory

    params = {"field_id": field_id, **territory_params}
    filter_survey_sql = ""
    if survey_ids:
        filter_survey_sql = "AND sr.survey_id = ANY(:survey_ids)"
        params["survey_ids"] = tuple(survey_ids)

    sql = f"""
        SELECT fr.answer
        FROM field_responses fr
        JOIN survey_respondents sr ON sr.id = fr.form_response_id {filter_survey_sql}
        JOIN respondents r ON r.id = sr.respondent_id AND r.deleted_at IS NULL
        WHERE fr.form_field_id = :field_id
          AND fr.deleted_at IS NULL
          AND fr.answer IS NOT NULL
          AND fr.answer != ''
          AND sr.status_approval = 'validated_lurah'
          {territory_filter}
    """
    result = await session.execute(text(sql), params)
    rows = result.all()

    points: list[tuple[float, float]] = []
    for row in rows:
        try:
            parsed = json.loads(row.answer)
        except (TypeError, ValueError):
            continue
        if not isinstance(parsed, list):
            continue
        for p in parsed:
            if not isinstance(p, dict):
                continue
            lat, lng = p.get("lat"), p.get("lng")
            if lat is None or lng is None:
                continue
            try:
                points.append((float(lat), float(lng)))
            except (TypeError, ValueError):
                continue

    return points


async def get_wilayah_names_for_report(session: AsyncSession, report: Report) -> list[str]:
    """
    Ambil daftar nama kecamatan/kelurahan (geo_name) yang jadi cakupan laporan,
    dipakai untuk filter boundary GeoJSON supaya peta hanya render wilayah
    yang relevan. Boundary polygon cuma ada di level kecamatan & kelurahan
    (RW belum ada boundary-nya), jadi breakdown "rw" tetap pakai boundary kelurahan.
    """
    kecamatan_ids = report.kecamatan_id if isinstance(report.kecamatan_id, list) else []
    kelurahan_ids = report.kelurahan_id if isinstance(report.kelurahan_id, list) else []

    breakdown = report.breakdown_level

    if breakdown == "kecamatan":
        if not kecamatan_ids:
            return []  # tingkat kota (semua kecamatan) -> tidak perlu highlight khusus
        result = await session.execute(
            text("SELECT geo_name FROM kecamatans WHERE id = ANY(:ids) AND geo_name IS NOT NULL"),
            {"ids": tuple(kecamatan_ids)},
        )
        return [row.geo_name for row in result.all()]

    if breakdown in ("kelurahan", "rw"):
        if not kelurahan_ids:
            return []
        result = await session.execute(
            text("SELECT geo_name FROM kelurahans WHERE id = ANY(:ids) AND geo_name IS NOT NULL"),
            {"ids": tuple(kelurahan_ids)},
        )
        return [row.geo_name for row in result.all()]

    return []

def _get_territory_filter(report: Report, alias: str = "r"):
    """
    Bangun filter SQL scope wilayah berdasarkan breakdown_level (BUKAN tingkat_wilayah lagi).
    Return:
        (filter_sql: str, extra_params: dict)  -> filter valid, termasuk filter kosong ("") untuk kasus admin tingkat kota
        None                                    -> scope tidak valid/kosong, caller HARUS return [] (safety, cegah query tanpa batas)
    """
    breakdown = report.breakdown_level
    kecamatan_ids = report.kecamatan_id if isinstance(report.kecamatan_id, list) else []
    kelurahan_ids = report.kelurahan_id if isinstance(report.kelurahan_id, list) else []
    rw_ids = report.rw_id if isinstance(report.rw_id, list) else []

    if breakdown == "kecamatan":
        if kecamatan_ids:
            return f"AND {alias}.kecamatan_id = ANY(:territory_ids)", {"territory_ids": tuple(kecamatan_ids)}
        # kecamatan_ids kosong -> laporan tingkat kota (Admin), tanpa filter tambahan = semua data
        return "", {}

    if breakdown == "kelurahan":
        if not kelurahan_ids:
            return None
        return f"AND {alias}.kelurahan_id = ANY(:territory_ids)", {"territory_ids": tuple(kelurahan_ids)}

    if breakdown == "rw":
        if not rw_ids:
            return None
        return f"AND {alias}.rw_id = ANY(:territory_ids)", {"territory_ids": tuple(rw_ids)}

    return None