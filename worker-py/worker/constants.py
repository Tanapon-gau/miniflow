STATUS_RUNNING = "running"
STATUS_SUCCESS = "success"
STATUS_FAILED = "failed"

TASK_TYPE_PYTHON = "python"
TASK_TYPE_ML = "ml"

QUEUE_PYTHON = "tasks:python"
QUEUE_ML = "tasks:ml"

EVENTS_STREAM = "events"

EVENT_STARTED = "started"
EVENT_LOG = "log"
EVENT_SUCCEEDED = "succeeded"
EVENT_FAILED = "failed"

# seconds to block on BRPOP before looping to check for shutdown
BRPOP_TIMEOUT = 5
