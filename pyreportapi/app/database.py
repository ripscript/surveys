from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from app import config


def build_dsn(user, password, host, port, dbname, ssl_mode):
    # asyncpg driver
    sslmode = "require" if ssl_mode == "require" else "disable"
    return (
        f"postgresql+asyncpg://{user}:{password}@{host}:{port}/{dbname}"
    )


# === MASTER ENGINE (untuk write operation) ===
master_engine = create_async_engine(
    build_dsn(
        config.DB_USER_MASTER,
        config.DB_PASSWORD_MASTER,
        config.DB_HOST_MASTER,
        config.DB_PORT_MASTER,
        config.DB_NAME_MASTER,
        config.DB_SSL_MASTER,
    ),
    pool_size=5,          # setara SetMaxOpenConns(5)
    max_overflow=0,
    pool_recycle=3600,    # 1 jam, setara ConnMaxLifetime
    pool_timeout=config.TIME_OUT_DB_MASTER / 1000,  # ms -> detik
    echo=False,
)

MasterSession = async_sessionmaker(
    bind=master_engine,
    expire_on_commit=False,
    class_=AsyncSession,
)


# === SLAVE ENGINE (untuk read operation) ===
slave_engine = create_async_engine(
    build_dsn(
        config.DB_USER_SLAVE,
        config.DB_PASSWORD_SLAVE,
        config.DB_HOST_SLAVE,
        config.DB_PORT_SLAVE,
        config.DB_NAME_SLAVE,
        config.DB_SSL_SLAVE,
    ),
    pool_size=5,
    max_overflow=0,
    pool_recycle=3600,
    pool_timeout=config.TIME_OUT_DB_SLAVE / 1000,
    echo=True,   # setara logger.Info di GORM slave
)

SlaveSession = async_sessionmaker(
    bind=slave_engine,
    expire_on_commit=False,
    class_=AsyncSession,
)


async def get_master_session() -> AsyncSession:
    async with MasterSession() as session:
        yield session


async def get_slave_session() -> AsyncSession:
    async with SlaveSession() as session:
        yield session