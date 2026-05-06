import json
from datetime import datetime, timezone
from typing import Any
from uuid import UUID

import redis

from worker import constants


class EventPublisher:
    def __init__(self, client: redis.Redis) -> None:  # type: ignore[type-arg]
        self._client = client

    def publish(
        self,
        task_id: UUID,
        run_id: UUID,
        event_type: str,
        data: dict[str, Any] | None = None,
    ) -> None:
        event = {
            "task_id": str(task_id),
            "run_id": str(run_id),
            "event_type": event_type,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "data": data,
        }
        self._client.xadd(
            constants.EVENTS_STREAM,
            {"data": json.dumps(event)},
        )
