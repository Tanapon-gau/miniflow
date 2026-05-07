STATUS_RUNNING = "running"
STATUS_SUCCESS = "success"
STATUS_FAILED = "failed"
STATUS_RETRYING = "retrying"

TASK_TYPE_PYTHON = "python"
TASK_TYPE_ML = "ml"

QUEUE_PYTHON = "tasks:python"
QUEUE_ML = "tasks:ml"

EVENTS_STREAM = "events"

EVENT_STARTED = "started"
EVENT_LOG = "log"
EVENT_SUCCEEDED = "succeeded"
EVENT_FAILED = "failed"
EVENT_RETRYING = "retrying"

RETRY_BASE_DELAY_SECONDS = 1
RETRY_MAX_DELAY_SECONDS = 30

# seconds to block on BRPOP before looping to check for shutdown
BRPOP_TIMEOUT = 5
