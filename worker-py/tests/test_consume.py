import json
from threading import Event
from unittest.mock import MagicMock, call, patch
from uuid import uuid4

import pytest

from worker import constants
from worker.consume import _handle, consume_forever
from worker.models import TaskMessage


def make_message(script: str = "print('ok')", timeout: int = 10) -> TaskMessage:
    return TaskMessage(
        task_id=uuid4(),
        run_id=uuid4(),
        type=constants.TASK_TYPE_PYTHON,
        payload={"script": script},
        timeout_seconds=timeout,
    )


def make_raw(message: TaskMessage) -> bytes:
    return json.dumps(
        {
            "task_id": str(message.task_id),
            "run_id": str(message.run_id),
            "type": message.type,
            "payload": message.payload,
            "timeout_seconds": message.timeout_seconds,
        }
    ).encode()


class TestHandle:
    def setup_method(self) -> None:
        self.database = MagicMock()
        self.publisher = MagicMock()

    def test_success_path(self) -> None:
        message = make_message("print('hello')")
        _handle(message, self.database, self.publisher)

        self.database.mark_task_running.assert_called_once_with(message.task_id)
        self.database.mark_task_done.assert_called_once_with(
            message.task_id, constants.STATUS_SUCCESS
        )
        event_types = [c.args[2] for c in self.publisher.publish.call_args_list]
        assert constants.EVENT_STARTED in event_types
        assert constants.EVENT_LOG in event_types
        assert constants.EVENT_SUCCEEDED in event_types

    def test_failed_path(self) -> None:
        message = make_message("raise SystemExit(1)")
        _handle(message, self.database, self.publisher)

        self.database.mark_task_done.assert_called_once_with(
            message.task_id, constants.STATUS_FAILED
        )
        event_types = [c.args[2] for c in self.publisher.publish.call_args_list]
        assert constants.EVENT_FAILED in event_types

    def test_mark_running_failure_skips_execution(self) -> None:
        self.database.mark_task_running.side_effect = RuntimeError("db down")
        message = make_message()
        _handle(message, self.database, self.publisher)

        self.database.mark_task_done.assert_not_called()
        self.publisher.publish.assert_not_called()


class TestConsumeForever:
    def test_stops_on_shutdown_event(self) -> None:
        shutdown_event = Event()
        redis_client = MagicMock()
        database = MagicMock()

        # brpop returns None (timeout) then shutdown is set
        call_count = 0

        def brpop_side_effect(queues: list[str], timeout: int) -> None:
            nonlocal call_count
            call_count += 1
            if call_count >= 2:
                shutdown_event.set()
            return None

        redis_client.brpop.side_effect = brpop_side_effect
        consume_forever(redis_client, database, shutdown_event)
        assert call_count >= 2

    def test_dispatches_valid_message(self) -> None:
        shutdown_event = Event()
        redis_client = MagicMock()
        database = MagicMock()

        message = make_message()
        raw = make_raw(message)
        call_count = 0

        def brpop_side_effect(queues: list[str], timeout: int) -> tuple[bytes, bytes] | None:
            nonlocal call_count
            call_count += 1
            if call_count == 1:
                return (b"tasks:python", raw)
            shutdown_event.set()
            return None

        redis_client.brpop.side_effect = brpop_side_effect
        redis_client.xadd = MagicMock()

        consume_forever(redis_client, database, shutdown_event)
        database.mark_task_running.assert_called_once_with(message.task_id)
