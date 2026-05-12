import os
import subprocess
import sys

from tests.conftest import TEST_DATABASE_URL


def _alembic(args: list[str]) -> None:
    env = {**os.environ, "DATABASE_URL": TEST_DATABASE_URL}
    result = subprocess.run(
        [sys.executable, "-m", "alembic", *args],
        capture_output=True,
        text=True,
        env=env,
    )
    assert result.returncode == 0, f"alembic {' '.join(args)} failed:\n{result.stderr}"


def test_upgrade_and_downgrade() -> None:
    _alembic(["upgrade", "head"])
    _alembic(["downgrade", "base"])
