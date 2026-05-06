import logging
from threading import Event

import redis

from worker import constants
from worker.database import WorkerDatabase
from worker.events import EventPublisher
from worker.executor import run_python_task
from worker.models import TaskMessage

logger = logging.getLogger(__name__)


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

        _handle(message, database, publisher)


def _handle(
    message: TaskMessage,
    database: WorkerDatabase,
    publisher: EventPublisher,
) -> None:
    try:
        database.mark_task_running(message.task_id)
    except Exception as exc:
        logger.error("mark task %s running failed: %s", message.task_id, exc)
        return

    publisher.publish(message.task_id, message.run_id, constants.EVENT_STARTED)

    result = run_python_task(message)

    publisher.publish(
        message.task_id,
        message.run_id,
        constants.EVENT_LOG,
        {"output": result.output},
    )

    if result.error is not None:
        logger.error("task %s failed: %s", message.task_id, result.error)
        task_status = constants.STATUS_FAILED
        completion_event = constants.EVENT_FAILED
    else:
        task_status = constants.STATUS_SUCCESS
        completion_event = constants.EVENT_SUCCEEDED

    try:
        database.mark_task_done(message.task_id, task_status)
    except Exception as exc:
        logger.error("mark task %s done failed: %s", message.task_id, exc)

    publisher.publish(
        message.task_id,
        message.run_id,
        completion_event,
        {"status": task_status},
    )
