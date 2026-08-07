from app.handlers.healthy_handler import healthy
from app.handlers.report_handler import cetak_laporan

ROUTES = {
    ("/py-reportapi/healthy", "GET"): healthy,
    ("/monitoring-dan-laporan/laporan/cetak/:laporan_id", "GET"): cetak_laporan,
}


def get_handler(path: str, method: str):
    return ROUTES.get((path, method))