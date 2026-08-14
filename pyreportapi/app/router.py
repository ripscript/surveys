from app.handlers.healthy_handler import healthy
from app.handlers.report_handler import cetak_laporan, test_ws

ROUTES = {
    ("/py-reportapi/healthy", "GET"): healthy,
    ("/monitoring-dan-laporan/laporan/cetak/:laporan_id", "GET"): cetak_laporan,
    ("/py-reportapi/test-ws", "POST"): test_ws,
}


def get_handler(path: str, method: str):
    return ROUTES.get((path, method))