import subprocess
import sys


def _alembic(args: list[str]) -> None:
    result = subprocess.run(
        [sys.executable, "-m", "alembic", *args],
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, f"alembic {' '.join(args)} failed:\n{result.stderr}"


def test_upgrade_and_downgrade() -> None:
    _alembic(["upgrade", "head"])
    _alembic(["downgrade", "base"])
