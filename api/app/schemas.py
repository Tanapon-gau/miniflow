import uuid
from datetime import datetime

from pydantic import BaseModel, field_validator


class WorkflowCreate(BaseModel):
    name: str
    description: str | None = None
    dag: dict
    version: int = 1

    @field_validator("name")
    @classmethod
    def name_not_empty(cls, v: str) -> str:
        if not v.strip():
            raise ValueError("name must not be empty")
        return v

    @field_validator("dag")
    @classmethod
    def dag_has_tasks(cls, v: dict) -> dict:
        if "tasks" not in v:
            raise ValueError("dag must contain a 'tasks' key")
        return v


class WorkflowUpdate(BaseModel):
    name: str | None = None
    description: str | None = None
    dag: dict | None = None

    @field_validator("name")
    @classmethod
    def name_not_empty(cls, v: str | None) -> str | None:
        if v is not None and not v.strip():
            raise ValueError("name must not be empty")
        return v

    @field_validator("dag")
    @classmethod
    def dag_has_tasks(cls, v: dict | None) -> dict | None:
        if v is not None and "tasks" not in v:
            raise ValueError("dag must contain a 'tasks' key")
        return v


class WorkflowRead(BaseModel):
    id: uuid.UUID
    name: str
    description: str | None
    dag: dict
    version: int
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}
