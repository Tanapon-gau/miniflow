import subprocess
import textwrap
from dataclasses import dataclass

from worker.models import TaskMessage


@dataclass
class Result:
    output: str
    error: Exception | None = None


def run_python_task(message: TaskMessage) -> Result:
    payload = message.payload
    script = payload.get("script", "")
    if not script:
        return Result(
            output="",
            error=ValueError(
                f"task {message.task_id}: python payload missing required field 'script'"
            ),
        )

    try:
        completed = subprocess.run(
            ["python3", "-c", textwrap.dedent(script)],
            capture_output=True,
            text=True,
            timeout=message.timeout_seconds,
        )
    except subprocess.TimeoutExpired as exc:
        return Result(
            output="",
            error=TimeoutError(
                f"task {message.task_id}: script timed out after {message.timeout_seconds}s"
            ),
        )

    output = completed.stdout + completed.stderr
    if completed.returncode != 0:
        return Result(
            output=output,
            error=RuntimeError(
                f"task {message.task_id}: script exited with code {completed.returncode}"
            ),
        )
    return Result(output=output)
