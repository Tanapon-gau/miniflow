import json
from dataclasses import dataclass
from uuid import UUID


@dataclass
class TaskMessage:
    task_id: UUID
    run_id: UUID
    type: str
    payload: dict  # type: ignore[type-arg]
    timeout_seconds: int
    max_retries: int = 0
    attempt: int = 1

    @classmethod
    def from_bytes(cls, data: bytes) -> "TaskMessage":
        raw = json.loads(data)
        return cls(
            task_id=UUID(raw["task_id"]),
            run_id=UUID(raw["run_id"]),
            type=raw["type"],
            payload=raw["payload"],
            timeout_seconds=int(raw["timeout_seconds"]),
            max_retries=int(raw.get("max_retries", 0)),
            attempt=int(raw.get("attempt", 1)),
        )
