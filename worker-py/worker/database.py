from uuid import UUID

import psycopg

from worker import constants


class WorkerDatabase:
    def __init__(self, dsn: str) -> None:
        self._connection = psycopg.connect(dsn)

    def close(self) -> None:
        self._connection.close()

    def mark_task_running(self, task_id: UUID) -> None:
        with self._connection.cursor() as cursor:
            cursor.execute(
                "UPDATE tasks SET status = %s, started_at = NOW() WHERE id = %s",
                (constants.STATUS_RUNNING, str(task_id)),
            )
        self._connection.commit()

    def mark_task_done(self, task_id: UUID, status: str) -> None:
        with self._connection.cursor() as cursor:
            cursor.execute(
                "UPDATE tasks SET status = %s, finished_at = NOW() WHERE id = %s",
                (status, str(task_id)),
            )
        self._connection.commit()
