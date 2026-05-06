import logging
import os
import signal
from threading import Event

import redis

from worker.consume import consume_forever
from worker.database import WorkerDatabase

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)


def main() -> None:
    redis_url = os.environ["REDIS_URL"]
    database_url = os.environ["DATABASE_URL"]

    redis_client = redis.from_url(redis_url)
    database = WorkerDatabase(database_url)
    shutdown_event = Event()

    def handle_shutdown(signum: int, frame: object) -> None:
        logger.info("received signal %d, shutting down", signum)
        shutdown_event.set()

    signal.signal(signal.SIGTERM, handle_shutdown)
    signal.signal(signal.SIGINT, handle_shutdown)

    logger.info("worker-py started")
    try:
        consume_forever(redis_client, database, shutdown_event)
    finally:
        database.close()
        redis_client.close()
        logger.info("worker-py stopped")


if __name__ == "__main__":
    main()
