import logging
from logging.handlers import RotatingFileHandler

from app.config import DEBUG_TO_FILE


def setup_logging():
    handlers = [logging.StreamHandler()]

    if DEBUG_TO_FILE:
        file_handler = RotatingFileHandler(
            "debug.log",
            maxBytes=5 * 1024 * 1024,
            backupCount=3,
            encoding="utf-8",
        )
        handlers.append(file_handler)

    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        handlers=handlers,
        force=True,
    )

    if DEBUG_TO_FILE:
        logging.getLogger("py-reportapi").info("Debug logging ke file AKTIF (debug.log)")