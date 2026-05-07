import json
from threading import Event
from unittest.mock import MagicMock, patch
from uuid import uuid4

from worker import constants
from worker.consume import _handle, _retry_delay, consume_forever
from worker.models import TaskMessage


def make_message(
    script: str = "print('ok')",
    timeout: int = 10,
    max_retries: int = 0,
    attempt: int = 1,
) -> TaskMessage:
    return TaskMessage(
        task_id=uuid4(),
        run_id=uuid4(),
        type=constants.TASK_TYPE_PYTHON,
        payload={"script": script},
        timeout_seconds=timeout,
        max_retries=max_retries,
        attempt=attempt,
    )


def make_raw(message: TaskMessage) -> bytes:
    return json.dumps(
        {
            "task_id": str(message.task_id),
            "run_id": str(message.run_id),
            "type": message.type,
            "payload": message.payload,
            "timeout_seconds": message.timeout_seconds,
            "max_retries": message.max_retries,
            "attempt": message.attempt,
        }
    ).encode()


class TestRetryDelay:
    def test_attempt_one_returns_base_delay(self) -> None:
        assert _retry_delay(1) == constants.RETRY_BASE_DELAY_SECONDS

    def test_doubles_each_attempt(self) -> None:
        assert _retry_delay(2) == _retry_delay(1) * 2
        assert _retry_delay(3) == _retry_delay(2) * 2
        assert _retry_delay(4) == _retry_delay(3) * 2

    def test_capped_at_max(self) -> None:
        for attempt in range(10, 20):
            assert _retry_delay(attempt) == constants.RETRY_MAX_DELAY_SECONDS


class TestHandle:
    def setup_method(self) -> None:
        self.database = MagicMock()
        self.publisher = MagicMock()
        self.redis_client = MagicMock()
        self.redis_client.xadd = MagicMock()

    def test_success_path(self) -> None:
        message = make_message("print('hello')")
        _handle(message, self.database, self.publisher, self.redis_client)

        self.database.mark_task_running.assert_called_once_with(message.task_id, message.attempt)
        self.database.mark_task_done.assert_called_once_with(
            message.task_id, constants.STATUS_SUCCESS
        )
        event_types = [c.args[2] for c in self.publisher.publish.call_args_list]
        assert constants.EVENT_SUCCEEDED in event_types

    def test_failed_path_no_retries(self) -> None:
        message = make_message("raise SystemExit(1)", max_retries=0, attempt=1)
        _handle(message, self.database, self.publisher, self.redis_client)

        self.database.mark_task_done.assert_called_once_with(
            message.task_id, constants.STATUS_FAILED
        )
        event_types = [c.args[2] for c in self.publisher.publish.call_args_list]
        assert constants.EVENT_FAILED in event_types
        assert constants.EVENT_RETRYING not in event_types

    def test_retry_path_when_attempts_remain(self) -> None:
        message = make_message("raise SystemExit(1)", max_retries=2, attempt=1)
        with patch("worker.consume.threading.Thread") as mock_thread_cls:
            mock_thread = MagicMock()
            mock_thread_cls.return_value = mock_thread
            _handle(message, self.database, self.publisher, self.redis_client)

        self.database.mark_task_retrying.assert_called_once_with(message.task_id, 1)
        event_types = [c.args[2] for c in self.publisher.publish.call_args_list]
        assert constants.EVENT_RETRYING in event_types
        assert constants.EVENT_FAILED not in event_types
        mock_thread.start.assert_called_once()

    def test_no_retry_when_attempts_exhausted(self) -> None:
        # attempt=3 > max_retries=2 → should fail, not retry
        message = make_message("raise SystemExit(1)", max_retries=2, attempt=3)
        _handle(message, self.database, self.publisher, self.redis_client)

        self.database.mark_task_done.assert_called_once_with(
            message.task_id, constants.STATUS_FAILED
        )
        self.database.mark_task_retrying.assert_not_called()

    def test_mark_running_failure_skips_execution(self) -> None:
        self.database.mark_task_running.side_effect = RuntimeError("db down")
        message = make_message()
        _handle(message, self.database, self.publisher, self.redis_client)

        self.database.mark_task_done.assert_not_called()
        self.publisher.publish.assert_not_called()


class TestConsumeForever:
    def test_stops_on_shutdown_event(self) -> None:
        shutdown_event = Event()
        redis_client = MagicMock()
        database = MagicMock()

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
        database.mark_task_running.assert_called_once_with(message.task_id, message.attempt)
