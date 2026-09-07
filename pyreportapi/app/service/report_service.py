import geopandas as gpd
from scipy.stats import gaussian_kde
from shapely.geometry import Point

import io
import base64
import os
import jinja2
from weasyprint import HTML, CSS
from weasyprint.text.fonts import FontConfiguration
import numpy as np
from scipy.interpolate import make_interp_spline
from matplotlib.patheffects import withStroke
from matplotlib.patches import FancyBboxPatch, Rectangle, Patch
from matplotlib.lines import Line2D

import matplotlib
matplotlib.use("Agg")  # non-interactive backend, wajib untuk server tanpa display
import matplotlib.pyplot as plt
import json as _json
from app.utils.debug import dprint, dprint_file

from app.repository import file_repo

from app.database import SlaveSession
from app.repository import report_repo, report_queue_repo, user_repo
from app.database import MasterSession

from app.utils.grpc_client import hit_backend_grpc
from app import config
from datetime import datetime

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
        "map_type": comp.map_type if hasattr(comp, 'map_type') else None,
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

async def build_report_pdf(user_id: int,laporan_id: int) -> bytes:
    async with SlaveSession() as session:
        report_data = await report_repo.find_by_id(session, laporan_id)

        if not report_data:
            raise Exception("Laporan tidak ditemukan")

        userData = await user_repo.find_by_id(session, user_id)

        report_dict = serialize_report(report_data)
        cover_data = report_data.cover

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

        BULAN_ID = {
            1: "Januari", 2: "Februari", 3: "Maret", 4: "April",
            5: "Mei", 6: "Juni", 7: "Juli", 8: "Agustus",
            9: "September", 10: "Oktober", 11: "November", 12: "Desember",
        }

        now = datetime.now()
        tanggal_format = f"{now.day:02d} {BULAN_ID[now.month]} {now.year}"
        # dprint_file(userData)
        html_out = await template.render_async(
            report=report_dict,  
            report_data=report_data,    
            cover=cover_data,
            bg_depan=bg_depan_data_uri,
            bg_belakang=bg_belakang_data_uri,
            render_table=render_table_helper,
            render_narrative=render_narrative_helper,
            render_chart=render_chart_helper,
            user_data=userData,
            tanggal_format=tanggal_format
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

    if table_style in ("respondent_text", "respondent_text_grouped"):
        return await build_html_table_respondent_text(session, report_orm, comp_dict)

    if table_style in ("respondent_media", "respondent_media_grouped"):
        return await build_html_table_respondent_media(session, report_orm, comp_dict)

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

async def build_html_table_respondent_text(session, report_orm, comp_dict) -> dict:
    """
    Builder untuk table_style = 'respondent_text' (flat) dan
    'respondent_text_grouped' (dengan group header, mirip pola grouped_header).

    Beda dengan build_html_table_from_component (yang agregasi angka per wilayah),
    tabel ini menampilkan data PER RESPONDENT (1 baris = 1 respondent) dengan
    kolom wajib: No | Kecamatan | Kelurahan | Rw | RT | Jawaban...

    Untuk saat ini fokus tipe pertanyaan long-answer, jawaban ditampilkan
    apa adanya dari field_responses.answer (tanpa parsing/resolusi khusus).
    """
    table_config = comp_dict.get('table_config')
    if not table_config:
        return {}

    table_style = table_config.get('table_style', 'respondent_text')
    is_grouped = table_style == "respondent_text_grouped"

    field_ids = []
    label_overrides = {}  # field_id -> custom_label

    if is_grouped:
        for g in table_config.get('groups', []) or []:
            for col in g.get('columns', []) or []:
                f_id = col.get('form_field_id')
                if f_id is None:
                    continue
                field_ids.append(f_id)
                if col.get('custom_label'):
                    label_overrides[f_id] = col['custom_label']
    else:
        for c in table_config.get('columns', []) or []:
            f_id = c.get('form_field_id')
            if f_id is None:
                continue
            field_ids.append(f_id)
            if c.get('custom_label'):
                label_overrides[f_id] = c['custom_label']

    if not field_ids:
        return {}

    field_labels = await report_repo.get_form_field_labels(session, field_ids)
    for f_id, custom in label_overrides.items():
        field_labels[f_id] = custom

    respondent_rows = await report_repo.get_respondent_text_data(session, report_orm, field_ids)

    fixed_cols = ["No", "Kecamatan", "Kelurahan", "Rw", "RT"]
    headers = []
    header_row_1 = []
    header_row_2 = []

    fixed_rowspan = 2 if is_grouped else 1
    for label in fixed_cols:
        header_row_1.append({"label": label, "rowspan": fixed_rowspan, "colspan": 1})

    if is_grouped:
        for g in table_config.get('groups', []) or []:
            cols = g.get('columns', []) or []
            colspan = len(cols)
            if colspan == 0:
                continue
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
    for idx, r in enumerate(respondent_rows, start=1):
        row_data = [
            {"value": idx, "align": "center", "type": "text"},
            {"value": r.get('kecamatan_name') or '-', "align": "left", "type": "text"},
            {"value": r.get('kelurahan_name') or '-', "align": "left", "type": "text"},
            {"value": r.get('rw_name') or '-', "align": "left", "type": "text"},
            {"value": r.get('rt_name') or '-', "align": "left", "type": "text"},
        ]
        for f_id in field_ids:
            val = r.get('answers', {}).get(f_id)
            row_data.append({"value": val if val not in (None, "") else "-", "align": "left", "type": "text"})
        rows.append(row_data)

    return {
        "is_valid": True,
        "headers": headers,
        "rows": rows
    }

CHART_COLOR_PALETTE = [
    "#DE1224",  # Cherry Red
    "#141F6B",  # Navy Blue
    "#7DD126",  # Lime Green
    "#FF8C1A",  # Tangerine
    "#8726BF",  # Violet
    "#00C7BA",  # Turquoise
    "#F22685",  # Hot Pink
    "#738021",  # Olive
    "#59ADF2",  # Sky Blue
    "#800D26",  # Burgundy
    "#F5C70D",  # Saffron Yellow
    "#616B7A",  # Slate Gray
    "#F26652",  # Coral
    "#008C59",  # Emerald
    "#AD8CD9",  # Lavender
    "#006B80",  # Teal
    "#FAB894",  # Peach
    "#333338",  # Charcoal
    "#8FE6B8",  # Mint
    "#D98500",  # Marigold
]


def _add_bar_shadow(ax, bars, orientation="vertical"):
    """Ganti bar persegi tajam jadi rounded bar + shadow tipis di belakangnya."""
    for bar in bars:
        x, y = bar.get_x(), bar.get_y()
        w, h = bar.get_width(), bar.get_height()
        color = bar.get_facecolor()

        if w <= 0 or h <= 0:
            bar.set_visible(False)
            continue

        # Offset harus proporsional ke sumbu yang relevan saja:
        # vertical -> geser sedikit secara horizontal (pakai w, bukan h)
        # horizontal -> geser sedikit secara vertical (pakai h, bukan w)
        if orientation == "vertical":
            shadow_offset = w * 0.12
            shadow = Rectangle(
                (x + shadow_offset, y), w, h,
                facecolor="#000000", alpha=0.06, zorder=bar.zorder - 1,
                linewidth=0,
            )
        else:
            shadow_offset = h * 0.12
            shadow = Rectangle(
                (x, y - shadow_offset), w, h,
                facecolor="#000000", alpha=0.06, zorder=bar.zorder - 1,
                linewidth=0,
            )
        ax.add_patch(shadow)

        bar.set_visible(False)
        rounded = FancyBboxPatch(
            (x, y), w, h,
            boxstyle=f"round,pad=0,rounding_size={min(w, h) * 0.15}",
            facecolor=color,
            edgecolor="white",
            linewidth=1.2,
            zorder=bar.zorder,
            mutation_aspect=1,
        )
        ax.add_patch(rounded)


def _smooth_curve(x_vals, y_vals):
    """Smoothing spline halus untuk line chart, hanya jika titik cukup (>=4)."""
    x_vals = np.asarray(x_vals, dtype=float)
    y_vals = np.asarray(y_vals, dtype=float)
    if len(x_vals) >= 4:
        x_new = np.linspace(x_vals.min(), x_vals.max(), 300)
        spline = make_interp_spline(x_vals, y_vals, k=3)
        y_new = np.clip(spline(x_new), 0, None)
        return x_new, y_new
    return x_vals, y_vals


def _render_pie(sizes, labels):
    """Pie chart dengan styling konsisten dengan bar/line: edge putih,
    legend di luar (bukan label nempel di wedge) supaya slice 0%/kecil
    tetap kebaca rapi, bukan cuma jadi lingkaran polos."""
    colors = CHART_COLOR_PALETTE[: len(sizes)]

    fig, ax = plt.subplots(figsize=(6.5, 6))

    wedges, _ = ax.pie(
        sizes,
        colors=colors,
        startangle=90,
        counterclock=False,
        wedgeprops=dict(edgecolor="white", linewidth=2),
        labels=None,  # label ditaruh di legend, bukan nempel di pie
    )

    total = sum(sizes)
    legend_labels = []
    for label, val in zip(labels, sizes):
        pct = (val / total * 100) if total > 0 else 0
        legend_labels.append(f"{label} — {pct:.1f}% ({val})")

    legend = ax.legend(
        wedges,
        legend_labels,
        loc="center left",
        bbox_to_anchor=(1.0, 0.5),
        fontsize=9.5,
        frameon=True,
        framealpha=0.9,
        edgecolor="#e0e0e0",
        handletextpad=0.8,
        labelspacing=0.9,
        borderpad=0.9,
    )
    legend.get_frame().set_linewidth(0.6)

    # Angka % langsung di slice, tapi hanya kalau slice cukup besar
    # biar nggak numpuk/kepotong buat slice kecil (misal 0%)
    for wedge, val in zip(wedges, sizes):
        pct = (val / total * 100) if total > 0 else 0
        if pct < 5:
            continue
        ang = (wedge.theta2 + wedge.theta1) / 2.0
        x = 0.68 * np.cos(np.deg2rad(ang))
        y = 0.68 * np.sin(np.deg2rad(ang))
        ax.annotate(
            f"{pct:.1f}%",
            xy=(x, y),
            ha="center", va="center",
            fontsize=9.5, fontweight="medium",
            color="white",
            path_effects=[withStroke(linewidth=2, foreground="#00000055")],
        )

    ax.axis("equal")
    fig.patch.set_facecolor("white")
    return fig

async def build_chart_image_from_component(session, report_orm, comp_dict) -> dict:
    chart_type = comp_dict.get("chart_type")
    field_ids = _normalize_field_ids(comp_dict.get("form_field_ids"))
    chart_direction = comp_dict.get("chart_direction") or "vertical"

    if not chart_type or not field_ids:
        return {"is_valid": False}

    plt.rcParams["font.family"] = "serif"

    # =========================================================
    # MODE MAP (baru)
    # =========================================================
    if chart_type == "map":
        if len(field_ids) != 1:
            return {"is_valid": False}

        map_type = comp_dict.get("map_type") or "point"
        points = await report_repo.get_maps_response_data(session, report_orm, field_ids[0])

        if map_type == "point":
            return await build_map_point_image(session, report_orm, points)
        if map_type == "heatmap":
            return await build_map_heatmap_image(session, report_orm, points)
        if map_type == "choropleth":
            return await build_map_choropleth_image(session, report_orm, points)

        return {"is_valid": False}  # heatmap & choropleth: menyusul

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

            fig = _render_pie(sizes, labels)
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

        fig = _render_pie(sizes, labels)

    else:
        # bar / line -> sumbu kategori = wilayah, tiap form_field_id jadi 1 series
        territories = [r["territory_name"] for r in table_rows]
        n_territories = len(territories)
        n_series = len(field_ids)

        # Figure width dinamis: sedikit kategori -> figure sempit, biar bar tidak "tenggelam"
        fig_width = max(5.5, min(4 + n_territories * 1.3, 14))
        fig, ax = plt.subplots(figsize=(fig_width, 5.5))

        all_values = []

        # Bar width proporsional, tidak terlalu kurus walau n_series banyak
        group_width = 0.72
        bar_width = max(group_width / max(n_series, 1), 0.10)

        for idx, f_id in enumerate(field_ids):
            values = [r["values"].get(f_id, 0) for r in table_rows]
            all_values.extend(values)
            label = field_labels.get(f_id, f"#{f_id}")
            color = CHART_COLOR_PALETTE[idx % len(CHART_COLOR_PALETTE)]

            if chart_type == "bar":
                positions = [
                    i + (idx - (n_series - 1) / 2) * bar_width
                    for i in range(n_territories)
                ]

                bar_kwargs = dict(
                    color=color,
                    edgecolor="white",
                    linewidth=1.2,
                    zorder=3,
                )

                if chart_direction == "horizontal":
                    bars = ax.barh(positions, values, height=bar_width * 0.88, **bar_kwargs)
                    _add_bar_shadow(ax, bars, orientation="horizontal")
                    for bar, val in zip(bars, values):
                        ax.annotate(
                            str(val),
                            xy=(bar.get_width(), bar.get_y() + bar.get_height() / 2),
                            xytext=(6, 0),
                            textcoords="offset points",
                            fontsize=8.5, fontweight="medium",
                            color="#333338",
                            va="center",
                            path_effects=[withStroke(linewidth=2.5, foreground="white")],
                        )
                else:
                    bars = ax.bar(positions, values, width=bar_width * 0.88, **bar_kwargs)
                    _add_bar_shadow(ax, bars, orientation="vertical")
                    for bar, val in zip(bars, values):
                        ax.annotate(
                            str(val),
                            xy=(bar.get_x() + bar.get_width() / 2, bar.get_height()),
                            xytext=(0, 6),
                            textcoords="offset points",
                            fontsize=8.5, fontweight="medium",
                            color="#333338",
                            ha="center",
                            path_effects=[withStroke(linewidth=2.5, foreground="white")],
                        )

            elif chart_type == "line":
                x_raw = np.arange(n_territories)

                if chart_direction == "horizontal":
                    y_new, x_new = _smooth_curve(x_raw, values)
                    ax.plot(
                        x_new, y_new,
                        color=color, linewidth=2.6, solid_capstyle="round",
                        zorder=3, alpha=0.95,
                    )
                    ax.fill_betweenx(y_new, 0, x_new, color=color, alpha=0.08, zorder=1)
                    ax.plot(
                        values, x_raw,
                        marker="o", markersize=7, linewidth=0,
                        markerfacecolor=color, markeredgecolor="white", markeredgewidth=1.6,
                        label=label, zorder=4,
                    )
                    for y_pos, val in zip(x_raw, values):
                        ax.annotate(
                            str(val),
                            xy=(val, y_pos),
                            xytext=(8, 0),
                            textcoords="offset points",
                            fontsize=8.5, fontweight="medium",
                            color="#333338",
                            va="center",
                            path_effects=[withStroke(linewidth=2.5, foreground="white")],
                        )
                else:
                    x_new, y_new = _smooth_curve(x_raw, values)
                    ax.plot(
                        x_new, y_new,
                        color=color, linewidth=2.6, solid_capstyle="round",
                        zorder=3, alpha=0.95,
                    )
                    ax.fill_between(x_new, 0, y_new, color=color, alpha=0.08, zorder=1)
                    ax.plot(
                        x_raw, values,
                        marker="o", markersize=7, linewidth=0,
                        markerfacecolor=color, markeredgecolor="white", markeredgewidth=1.6,
                        label=label, zorder=4,
                    )
                    for x_pos, val in zip(x_raw, values):
                        ax.annotate(
                            str(val),
                            xy=(x_pos, val),
                            xytext=(0, 10),
                            textcoords="offset points",
                            fontsize=8.5, fontweight="medium",
                            color="#333338",
                            ha="center",
                            path_effects=[withStroke(linewidth=2.5, foreground="white")],
                        )

        max_val = max(all_values) if all_values else 0
        headroom = max(max_val * 0.12, 1)

        tick_positions = list(range(n_territories))
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

        # Legend: hanya kalau multi-series, dengan label yang jelas & tidak duplikat
        if n_series >= 1:
            if chart_type == "line":
                legend_handles = [
                    Line2D(
                        [0], [0],
                        color=CHART_COLOR_PALETTE[i % len(CHART_COLOR_PALETTE)],
                        linewidth=2.6,
                        marker="o",
                        markersize=6,
                        markerfacecolor=CHART_COLOR_PALETTE[i % len(CHART_COLOR_PALETTE)],
                        markeredgecolor="white",
                        label=field_labels.get(f_id, f"#{f_id}"),
                    )
                    for i, f_id in enumerate(field_ids)
                ]
            else:
                legend_handles = [
                    Patch(
                        facecolor=CHART_COLOR_PALETTE[i % len(CHART_COLOR_PALETTE)],
                        edgecolor="white",
                        linewidth=0.8,
                        label=field_labels.get(f_id, f"#{f_id}"),
                    )
                    for i, f_id in enumerate(field_ids)
                ]

            legend = ax.legend(
                handles=legend_handles,
                fontsize=8.5,
                loc="upper center",
                bbox_to_anchor=(0.5, 1.22 if n_series > 1 else 1.12),
                ncol=1,
                frameon=True,
                framealpha=0.9,
                edgecolor="#e0e0e0",
                handletextpad=0.8,
                labelspacing=0.6,
                markerscale=1.4,
                borderpad=0.8,
            )
            legend.get_frame().set_linewidth(0.6)

        # Styling umum: berlaku untuk bar & line
        ax.spines["top"].set_visible(False)
        ax.spines["right"].set_visible(False)
        ax.spines["left"].set_visible(False)
        ax.tick_params(left=False)
        fig.patch.set_facecolor("white")

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

async def build_html_table_respondent_media(session, report_orm, comp_dict) -> dict:
    """
    Builder untuk table_style = 'respondent_media' (flat) dan
    'respondent_media_grouped' (dengan group header).

    Sama seperti build_html_table_respondent_text (1 baris = 1 respondent,
    kolom wajib No | Kecamatan | Kelurahan | Rw | RT), tapi kolom jawaban
    berisi GAMBAR (di-resolve dari path -> base64 data URI), bukan teks mentah.
    """
    table_config = comp_dict.get('table_config')
    if not table_config:
        return {}

    table_style = table_config.get('table_style', 'respondent_media')
    is_grouped = table_style == "respondent_media_grouped"

    field_ids = []
    label_overrides = {}  # field_id -> custom_label

    if is_grouped:
        for g in table_config.get('groups', []) or []:
            for col in g.get('columns', []) or []:
                f_id = col.get('form_field_id')
                if f_id is None:
                    continue
                field_ids.append(f_id)
                if col.get('custom_label'):
                    label_overrides[f_id] = col['custom_label']
    else:
        for c in table_config.get('columns', []) or []:
            f_id = c.get('form_field_id')
            if f_id is None:
                continue
            field_ids.append(f_id)
            if c.get('custom_label'):
                label_overrides[f_id] = c['custom_label']

    if not field_ids:
        return {}

    field_labels = await report_repo.get_form_field_labels(session, field_ids)
    for f_id, custom in label_overrides.items():
        field_labels[f_id] = custom

    # Reuse query yang sama dengan respondent_text — bedanya cuma
    # cara render nilai jawabannya (image, bukan text mentah)
    respondent_rows = await report_repo.get_respondent_text_data(session, report_orm, field_ids)

    fixed_cols = ["No", "Kecamatan", "Kelurahan", "Rw", "RT"]
    headers = []
    header_row_1 = []
    header_row_2 = []

    fixed_rowspan = 2 if is_grouped else 1
    for label in fixed_cols:
        header_row_1.append({"label": label, "rowspan": fixed_rowspan, "colspan": 1})

    if is_grouped:
        for g in table_config.get('groups', []) or []:
            cols = g.get('columns', []) or []
            colspan = len(cols)
            if colspan == 0:
                continue
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
    for idx, r in enumerate(respondent_rows, start=1):
        row_data = [
            {"value": idx, "align": "center", "type": "text"},
            {"value": r.get('kecamatan_name') or '-', "align": "left", "type": "text"},
            {"value": r.get('kelurahan_name') or '-', "align": "left", "type": "text"},
            {"value": r.get('rw_name') or '-', "align": "left", "type": "text"},
            {"value": r.get('rt_name') or '-', "align": "left", "type": "text"},
        ]
        for f_id in field_ids:
            raw_answer = r.get('answers', {}).get(f_id)
            links = _build_media_links(raw_answer)
            if links:
                row_data.append({"value": links, "align": "left", "type": "link_list"})
            else:
                row_data.append({"value": "-", "align": "center", "type": "text"})
        rows.append(row_data)

    return {
        "is_valid": True,
        "headers": headers,
        "rows": rows
    }

def _parse_media_answer(raw_answer) -> list[str]:
    """
    field_responses.answer untuk field media disimpan sebagai JSON array string, misal:
    '["2026-08-26/20260826115411000-1379cdd8.jpg", "2026-08-26/20260826115411000-56b33569.jpg"]'

    Fungsi ini parse jadi list of path relatif. Kalau bukan JSON array (misal cuma path polos /
    string kosong / None), tetap di-handle dengan aman.
    """
    if not raw_answer:
        return []

    if isinstance(raw_answer, list):
        return [p for p in raw_answer if p]

    if isinstance(raw_answer, str):
        try:
            parsed = _json.loads(raw_answer)
            if isinstance(parsed, list):
                return [p for p in parsed if p]
            if isinstance(parsed, str):
                return [parsed] if parsed else []
        except Exception:
            return [raw_answer]

    return []


def _build_media_links(raw_answer) -> list[dict]:
    paths = _parse_media_answer(raw_answer)
    if not paths:
        return []

    base_url = config.API_GATEWAY_URL.rstrip('/')
    links = []
    for idx, path in enumerate(paths, start=1):
        full_url = f"{base_url}/view-public-survey-image/{path.lstrip('/')}"
        links.append({"url": full_url, "label": f"Lihat Dokumen {idx}"})

    return links

async def build_map_point_image(session, report_orm, points: list[tuple[float, float]]) -> dict:
    if not points:
        return {"is_valid": False}

    kelurahan_gdf, kecamatan_gdf = _load_boundary_geojson()

    if str(report_orm.tingkat_wilayah) == "6":
        boundary_gdf = kecamatan_gdf
        wilayah_label = "Kecamatan"
    else:
        boundary_gdf = kelurahan_gdf
        wilayah_label = "Kelurahan"

    wilayah_names = await report_repo.get_wilayah_names_for_report(session, report_orm)
    highlight_gdf = None
    if wilayah_names:
        if str(report_orm.tingkat_wilayah) == "5":
            highlight_gdf = boundary_gdf[boundary_gdf["nama_kecamatan"].isin(wilayah_names)]
        elif str(report_orm.tingkat_wilayah) == "4":
            highlight_gdf = boundary_gdf[boundary_gdf["nama_kelurahan"].isin(wilayah_names)]

    fig, ax = plt.subplots(figsize=(7.5, 7))

    boundary_gdf.plot(
        ax=ax,
        facecolor="#F5F5F5",
        edgecolor="#BBBBBB",
        linewidth=0.5,
        zorder=1,
    )

    has_highlight = highlight_gdf is not None and not highlight_gdf.empty
    if has_highlight:
        highlight_gdf.plot(
            ax=ax,
            facecolor="#FDEBEC",
            edgecolor="#DE1224",
            linewidth=1.2,
            zorder=2,
        )

    lats = [p[0] for p in points]
    lngs = [p[1] for p in points]
    ax.scatter(
        lngs, lats,
        s=70, c="#DE1224", edgecolors="white", linewidths=1.4,
        zorder=3, alpha=0.9,
    )

    # ==== Legend ====
    legend_handles = [
        Line2D(
            [0], [0], marker="o", linewidth=0,
            markerfacecolor="#DE1224", markeredgecolor="white", markersize=9,
            label="Titik lokasi data",
        ),
    ]
    if has_highlight:
        legend_handles.append(
            Patch(facecolor="#FDEBEC", edgecolor="#DE1224", linewidth=1.2,
                  label=f"Cakupan {wilayah_label} laporan")
        )
    legend_handles.append(
        Patch(facecolor="#F5F5F5", edgecolor="#BBBBBB", linewidth=0.5,
              label=f"Batas {wilayah_label} lainnya")
    )

    legend = ax.legend(
        handles=legend_handles,
        loc="upper left",
        bbox_to_anchor=(1.0, 1.0),
        fontsize=9,
        frameon=True,
        framealpha=0.9,
        edgecolor="#e0e0e0",
        handletextpad=0.8,
        labelspacing=0.9,
        borderpad=0.9,
    )
    legend.get_frame().set_linewidth(0.6)

    ax.set_xticks([])
    ax.set_yticks([])
    for spine in ax.spines.values():
        spine.set_visible(False)
    ax.set_aspect("equal", adjustable="datalim")
    fig.patch.set_facecolor("white")

    return {"is_valid": True, "image_data_uri": _fig_to_data_uri(fig)}

_GEO_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), 'app','assets', 'geo')
_KELURAHAN_GDF = None
_KECAMATAN_GDF = None

def _load_boundary_geojson():
    global _KELURAHAN_GDF, _KECAMATAN_GDF
    if _KELURAHAN_GDF is None:
        _KELURAHAN_GDF = gpd.read_file(os.path.join(_GEO_DIR, "kelurahan.json"))
    if _KECAMATAN_GDF is None:
        _KECAMATAN_GDF = gpd.read_file(os.path.join(_GEO_DIR, "kecamatan.json"))
    return _KELURAHAN_GDF, _KECAMATAN_GDF


async def build_map_heatmap_image(session, report_orm, points: list[tuple[float, float]]) -> dict:
    if len(points) < 3:
        # KDE butuh minimal beberapa titik buat estimasi kepadatan yang berarti;
        # kalau terlalu sedikit, fallback ke titik biasa daripada gagal/kosong
        return await build_map_point_image(session, report_orm, points)

    kelurahan_gdf, kecamatan_gdf = _load_boundary_geojson()

    if str(report_orm.tingkat_wilayah) == "6":
        boundary_gdf = kecamatan_gdf
        wilayah_label = "Kecamatan"
    else:
        boundary_gdf = kelurahan_gdf
        wilayah_label = "Kelurahan"

    wilayah_names = await report_repo.get_wilayah_names_for_report(session, report_orm)
    highlight_gdf = None
    if wilayah_names:
        if str(report_orm.tingkat_wilayah) == "5":
            highlight_gdf = boundary_gdf[boundary_gdf["nama_kecamatan"].isin(wilayah_names)]
        elif str(report_orm.tingkat_wilayah) == "4":
            highlight_gdf = boundary_gdf[boundary_gdf["nama_kelurahan"].isin(wilayah_names)]

    lats = np.array([p[0] for p in points])
    lngs = np.array([p[1] for p in points])

    # Hitung densitas KDE. Kalau titiknya semua berhimpitan persis (variance = 0),
    # gaussian_kde akan gagal (singular matrix) -> fallback ke titik biasa.
    try:
        kde = gaussian_kde(np.vstack([lngs, lats]))
    except np.linalg.LinAlgError:
        return await build_map_point_image(session, report_orm, points)

    # Grid evaluasi kepadatan, dengan padding secukupnya di sekitar sebaran titik
    pad_x = max((lngs.max() - lngs.min()) * 0.3, 0.01)
    pad_y = max((lats.max() - lats.min()) * 0.3, 0.01)
    grid_x = np.linspace(lngs.min() - pad_x, lngs.max() + pad_x, 200)
    grid_y = np.linspace(lats.min() - pad_y, lats.max() + pad_y, 200)
    xx, yy = np.meshgrid(grid_x, grid_y)
    zz = kde(np.vstack([xx.ravel(), yy.ravel()])).reshape(xx.shape)

    fig, ax = plt.subplots(figsize=(7.5, 7))

    boundary_gdf.plot(
        ax=ax,
        facecolor="none",
        edgecolor="#999999",
        linewidth=0.5,
        zorder=1,
    )

    has_highlight = highlight_gdf is not None and not highlight_gdf.empty
    if has_highlight:
        highlight_gdf.plot(
            ax=ax,
            facecolor="none",
            edgecolor="#DE1224",
            linewidth=1.4,
            zorder=2,
        )

    contour = ax.contourf(
        xx, yy, zz,
        levels=12,
        cmap="YlOrRd",
        alpha=0.75,
        zorder=3,
    )

    cbar = fig.colorbar(contour, ax=ax, shrink=0.7, pad=0.02)
    cbar.set_label("Kepadatan titik dilaporkan", fontsize=9)
    cbar.ax.tick_params(labelsize=8)

    legend_handles = []
    if has_highlight:
        legend_handles.append(
            Patch(facecolor="none", edgecolor="#DE1224", linewidth=1.4,
                  label=f"Cakupan {wilayah_label} laporan")
        )
    legend_handles.append(
        Patch(facecolor="none", edgecolor="#999999", linewidth=0.5,
              label=f"Batas {wilayah_label}")
    )
    if legend_handles:
        legend = ax.legend(
            handles=legend_handles,
            loc="upper left",
            bbox_to_anchor=(1.25, 1.0),
            fontsize=9,
            frameon=True,
            framealpha=0.9,
            edgecolor="#e0e0e0",
        )
        legend.get_frame().set_linewidth(0.6)

    ax.set_xticks([])
    ax.set_yticks([])
    for spine in ax.spines.values():
        spine.set_visible(False)
    ax.set_aspect("equal", adjustable="datalim")
    fig.patch.set_facecolor("white")

    return {"is_valid": True, "image_data_uri": _fig_to_data_uri(fig)}


async def build_map_choropleth_image(session, report_orm, points: list[tuple[float, float]]) -> dict:
    if not points:
        return {"is_valid": False}

    tingkat = str(report_orm.tingkat_wilayah)
    if tingkat not in ("5", "6"):
        # Safety net — seharusnya sudah diblok di validasi request (Go),
        # tapi tetap dijaga di sini karena choropleth butuh boundary RW yang belum ada.
        return {"is_valid": False}

    kelurahan_gdf, kecamatan_gdf = _load_boundary_geojson()

    if tingkat == "6":
        boundary_gdf = kecamatan_gdf.copy()
        wilayah_label = "Kecamatan"
    else:
        boundary_gdf = kelurahan_gdf.copy()
        wilayah_label = "Kelurahan"

    # Bungkus titik lat/lng jadi GeoDataFrame of Point, CRS harus SAMA
    # dengan boundary supaya spatial join akurat
    points_gdf = gpd.GeoDataFrame(
        geometry=[Point(lng, lat) for lat, lng in points],
        crs=boundary_gdf.crs,
    )

    # Titik yang jatuh DI LUAR semua polygon (misal koordinat aneh/typo input)
    # otomatis punya index_right NaN dari sjoin, dan tidak ikut ke-count wilayah manapun
    joined = gpd.sjoin(points_gdf, boundary_gdf, predicate="within", how="left")
    counts = joined.groupby("index_right").size()

    boundary_gdf["jumlah_titik"] = boundary_gdf.index.map(counts).fillna(0).astype(int)

    fig, ax = plt.subplots(figsize=(8, 7))

    boundary_gdf.plot(
        column="jumlah_titik",
        ax=ax,
        cmap="YlOrRd",
        edgecolor="white",
        linewidth=0.6,
        legend=True,
        legend_kwds={"label": "Jumlah titik dilaporkan", "shrink": 0.7},
        zorder=1,
    )

    # Angka jumlah di tengah polygon — hanya kalau jumlah wilayah tidak
    # terlalu banyak (mis. 30 kecamatan masih rapi, 151 kelurahan akan
    # penuh sesak dan justru bikin peta susah dibaca)
    if len(boundary_gdf) <= 40:
        for _, row in boundary_gdf.iterrows():
            if row["jumlah_titik"] > 0:
                centroid = row.geometry.centroid
                ax.annotate(
                    str(row["jumlah_titik"]),
                    xy=(centroid.x, centroid.y),
                    ha="center", va="center",
                    fontsize=7, fontweight="bold", color="#333338",
                    path_effects=[withStroke(linewidth=2, foreground="white")],
                )

    ax.set_xticks([])
    ax.set_yticks([])
    for spine in ax.spines.values():
        spine.set_visible(False)
    ax.set_aspect("equal")
    fig.patch.set_facecolor("white")

    return {"is_valid": True, "image_data_uri": _fig_to_data_uri(fig)}