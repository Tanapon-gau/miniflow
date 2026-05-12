import json
import logging
import threading
import time
from threading import Event

import redis

from worker import constants
from worker.database import WorkerDatabase
from worker.events import EventPublisher
from worker.executor import run_python_task
from worker.models import TaskMessage

logger = logging.getLogger(__name__)


def _retry_delay(attempt: int) -> float:
    """Return backoff seconds for the given failed attempt (1-based).

    Formula: base * 2^(attempt-1), capped at RETRY_MAX_DELAY_SECONDS.
    """
    shift = max(0, attempt - 1)
    delay = constants.RETRY_BASE_DELAY_SECONDS * (2**shift)
    return float(min(delay, constants.RETRY_MAX_DELAY_SECONDS))


def consume_forever(
    redis_client: redis.Redis,  # type: ignore[type-arg]
    database: WorkerDatabase,
    shutdown_event: Event,
) -> None:
    publisher = EventPublisher(redis_client)
    queues = [constants.QUEUE_PYTHON, constants.QUEUE_ML]

    while not shutdown_event.is_set():
        result = redis_client.brpop(queues, timeout=constants.BRPOP_TIMEOUT)
        if result is None:
            continue

        _queue_name, raw_data = result
        try:
            message = TaskMessage.from_bytes(raw_data)
        except Exception as exc:
            logger.error("failed to parse task message: %s — raw: %.200s", exc, raw_data)
            continue

        _handle(message, database, publisher, redis_client)


def _handle(
    message: TaskMessage,
    database: WorkerDatabase,
    publisher: EventPublisher,
    redis_client: redis.Redis,  # type: ignore[type-arg]
) -> None:
    try:
        database.mark_task_running(message.task_id, message.attempt)
    except Exception as exc:
        logger.error("mark task %s running failed: %s", message.task_id, exc)
        return

    publisher.publish(message.task_id, message.run_id, constants.EVENT_STARTED)
    logger.info(
        "task %s (type=%s, attempt=%d) started",
        message.task_id,
        message.type,
        message.attempt,
    )

    result = run_python_task(message)

    publisher.publish(
        message.task_id,
        message.run_id,
        constants.EVENT_LOG,
        {"output": result.output},
    )

    if result.error is not None:
        logger.error("task %s attempt %d failed: %s", message.task_id, message.attempt, result.error)
        if message.attempt <= message.max_retries:
            _schedule_retry(message, database, publisher, redis_client)
            return

        try:
            database.mark_task_done(message.task_id, constants.STATUS_FAILED)
        except Exception as exc:
            logger.error("mark task %s done failed: %s", message.task_id, exc)
        publisher.publish(
            message.task_id,
            message.run_id,
            constants.EVENT_FAILED,
            {"status": constants.STATUS_FAILED},
        )
        return

    try:
        database.mark_task_done(message.task_id, constants.STATUS_SUCCESS)
    except Exception as exc:
        logger.error("mark task %s done failed: %s", message.task_id, exc)
    publisher.publish(
        message.task_id,
        message.run_id,
        constants.EVENT_SUCCEEDED,
        {"status": constants.STATUS_SUCCESS},
    )
    logger.info("task %s (type=%s) succeeded", message.task_id, message.type)


def _schedule_retry(
    message: TaskMessage,
    database: WorkerDatabase,
    publisher: EventPublisher,
    redis_client: redis.Redis,  # type: ignore[type-arg]
) -> None:
    next_attempt = message.attempt + 1
    delay = _retry_delay(message.attempt)

    try:
        database.mark_task_retrying(message.task_id, message.attempt)
    except Exception as exc:
        logger.error("mark task %s retrying failed: %s", message.task_id, exc)

    publisher.publish(
        message.task_id,
        message.run_id,
        constants.EVENT_RETRYING,
        {"attempt": message.attempt, "next_in": f"{delay:.0f}s"},
    )
    logger.info(
        "task %s: attempt %d/%d failed, retrying in %.0fs",
        message.task_id,
        message.attempt,
        message.max_retries + 1,
        delay,
    )

    retry_message = TaskMessage(
        task_id=message.task_id,
        run_id=message.run_id,
        type=message.type,
        payload=message.payload,
        timeout_seconds=message.timeout_seconds,
        max_retries=message.max_retries,
        attempt=next_attempt,
    )
    raw = json.dumps(
        {
            "task_id": str(retry_message.task_id),
            "run_id": str(retry_message.run_id),
            "type": retry_message.type,
            "payload": retry_message.payload,
            "timeout_seconds": retry_message.timeout_seconds,
            "max_retries": retry_message.max_retries,
            "attempt": retry_message.attempt,
        }
    ).encode()

    queue_name = constants.QUEUE_PYTHON if message.type == constants.TASK_TYPE_PYTHON else constants.QUEUE_ML

    # sleep and re-queue in a background thread so the consumer loop stays unblocked
    def _push() -> None:
        time.sleep(delay)
        redis_client.lpush(queue_name, raw)

    thread = threading.Thread(target=_push, daemon=True)
    thread.start()
