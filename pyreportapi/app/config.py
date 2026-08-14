import os
from dotenv import load_dotenv

load_dotenv()

PORT = os.getenv("PORT", "8086")
HOST = os.getenv("HOST", "localhost")
JWT_SECRET_KEY = os.getenv("JWT_SECRET_KEY", "")

DEBUG_TO_FILE = os.getenv("DEBUG_TO_FILE", "false").lower() == "true"

# DB Master
DB_HOST_MASTER = os.getenv("DB_HOST_MASTER")
DB_PORT_MASTER = os.getenv("DB_PORT_MASTER")
DB_USER_MASTER = os.getenv("DB_USER_MASTER")
DB_PASSWORD_MASTER = os.getenv("DB_PASSWORD_MASTER")
DB_NAME_MASTER = os.getenv("DB_NAME_MASTER")
DB_SSL_MASTER = os.getenv("DB_SSL_MASTER", "disable")
TIME_OUT_DB_MASTER = int(os.getenv("TIME_OUT_DB_MASTER", "5000"))

# DB Slave
DB_HOST_SLAVE = os.getenv("DB_HOST_SLAVE")
DB_PORT_SLAVE = os.getenv("DB_PORT_SLAVE")
DB_USER_SLAVE = os.getenv("DB_USER_SLAVE")
DB_PASSWORD_SLAVE = os.getenv("DB_PASSWORD_SLAVE")
DB_NAME_SLAVE = os.getenv("DB_NAME_SLAVE")
DB_SSL_SLAVE = os.getenv("DB_SSL_SLAVE", "disable")
TIME_OUT_DB_SLAVE = int(os.getenv("TIME_OUT_DB_SLAVE", "5000"))

DOCAPI_HOST = os.getenv("DOCAPI_HOST", "localhost")
DOCAPI_PORT = os.getenv("DOCAPI_PORT", "8085")  # Sesuai port grpc DOCAPI
DOCAPI_URL = f"{DOCAPI_HOST}:{DOCAPI_PORT}"
WSAPI_HOST = os.getenv("WSAPI_HOST", "localhost")
WSAPI_PORT = os.getenv("WSAPI_PORT", "8087")
WSAPI_URL = f"{WSAPI_HOST}:{WSAPI_PORT}"

GRPC_STOP_GRACE_SECONDS = int(os.getenv("GRPC_STOP_GRACE_SECONDS", "10"))
STALE_TIMEOUT_MINUTES = int(os.getenv("STALE_TIMEOUT_MINUTES", "10"))
POLL_INTERVAL_SECONDS = int(os.getenv("POLL_INTERVAL_SECONDS", "2"))
RECOVERY_CHECK_INTERVAL_SECONDS = int(os.getenv("RECOVERY_CHECK_INTERVAL_SECONDS", "60"))