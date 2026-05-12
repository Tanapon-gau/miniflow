import uuid
from datetime import datetime

from pydantic import BaseModel, computed_field, field_validator


class WorkflowCreate(BaseModel):
    name: str
    description: str | None = None
    dag: dict
    version: int = 1

    @field_validator("name")
    @classmethod
    def name_not_empty(cls, value: str) -> str:
        if not value.strip():
            raise ValueError(f"name must not be empty, got: {value!r}")
        return value

    @field_validator("dag")
    @classmethod
    def dag_has_tasks(cls, value: dict) -> dict:
        if "tasks" not in value:
            raise ValueError(
                f"dag is missing required 'tasks' key, got keys: {sorted(value.keys())}"
            )
        return value


class WorkflowUpdate(BaseModel):
    name: str | None = None
    description: str | None = None
    dag: dict | None = None

    @field_validator("name")
    @classmethod
    def name_not_empty(cls, value: str | None) -> str | None:
        if value is not None and not value.strip():
            raise ValueError(f"name must not be empty, got: {value!r}")
        return value

    @field_validator("dag")
    @classmethod
    def dag_has_tasks(cls, value: dict | None) -> dict | None:
        if value is not None and "tasks" not in value:
            raise ValueError(
                f"dag is missing required 'tasks' key, got keys: {sorted(value.keys())}"
            )
        return value


class WorkflowRead(BaseModel):
    id: uuid.UUID
    name: str
    description: str | None
    dag: dict
    version: int
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class TaskRead(BaseModel):
    id: uuid.UUID
    run_id: uuid.UUID
    name: str
    type: str
    status: str
    payload: dict | None
    attempt: int
    max_retries: int
    timeout_seconds: int
    started_at: datetime | None
    finished_at: datetime | None
    created_at: datetime

    model_config = {"from_attributes": True}


class RunRead(BaseModel):
    id: uuid.UUID
    workflow_id: uuid.UUID
    status: str
    triggered_at: datetime
    started_at: datetime | None
    finished_at: datetime | None

    model_config = {"from_attributes": True}


class RunDetail(RunRead):
    tasks: list[TaskRead]


class TaskTimeline(BaseModel):
    task_id: uuid.UUID
    name: str
    status: str
    queued_at: datetime
    started_at: datetime | None
    finished_at: datetime | None

    @computed_field
    @property
    def duration_seconds(self) -> float | None:
        if self.started_at and self.finished_at:
            return (self.finished_at - self.started_at).total_seconds()
        return None


class RunTimeline(BaseModel):
    run_id: uuid.UUID
    status: str
    triggered_at: datetime
    finished_at: datetime | None
    tasks: list[TaskTimeline]
