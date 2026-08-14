import io
import base64
import os
import jinja2
from weasyprint import HTML, CSS
from weasyprint.text.fonts import FontConfiguration

import matplotlib
matplotlib.use("Agg")  # non-interactive backend, wajib untuk server tanpa display
import matplotlib.pyplot as plt

from app.repository import file_repo

from app.database import SlaveSession
from app.repository import report_repo, report_queue_repo
from app.database import MasterSession

from app.utils.grpc_client import hit_backend_grpc
from app import config

# Inisiasi Jinja2
template_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), 'templates')
jinja_env = jinja2.Environment(loader=jinja2.FileSystemLoader(template_dir))

def serialize_component(comp) -> dict:
    return {
        "id": comp.id,
        "type": comp.type,
        "table_config": comp.table_config if hasattr(comp, 'table_config') else None,
        "chart_type": comp.chart_type if hasattr(comp, 'chart_type') else None,
        "chart_direction": comp.chart_direction if hasattr(comp, 'chart_direction') else None,
        "is_multiple_data": comp.is_multiple_data if hasattr(comp, 'is_multiple_data') else False,
        "form_field_ids": comp.form_field_ids if hasattr(comp, 'form_field_ids') else None,
        "narrative_text": comp.narrative_template if hasattr(comp, 'narrative_template') else "",
        "narrative_logic": comp.narrative_logic if hasattr(comp, 'narrative_logic') else None,
    }

def serialize_sub_section(sub) -> dict:
    return {
        "id": sub.id,
        "title": sub.title,
        "sequence": sub.sequence,
        "components": [serialize_component(c) for c in sub.components] if getattr(sub, 'components', None) else []
    }

def serialize_section(section) -> dict:
    return {
        "id": section.id,
        "title": section.title,
        "has_sub_section": section.has_sub_section,
        "sequence": section.sequence,
        "sub_sections": [serialize_sub_section(sub) for sub in section.sub_sections] if section.has_sub_section and getattr(section, 'sub_sections', None) else [],
        "components": [serialize_component(c) for c in section.components] if not section.has_sub_section and getattr(section, 'components', None) else []
    }

def serialize_report(report) -> dict:
    return {
        "id": report.id,
        "name": report.name,
        "sections": [serialize_section(s) for s in report.sections] if getattr(report, 'sections', None) else []
    }

async def get_report_detail(report_id: int):
    async with SlaveSession() as session:
        report = await report_repo.find_by_id(session, report_id)

    if not report:
        return None

    return serialize_report(report)


async def build_narrative_text(session, report_orm, comp_dict) -> str:
    template_text = comp_dict.get("narrative_text") or ""
    logic = comp_dict.get("narrative_logic")

    if not logic or not logic.get("variables"):
        return template_text

    for var in logic["variables"]:
        placeholder = var.get("placeholder")
        calc_type = var.get("calculation_type")
        field_id = var.get("source_form_field_id")

        value = await report_repo.calculate_narrative_variable(
            session, report_orm, calc_type, field_id
        )
        template_text = template_text.replace(placeholder, str(value))

    return template_text

async def build_report_pdf(laporan_id: int) -> bytes:
    async with SlaveSession() as session:
        # HANYA 1 QUERY SAJA KE DATABASE, SEMUANYA SUDAH TERBAWA
        report_data = await report_repo.find_by_id(session, laporan_id)

        if not report_data:
            raise Exception("Laporan tidak ditemukan")

        report_dict = serialize_report(report_data)
        cover_data = report_data.cover # Akses cover dari relasi langsung

        bg_depan_data_uri = None
        bg_belakang_data_uri = None
        if cover_data and cover_data.img_depan:
            try:
                img_bytes, mime_type = await file_repo.get_laporan_konten_image_bytes(
                    context=None, 
                    path=cover_data.img_depan
                )
                if img_bytes:
                    base64_str = base64.b64encode(img_bytes).decode('utf-8')
                    bg_depan_data_uri = f"data:{mime_type};base64,{base64_str}"
            except Exception as e:
                print(f"Gagal load cover depan: {e}")

        if cover_data and cover_data.img_belakang:
            try:
                img_bytes, mime_type = await file_repo.get_laporan_konten_image_bytes(
                    context=None, 
                    path=cover_data.img_belakang
                )
                if img_bytes:
                    base64_str = base64.b64encode(img_bytes).decode('utf-8')
                    bg_belakang_data_uri = f"data:{mime_type};base64,{base64_str}"
            except Exception as e:
                print(f"Gagal load cover belakang: {e}")

        # Helper untuk build HTML Component Table (Sama dengan sebelumnya)
        async def render_table_helper(component_dict):
            return await build_html_table_from_component(session, report_data, component_dict)

        async def render_narrative_helper(component_dict):
            return await build_narrative_text(session, report_data, component_dict)

        async def render_chart_helper(component_dict):
            return await build_chart_image_from_component(session, report_data, component_dict)

        env = jinja2.Environment(loader=jinja2.FileSystemLoader(template_dir), enable_async=True)
        template = env.get_template('report_template.html')

        # Lempar 1 variable aja (report_dict), HTML otomatis bisa baca `report.sections`, `section.sub_sections`, dst.
        html_out = await template.render_async(
            report=report_dict,      
            cover=cover_data,
            bg_depan=bg_depan_data_uri,
            bg_belakang=bg_belakang_data_uri,
            render_table=render_table_helper,
            render_narrative=render_narrative_helper,
            render_chart=render_chart_helper,
        )

    # Convert to PDF
    font_config = FontConfiguration()
    pdf_bytes = HTML(string=html_out).write_pdf(font_config=font_config)

    return pdf_bytes


async def build_html_table_from_component(session, report_orm, comp_dict) -> dict:
    table_config = comp_dict.get('table_config')
    if not table_config:
        return {}

    table_style = table_config.get('table_style', 'simple')
    show_territory = table_config.get('show_territory_col', False)
    territory_label = table_config.get('territory_col_label') or "Wilayah"

    # Kumpulkan field_id + custom_label sesuai style
    field_ids = []
    label_overrides = {}  # field_id -> custom_label

    if table_style == "grouped_header":
        for g in table_config.get('groups', []):
            for col in g.get('columns', []):
                f_id = col.get('form_field_id')
                field_ids.append(f_id)
                if col.get('custom_label'):
                    label_overrides[f_id] = col['custom_label']
    else:
        for c in table_config.get('columns', []):
            f_id = c.get('form_field_id')
            field_ids.append(f_id)
            if c.get('custom_label'):
                label_overrides[f_id] = c['custom_label']

    if not field_ids:
        return {}

    field_labels = await report_repo.get_form_field_labels(session, field_ids)
    # override label default dengan custom_label jika ada
    for f_id, custom in label_overrides.items():
        field_labels[f_id] = custom

    table_rows = await report_repo.get_table_response_data(session, report_orm, field_ids)

    headers = []
    header_row_1 = []
    header_row_2 = []

    if show_territory:
        rowspan = 2 if table_style == 'grouped_header' else 1
        header_row_1.append({"label": territory_label, "rowspan": rowspan, "colspan": 1})

    if table_style == "grouped_header":
        for g in table_config.get('groups', []):
            cols = g.get('columns', [])
            colspan = len(cols)
            header_row_1.append({"label": g.get('group_name'), "rowspan": 1, "colspan": colspan})

            for col in cols:
                f_id = col.get('form_field_id')
                lbl = field_labels.get(f_id, f"#{f_id}")
                header_row_2.append({"label": lbl, "rowspan": 1, "colspan": 1})

        headers.append(header_row_1)
        headers.append(header_row_2)
    else:
        for f_id in field_ids:
            lbl = field_labels.get(f_id, f"#{f_id}")
            header_row_1.append({"label": lbl, "rowspan": 1, "colspan": 1})
        headers.append(header_row_1)

    rows = []
    for r in table_rows:
        row_data = []
        if show_territory:
            row_data.append({"value": r.get('territory_name', '-'), "align": "left"})
        for f_id in field_ids:
            val = r.get("values", {}).get(f_id, 0)
            row_data.append({"value": val, "align": "center"})
        rows.append(row_data)

    return {
        "is_valid": True,
        "headers": headers,
        "rows": rows
    }

CHART_COLOR_PALETTE = [
    "#0070C0", "#ED7D31", "#A5A5A5", "#FFC000", "#5B9BD5",
    "#70AD47", "#264478", "#9E480E", "#636363", "#997300",
]


def _fig_to_data_uri(fig) -> str:
    buf = io.BytesIO()
    fig.savefig(buf, format="png", dpi=150, bbox_inches="tight")
    plt.close(fig)
    buf.seek(0)
    b64 = base64.b64encode(buf.read()).decode("utf-8")
    return f"data:image/png;base64,{b64}"


def _normalize_field_ids(raw) -> list[int]:
    import json as _json
    if raw is None:
        return []
    if isinstance(raw, str):
        try:
            raw = _json.loads(raw)
        except Exception:
            return []
    return [int(f) for f in raw]

def _make_pie_autopct(sizes):
    """
    autopct default matplotlib cuma dikasih persentase, bukan angka asli.
    Fungsi ini bikin closure yang balikin persentase + jumlah asli sekaligus,
    dihitung balik dari persentase karena itu satu-satunya yang dikirim matplotlib.
    """
    total = sum(sizes)

    def autopct_format(pct):
        val = int(round(pct * total / 100.0))
        return f"{pct:.1f}%\n({val})"

    return autopct_format

async def build_chart_image_from_component(session, report_orm, comp_dict) -> dict:
    chart_type = comp_dict.get("chart_type")
    field_ids = _normalize_field_ids(comp_dict.get("form_field_ids"))
    chart_direction = comp_dict.get("chart_direction") or "vertical"

    if not chart_type or not field_ids:
        return {"is_valid": False}

    plt.rcParams["font.family"] = "serif"

    # =========================================================
    # MODE PIE - MULTIPLE CHOICE (1 field, slice = opsi jawaban)
    # =========================================================
    if chart_type == "pie" and len(field_ids) == 1:
        field_types = await report_repo.get_form_field_types(session, field_ids)
        if field_types.get(field_ids[0]) == "multiple-choices":
            option_data = await report_repo.get_multiple_choice_totals(session, report_orm, field_ids[0])

            if not option_data:
                return {"is_valid": False}

            sizes = [o["count"] for o in option_data]
            labels = [o["option"] for o in option_data]

            fig, ax = plt.subplots(figsize=(6, 6))
            ax.pie(
                sizes,
                labels=labels,
                autopct=_make_pie_autopct(sizes),
                colors=CHART_COLOR_PALETTE[: len(sizes)],
                startangle=90,
            )
            ax.axis("equal")
            return {"is_valid": True, "image_data_uri": _fig_to_data_uri(fig)}

    # =========================================================
    # MODE NUMBER (existing) - bar, line, dan pie multi-field
    # =========================================================
    field_labels = await report_repo.get_form_field_labels(session, field_ids)
    table_rows = await report_repo.get_table_response_data(session, report_orm, field_ids)

    if not table_rows:
        return {"is_valid": False}

    if chart_type == "pie":
        sizes = [sum(r["values"].get(f_id, 0) for r in table_rows) for f_id in field_ids]
        labels = [field_labels.get(f_id, f"#{f_id}") for f_id in field_ids]

        if sum(sizes) <= 0:
            return {"is_valid": False}

        fig, ax = plt.subplots(figsize=(6, 6))
        ax.pie(
            sizes,
            labels=labels,
            autopct=_make_pie_autopct(sizes),
            colors=CHART_COLOR_PALETTE[: len(sizes)],
            startangle=90,
        )
        ax.axis("equal")

    else:
        # bar / line -> sumbu kategori = wilayah, tiap form_field_id jadi 1 series
        territories = [r["territory_name"] for r in table_rows]
        fig, ax = plt.subplots(figsize=(max(7.5, len(territories) * 1.0), 5.5))

        n_series = len(field_ids)
        all_values = []

        for idx, f_id in enumerate(field_ids):
            values = [r["values"].get(f_id, 0) for r in table_rows]
            all_values.extend(values)
            label = field_labels.get(f_id, f"#{f_id}")
            color = CHART_COLOR_PALETTE[idx % len(CHART_COLOR_PALETTE)]

            if chart_type == "bar":
                width = 0.8 / n_series
                positions = [i + (idx - (n_series - 1) / 2) * width for i in range(len(territories))]

                if chart_direction == "horizontal":
                    bars = ax.barh(positions, values, height=width, label=label, color=color)
                    for bar, val in zip(bars, values):
                        ax.annotate(
                            str(val),
                            xy=(bar.get_width(), bar.get_y() + bar.get_height() / 2),
                            xytext=(4, 0),
                            textcoords="offset points",
                            fontsize=8,
                            color="black",
                            va="center",
                        )
                else:
                    bars = ax.bar(positions, values, width=width, label=label, color=color)
                    for bar, val in zip(bars, values):
                        ax.annotate(
                            str(val),
                            xy=(bar.get_x() + bar.get_width() / 2, bar.get_height()),
                            xytext=(0, 4),
                            textcoords="offset points",
                            fontsize=8,
                            color="black",
                            ha="center",
                        )

            elif chart_type == "line":
                if chart_direction == "horizontal":
                    ax.plot(values, range(len(territories)), marker="o", label=label, color=color)
                    for y_pos, val in zip(range(len(territories)), values):
                        ax.annotate(
                            str(val),
                            xy=(val, y_pos),
                            xytext=(6, 0),
                            textcoords="offset points",
                            fontsize=8,
                            color="black",
                            va="center",
                        )
                else:
                    ax.plot(range(len(territories)), values, marker="o", label=label, color=color)
                    for x_pos, val in zip(range(len(territories)), values):
                        ax.annotate(
                            str(val),
                            xy=(x_pos, val),
                            xytext=(0, 8),
                            textcoords="offset points",
                            fontsize=8,
                            color="black",
                            ha="center",
                        )

        # Beri ruang ekstra di ujung sumbu nilai supaya label angka & legend
        # tidak pernah kepotong atau menimpa bar/garis paling tinggi.
        max_val = max(all_values) if all_values else 0
        headroom = max(max_val * 0.12, 1)

        tick_positions = list(range(len(territories)))
        if chart_direction == "horizontal":
            ax.set_xlim(0, max_val + headroom)
            ax.set_yticks(tick_positions)
            ax.set_yticklabels(territories, fontsize=9)
            ax.invert_yaxis()
            ax.grid(axis="x", linestyle="--", alpha=0.4)
        else:
            ax.set_ylim(0, max_val + headroom)
            ax.set_xticks(tick_positions)
            ax.set_xticklabels(territories, rotation=30, ha="right", fontsize=9)
            ax.grid(axis="y", linestyle="--", alpha=0.4)

        if chart_type == "line" or n_series > 1:
            ax.legend(
                fontsize=8,
                loc="lower center",
                bbox_to_anchor=(0.5, 1.02),
                ncol=min(n_series, 2),
                frameon=False,
            )

    return {"is_valid": True, "image_data_uri": _fig_to_data_uri(fig)}

async def test_ws(req: dict, claims: dict, param: dict, slug: dict):

    payload = {
        "user_id": 1,  # samakan dengan "id" di token JWT yang kamu pakai konek WS
        "type": "report_completed",
        "data": {
            "message": "Tes notifikasi dari pythonreport",
            "laporan_id": 123,
            "file_url": "https://example.com/laporan-123.pdf",
        },
    }

    bytesData, status_code, message = await hit_backend_grpc(
        context=None,
        target_host=config.WSAPI_URL,
        method="POST",
        path="/push-notification",
        slug_data=None,
        is_secure=False,
        req_body=payload
    )

    if status_code != 200:
        raise Exception(f"Gagal mengirim notifikasi: {message}")

    return bytesData, message, 200

async def enqueue_cetak_laporan(user_id: int, laporan_id: int) -> dict:
    async with MasterSession() as session:
        job_id = await report_queue_repo.enqueue(session, user_id, laporan_id)

    return {
        "job_id": job_id,
        "status": "processing",
    }