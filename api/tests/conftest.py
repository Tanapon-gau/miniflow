import asyncio
import os
from collections.abc import AsyncGenerator
from urllib.parse import urlparse, urlunparse

import asyncpg
import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.pool import NullPool

from app.db import DATABASE_URL as APP_DATABASE_URL
from app.db import Base
from app.deps import get_session
from app.main import app


def _derive_test_database_url() -> str:
    """Return a test DB URL distinct from the app's prod DB.

    Tests call ``Base.metadata.drop_all`` on teardown, which would wipe a
    shared database. Use ``TEST_DATABASE_URL`` if set, otherwise suffix the
    app DB name with ``_test`` so prod data stays untouched.
    """
    override = os.getenv("TEST_DATABASE_URL")
    if override:
        return override

    parsed = urlparse(APP_DATABASE_URL)
    db_name = parsed.path.lstrip("/")
    if db_name.endswith("_test"):
        return APP_DATABASE_URL
    return urlunparse(parsed._replace(path=f"/{db_name}_test"))


TEST_DATABASE_URL = _derive_test_database_url()


async def _ensure_test_database_exists() -> None:
    parsed = urlparse(TEST_DATABASE_URL)
    test_db_name = parsed.path.lstrip("/")
    # Connect to the default 'postgres' DB to issue CREATE DATABASE.
    admin_dsn = urlunparse(
        parsed._replace(scheme="postgresql", path="/postgres")
    )
    conn = await asyncpg.connect(dsn=admin_dsn)
    try:
        exists = await conn.fetchval(
            "SELECT 1 FROM pg_database WHERE datname = $1", test_db_name
        )
        if not exists:
            await conn.execute(f'CREATE DATABASE "{test_db_name}"')
    finally:
        await conn.close()


@pytest.fixture(scope="session")
def _setup_schema() -> None:
    async def _create() -> None:
        await _ensure_test_database_exists()
        engine = create_async_engine(TEST_DATABASE_URL, poolclass=NullPool)
        async with engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)
        await engine.dispose()

    async def _drop() -> None:
        engine = create_async_engine(TEST_DATABASE_URL, poolclass=NullPool)
        async with engine.begin() as conn:
            await conn.run_sync(Base.metadata.drop_all)
        await engine.dispose()

    asyncio.run(_create())
    yield
    asyncio.run(_drop())


@pytest_asyncio.fixture
async def client(_setup_schema: None) -> AsyncGenerator[AsyncClient, None]:
    engine = create_async_engine(TEST_DATABASE_URL, poolclass=NullPool)
    factory = async_sessionmaker(engine, expire_on_commit=False)

    async def _override() -> AsyncGenerator[AsyncSession, None]:
        async with factory() as session:
            yield session

    app.dependency_overrides[get_session] = _override

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as c:
        yield c

    app.dependency_overrides.clear()

    async with engine.begin() as conn:
        for table in reversed(Base.metadata.sorted_tables):
            await conn.execute(table.delete())
    await engine.dispose()
