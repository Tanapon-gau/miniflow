import json
from uuid import uuid4

import pytest

from worker.executor import run_python_task
from worker.models import TaskMessage


def python_message(script: str, timeout: int = 10) -> TaskMessage:
    return TaskMessage(
        task_id=uuid4(),
        run_id=uuid4(),
        type="python",
        payload={"script": script},
        timeout_seconds=timeout,
    )


def test_run_python_task_success() -> None:
    result = run_python_task(python_message("print('hello')"))
    assert result.error is None
    assert "hello" in result.output


def test_run_python_task_nonzero_exit() -> None:
    result = run_python_task(python_message("raise SystemExit(1)"))
    assert result.error is not None


def test_run_python_task_runtime_error() -> None:
    result = run_python_task(python_message("1/0"))
    assert result.error is not None
    assert "ZeroDivisionError" in result.output


def test_run_python_task_timeout() -> None:
    result = run_python_task(python_message("import time; time.sleep(60)", timeout=1))
    assert result.error is not None
    assert isinstance(result.error, TimeoutError)


def test_run_python_task_missing_script() -> None:
    message = TaskMessage(
        task_id=uuid4(),
        run_id=uuid4(),
        type="python",
        payload={},
        timeout_seconds=5,
    )
    result = run_python_task(message)
    assert result.error is not None
    assert "script" in str(result.error)


def test_run_python_task_captures_stderr() -> None:
    result = run_python_task(python_message("import sys; sys.stderr.write('err\\n')"))
    assert result.error is None
    assert "err" in result.output
