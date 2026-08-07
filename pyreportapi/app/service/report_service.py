import base64
import os
import jinja2
from weasyprint import HTML, CSS
from weasyprint.text.fonts import FontConfiguration

from app.repository import file_repo

from app.database import SlaveSession
from app.repository import report_repo

# Inisiasi Jinja2
template_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), 'templates')
jinja_env = jinja2.Environment(loader=jinja2.FileSystemLoader(template_dir))

def serialize_component(comp) -> dict:
    return {
        "id": comp.id,
        "type": comp.type,
        "table_config": comp.table_config if hasattr(comp, 'table_config') else None,
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

        # Helper untuk build HTML Component Table (Sama dengan sebelumnya)
        async def render_table_helper(component_dict):
            return await build_html_table_from_component(session, report_data, component_dict)

        async def render_narrative_helper(component_dict):
            return await build_narrative_text(session, report_data, component_dict)

        env = jinja2.Environment(loader=jinja2.FileSystemLoader(template_dir), enable_async=True)
        template = env.get_template('report_template.html')

        # Lempar 1 variable aja (report_dict), HTML otomatis bisa baca `report.sections`, `section.sub_sections`, dst.
        html_out = await template.render_async(
            report=report_dict,      
            cover=cover_data,
            bg_depan=bg_depan_data_uri,
            render_table=render_table_helper,
            render_narrative=render_narrative_helper,
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